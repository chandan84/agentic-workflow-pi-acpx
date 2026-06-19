// Command orchestrator is the entry point for the orchestrator service.
package main

import (
	"context"
	"encoding/json"
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
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/pgxconn"
	executionv1 "github.com/chandan84/agentic-workflow-pi-acpx/pkg/protogen/execution/v1"
	"github.com/chandan84/agentic-workflow-pi-acpx/pkg/runtime"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/app"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/flowsrc"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/grpcsrv"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/persist"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/ports"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/store"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/stub"
	"github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/temporalio"
	wfpkg "github.com/chandan84/agentic-workflow-pi-acpx/services/orchestrator/internal/workflow"

	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config is the orchestrator typed config.
type Config struct {
	config.Base `yaml:",inline"`
	Upstreams   struct {
		Flow string `yaml:"flow"`
	} `yaml:"upstreams"`
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
	} else if nb, err := events.DialNATS(cfg.NATS.URL); err == nil {
		bus = nb
		log.Info("bus: nats", "url", cfg.NATS.URL)
	} else {
		log.Warn("nats unavailable; in-memory bus", "err", err)
		bus = events.NewInMemoryBus()
	}
	defer func() { _ = bus.Close() }()

	var rt runtime.AgentRuntime
	if cfg.Stubs["runtime"] == "stub" {
		rt = runtime.NewStub()
		log.Info("runtime: stub")
	} else {
		rt = runtime.NewExec("acpx", "pi", "")
		log.Info("runtime: exec")
	}

	// Store: real Postgres adapter or in-memory stub.
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

	// FlowSource: gRPC client to flow-service, or stub.
	var fs flowsrc.FlowSource = flowsrc.NewStub()
	if cfg.Stubs["flowsrc"] != "stub" && cfg.Upstreams.Flow != "" {
		cc, err := grpc.NewClient(cfg.Upstreams.Flow, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			fs = flowsrc.NewGRPCClient(cc)
			defer func() { _ = cc.Close() }()
			log.Info("flowsrc: grpc", "addr", cfg.Upstreams.Flow)
		} else {
			log.Warn("flow-service dial failed; using stub flowsrc", "err", err)
		}
	}

	// Temporal client + worker + activities with real persistence and bus notify.
	var starter app.WorkflowStarter
	var signaler app.WorkflowSignaler
	persister := persist.New(st, bus)
	notify := func(kind string, payload any) {
		var subject string
		switch kind {
		case "segment_started":
			subject = events.SubjectRunSegmentStart
		case "segment_complete":
			subject = events.SubjectRunSegmentEnd
		default:
			subject = events.SubjectRunUpdated
		}
		_ = bus.Publish(context.Background(), subject, payload)
	}

	if cfg.Stubs["temporal"] == "stub" {
		s := temporalio.NewStub()
		starter, signaler = s, s
		log.Info("temporal: stub")
	} else if t, err := temporalio.New(cfg.Temporal.HostPort, cfg.Temporal.Namespace, cfg.Temporal.TaskQueue); err != nil {
		log.Warn("temporal dial failed; falling back to stub", "err", err)
		s := temporalio.NewStub()
		starter, signaler = s, s
	} else {
		starter, signaler = t, t
		defer t.Close()
		w := worker.New(t.C, cfg.Temporal.TaskQueue, worker.Options{})
		w.RegisterWorkflow(wfpkg.RunFlow)
		acts := &wfpkg.Activities{Runtime: rt, Notify: notify, Persist: persister}
		w.RegisterActivity(acts.ExecuteSegment)
		w.RegisterActivity(acts.RecordCheckpointAwaiting)
		w.RegisterActivity(acts.ResolveCheckpoint)
		if err := w.Start(); err != nil {
			log.Error("worker start", "err", err)
		} else {
			defer w.Stop()
			log.Info("temporal worker started", "queue", cfg.Temporal.TaskQueue)
		}
	}

	svc := app.New(st, bus, starter, signaler, fs)
	gs := grpc.NewServer()
	executionv1.RegisterExecutionServiceServer(gs, grpcsrv.NewServer(svc, bus))
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

	// Subscribe to segment events and update run status. This makes
	// `runs.run.status` reflect workflow progress even when Temporal lives
	// behind a slow consumer.
	go subscribeRunUpdates(ctx, bus, st, log)

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

// subscribeRunUpdates wires NATS run events back into the run.status / current_node.
// Idempotent: re-applying the same status/node is a no-op in Postgres.
func subscribeRunUpdates(ctx context.Context, bus events.Bus, st ports.Store, log interface{ Info(string, ...any) }) {
	subjects := map[string]string{
		events.SubjectRunSegmentEnd:    "running",
		events.SubjectRunCheckpointWait: "awaiting_checkpoint",
		events.SubjectRunCheckpointDone: "running",
		events.SubjectRunCompleted:     "completed",
		events.SubjectRunFailed:        "failed",
	}
	for subj, status := range subjects {
		subject, mapped := subj, status
		_, _ = bus.Subscribe(ctx, subject, func(_ string, data []byte) {
			var m map[string]any
			if err := json.Unmarshal(data, &m); err != nil {
				return
			}
			runID, _ := m["runId"].(string)
			if runID == "" {
				runID, _ = m["RunID"].(string)
			}
			if runID == "" {
				return
			}
			node, _ := m["nodeId"].(string)
			_ = st.UpdateRun(context.Background(), runID, mapped, node)
		})
	}
	log.Info("subscribed to run-update events", "count", len(subjects))
}
