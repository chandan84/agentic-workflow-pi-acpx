// Command agents is the entry point for the agents bounded-context service.
package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/authn"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/config"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/health"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/obs"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/pgxconn"
	agentsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/agents/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/app"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/grpcsrv"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/ports"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/store"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/agents/internal/stub"

	"google.golang.org/grpc"
)

// Config is the agents-service typed config.
type Config struct {
	config.Base `yaml:",inline"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "agents:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := Config{Base: config.Defaults("agents")}
	if err := config.Load("agents", &cfg); err != nil {
		return err
	}
	if err := cfg.Base.Validate(); err != nil {
		return err
	}

	log := obs.Logger("agents", cfg.LogLevel, cfg.LogFormat)
	log.Info("starting", "grpc", cfg.GRPCAddr, "http", cfg.HTTPAddr, "env", cfg.Env)

	otelShutdown, err := obs.InitTracing(ctx, "agents", cfg.OTel.Endpoint)
	if err != nil {
		return err
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	// Bus.
	var bus events.Bus
	if cfg.Stubs["bus"] == "stub" {
		bus = events.NewInMemoryBus()
		log.Info("bus: in-memory stub")
	} else {
		nb, err := events.DialNATS(cfg.NATS.URL)
		if err != nil {
			log.Warn("nats dial failed; falling back to stub", "err", err)
			bus = events.NewInMemoryBus()
		} else {
			bus = nb
			log.Info("bus: nats", "url", cfg.NATS.URL)
		}
	}
	defer func() { _ = bus.Close() }()

	// Store.
	var st ports.Store = stub.New()
	if cfg.Stubs["store"] != "stub" {
		db, err := pgxconn.Open(ctx, pgxconn.Options{DSN: cfg.Postgres.DSN, Schema: cfg.Postgres.Schema})
		if err != nil {
			log.Warn("postgres unavailable; using stub store", "err", err)
		} else {
			st = store.New(db)
			defer func() { _ = db.Close() }()
			log.Info("store: postgres", "schema", cfg.Postgres.Schema)
		}
	} else {
		log.Info("store: stub (configured)")
	}

	// AuthN (interface in scope; gRPC interceptor below applies it).
	au := authn.From(cfg.Auth.Mode, cfg.Auth.Token)
	log.Info("authn", "mode", cfg.Auth.Mode)

	// gRPC server.
	gs := grpc.NewServer(grpc.UnaryInterceptor(authUnary(au)))
	agentsv1.RegisterAgentsServiceServer(gs, grpcsrv.New(app.New(st, bus)))

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	go func() {
		log.Info("grpc listening", "addr", lis.Addr().String())
		if err := gs.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Error("grpc serve", "err", err)
		}
	}()

	// Health HTTP.
	h := health.New()
	h.Register("store", st.Ping)
	hs := &http.Server{Addr: cfg.HTTPAddr, Handler: h.Mux(), ReadHeaderTimeout: 3 * time.Second}
	go func() {
		log.Info("http listening", "addr", cfg.HTTPAddr)
		if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http serve", "err", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = hs.Shutdown(shutdownCtx)
	gs.GracefulStop()
	log.Info("bye")
	return nil
}

// authUnary applies the configured Authenticator. With Auth.Mode=none this is
// a passthrough; with static-token it requires a "authorization" metadata key.
func authUnary(a authn.Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// In none mode, skip cheaply.
		if _, ok := a.(authn.None); ok {
			return handler(ctx, req)
		}
		// metadata extraction omitted in skeleton — real wiring uses grpc/metadata.
		_, err := a.Authenticate(ctx, "")
		if err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}
