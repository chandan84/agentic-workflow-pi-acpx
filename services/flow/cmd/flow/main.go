// Command flow is the entry point for the flow bounded-context service.
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

	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/config"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/events"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/health"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/obs"
	flowsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/flows/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/app"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/author"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/grpcsrv"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/ports"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/flow/internal/stub"

	"google.golang.org/grpc"
)

// Config carries the typed config for flow-service.
type Config struct {
	config.Base `yaml:",inline"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "flow:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := Config{Base: config.Defaults("flow")}
	if err := config.Load("flow", &cfg); err != nil {
		return err
	}
	if err := cfg.Base.Validate(); err != nil {
		return err
	}

	log := obs.Logger("flow", cfg.LogLevel, cfg.LogFormat)
	log.Info("starting", "grpc", cfg.GRPCAddr, "http", cfg.HTTPAddr, "env", cfg.Env)

	otelShutdown, _ := obs.InitTracing(ctx, "flow", cfg.OTel.Endpoint)
	defer func() { _ = otelShutdown(context.Background()) }()

	// Bus.
	var bus events.Bus
	if cfg.Stubs["bus"] == "stub" {
		bus = events.NewInMemoryBus()
	} else {
		if nb, err := events.DialNATS(cfg.NATS.URL); err == nil {
			bus = nb
		} else {
			log.Warn("nats dial failed; in-memory bus", "err", err)
			bus = events.NewInMemoryBus()
		}
	}
	defer func() { _ = bus.Close() }()

	// Runtime (acpx/pi).
	var rt runtime.AgentRuntime
	if cfg.Stubs["runtime"] == "stub" {
		rt = runtime.NewStub()
		log.Info("runtime: stub")
	} else {
		rt = runtime.NewExec("acpx", "pi", "")
		log.Info("runtime: exec (acpx/pi)")
	}

	// Store.
	var st ports.Store = stub.New()
	log.Info("store: stub (postgres adapter wires on driver registration)")

	// App.
	svc := app.New(st, bus, author.New(rt))

	// gRPC.
	gs := grpc.NewServer()
	flowsv1.RegisterFlowServiceServer(gs, grpcsrv.New(svc))
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return err
	}
	go func() {
		log.Info("grpc listening", "addr", lis.Addr().String())
		if err := gs.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Error("grpc", "err", err)
		}
	}()

	// Health HTTP.
	h := health.New()
	h.Register("store", st.Ping)
	hs := &http.Server{Addr: cfg.HTTPAddr, Handler: h.Mux(), ReadHeaderTimeout: 3 * time.Second}
	go func() {
		log.Info("http listening", "addr", cfg.HTTPAddr)
		if err := hs.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http", "err", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutdown")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = hs.Shutdown(shutdownCtx)
	gs.GracefulStop()
	return nil
}
