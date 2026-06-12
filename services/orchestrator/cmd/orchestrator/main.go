// Command orchestrator is the entry point for the orchestrator service.
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
	executionv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/execution/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/app"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/grpcsrv"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/ports"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/stub"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/temporalio"
	wfpkg "github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/workflow"

	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
)

// Config is the orchestrator typed config.
type Config struct {
	config.Base `yaml:",inline"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "orchestrator:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := Config{Base: config.Defaults("orchestrator")}
	if err := config.Load("orchestrator", &cfg); err != nil {
		return err
	}
	if err := cfg.Base.Validate(); err != nil {
		return err
	}

	log := obs.Logger("orchestrator", cfg.LogLevel, cfg.LogFormat)
	log.Info("starting", "grpc", cfg.GRPCAddr, "http", cfg.HTTPAddr, "env", cfg.Env)
	otelShutdown, _ := obs.InitTracing(ctx, "orchestrator", cfg.OTel.Endpoint)
	defer func() { _ = otelShutdown(context.Background()) }()

	var bus events.Bus
	if cfg.Stubs["bus"] == "stub" {
		bus = events.NewInMemoryBus()
	} else {
		if nb, err := events.DialNATS(cfg.NATS.URL); err == nil {
			bus = nb
		} else {
			log.Warn("nats unavailable; in-memory bus", "err", err)
			bus = events.NewInMemoryBus()
		}
	}
	defer func() { _ = bus.Close() }()

	var rt runtime.AgentRuntime
	if cfg.Stubs["runtime"] == "stub" {
		rt = runtime.NewStub()
	} else {
		rt = runtime.NewExec("acpx", "pi", "")
	}

	var st ports.Store = stub.New()

	var starter app.WorkflowStarter
	var signaler app.WorkflowSignaler
	if cfg.Stubs["temporal"] == "stub" {
		s := temporalio.NewStub()
		starter, signaler = s, s
		log.Info("temporal: stub")
	} else {
		t, err := temporalio.New(cfg.Temporal.HostPort, cfg.Temporal.Namespace, cfg.Temporal.TaskQueue)
		if err != nil {
			log.Warn("temporal dial failed; falling back to stub", "err", err)
			s := temporalio.NewStub()
			starter, signaler = s, s
		} else {
			starter, signaler = t, t
			defer t.Close()
			// Worker.
			w := worker.New(t.C, cfg.Temporal.TaskQueue, worker.Options{})
			w.RegisterWorkflow(wfpkg.RunFlow)
			activities := &wfpkg.Activities{Runtime: rt}
			w.RegisterActivity(activities.ExecuteSegment)
			w.RegisterActivity(activities.RecordCheckpointAwaiting)
			w.RegisterActivity(activities.ResolveCheckpoint)
			if err := w.Start(); err != nil {
				log.Error("worker start", "err", err)
			} else {
				defer w.Stop()
				log.Info("temporal worker started", "queue", cfg.Temporal.TaskQueue)
			}
		}
	}

	svc := app.New(st, bus, starter, signaler)
	gs := grpc.NewServer()
	executionv1.RegisterExecutionServiceServer(gs, grpcsrv.New(svc))
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
