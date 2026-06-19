// Command gateway is the single gRPC aggregator exposed to the desktop app.
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
	agentsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/agents/v1"
	auditv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/audit/v1"
	executionv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/execution/v1"
	flowsv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/flows/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/gateway/internal/audit"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/gateway/internal/grpcsrv"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/gateway/internal/upstreams"

	"google.golang.org/grpc"
)

// Config is the gateway typed config.
type Config struct {
	config.Base `yaml:",inline"`
	Upstreams   struct {
		Agents       string `yaml:"agents"`
		Flow         string `yaml:"flow"`
		Orchestrator string `yaml:"orchestrator"`
	} `yaml:"upstreams"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "gateway:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := Config{Base: config.Defaults("gateway")}
	if err := config.Load("gateway", &cfg); err != nil {
		return err
	}
	if err := cfg.Base.Validate(); err != nil {
		return err
	}

	log := obs.Logger("gateway", cfg.LogLevel, cfg.LogFormat)
	log.Info("starting", "grpc", cfg.GRPCAddr, "http", cfg.HTTPAddr, "env", cfg.Env)
	otelShutdown, _ := obs.InitTracing(ctx, "gateway", cfg.OTel.Endpoint)
	defer func() { _ = otelShutdown(context.Background()) }()

	var bus events.Bus
	if nb, err := events.DialNATS(cfg.NATS.URL); err == nil {
		bus = nb
	} else {
		log.Warn("nats unavailable; in-memory bus", "err", err)
		bus = events.NewInMemoryBus()
	}
	defer func() { _ = bus.Close() }()

	// Audit source: in-memory, fed by NATS subscriptions to runs.* and audit.event.
	src := audit.New()
	go subscribeAuditTopics(ctx, bus, src, log)

	gs := grpc.NewServer()
	auditv1.RegisterAuditServiceServer(gs, grpcsrv.NewAudit(src))

	// Proxies to upstream services. Dials are non-blocking — failures appear
	// at request time rather than at boot so the gateway is always live.
	if addr := cfg.Upstreams.Agents; addr != "" {
		if cc, err := upstreams.Dial(addr); err == nil {
			agentsv1.RegisterAgentsServiceServer(gs, upstreams.NewAgentsProxy(cc))
			defer func() { _ = cc.Close() }()
			log.Info("proxy: agents", "addr", addr)
		} else {
			log.Warn("agents dial failed", "addr", addr, "err", err)
		}
	}
	if addr := cfg.Upstreams.Flow; addr != "" {
		if cc, err := upstreams.Dial(addr); err == nil {
			flowsv1.RegisterFlowServiceServer(gs, upstreams.NewFlowsProxy(cc))
			defer func() { _ = cc.Close() }()
			log.Info("proxy: flow", "addr", addr)
		} else {
			log.Warn("flow dial failed", "addr", addr, "err", err)
		}
	}
	if addr := cfg.Upstreams.Orchestrator; addr != "" {
		if cc, err := upstreams.Dial(addr); err == nil {
			executionv1.RegisterExecutionServiceServer(gs, upstreams.NewExecutionProxy(cc))
			defer func() { _ = cc.Close() }()
			log.Info("proxy: orchestrator", "addr", addr)
		} else {
			log.Warn("orchestrator dial failed", "addr", addr, "err", err)
		}
	}

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

	h := health.New()
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

func subscribeAuditTopics(ctx context.Context, bus events.Bus, sink *audit.Memory, log interface{ Info(string, ...any) }) {
	subjects := []string{
		events.SubjectRunStarted, events.SubjectRunUpdated, events.SubjectRunCompleted, events.SubjectRunFailed,
		events.SubjectRunCheckpointWait, events.SubjectRunCheckpointDone,
		events.SubjectRunSegmentStart, events.SubjectRunSegmentEnd, events.SubjectRunWorkItemEmitted,
		events.SubjectAuditEvent,
	}
	for _, subj := range subjects {
		subject := subj
		_, _ = bus.Subscribe(ctx, subject, func(_ string, data []byte) {
			sink.Add(audit.Event{
				Layer: "app", Kind: subject, PayloadJSON: string(data), At: time.Now(),
			})
		})
	}
	log.Info("subscribed to audit topics", "count", len(subjects))
}
