import { useEffect } from "react";
import { useParams } from "react-router-dom";
import {
  useAppStore,
  type CheckpointRequest,
  type RunSummary,
} from "../state/store";

const MOCK_RUNS: RunSummary[] = [
  {
    id: "run_001",
    flowId: "flw_bug_triage",
    flowVersionId: "fv_3",
    status: "awaiting_checkpoint",
    currentNode: "approve",
    startedAt: "2026-06-10T08:15:00Z",
    updatedAt: "2026-06-10T09:02:00Z",
  },
  {
    id: "run_002",
    flowId: "flw_release_notes",
    flowVersionId: "fv_1",
    status: "completed",
    currentNode: "done",
    startedAt: "2026-06-09T16:00:00Z",
    updatedAt: "2026-06-09T16:21:00Z",
  },
];

const MOCK_CHECKPOINTS: CheckpointRequest[] = [
  {
    id: "ckp_001",
    runId: "run_001",
    nodeId: "approve",
    prompt: "Approve the fix for PAY-1432 before merge?",
    status: "awaiting",
    deadline: "2026-06-11T09:02:00Z",
  },
];

const STATUS_COLOR: Record<RunSummary["status"], string> = {
  pending: "#6b7280",
  running: "#2563eb",
  awaiting_checkpoint: "#b45309",
  completed: "#047857",
  failed: "#dc2626",
  cancelled: "#6b7280",
};

export function Runs(): JSX.Element {
  const { runId } = useParams();
  const runs = useAppStore((s) => s.runs);
  const checkpoints = useAppStore((s) => s.checkpoints);
  const setRuns = useAppStore((s) => s.setRuns);
  const setCheckpoints = useAppStore((s) => s.setCheckpoints);

  useEffect(() => {
    // TODO(generated): executionClient.listRuns({ page: { pageSize: 50 } })
    if (runs.length === 0) setRuns(MOCK_RUNS);
    // TODO(generated): executionClient.listCheckpoints({ status: AWAITING })
    if (checkpoints.length === 0) setCheckpoints(MOCK_CHECKPOINTS);
    // TODO(generated): executionClient.watchRun({ runId }) — server streaming
    // fan-out from NATS run events for the live replay view.
  }, [runs.length, checkpoints.length, setRuns, setCheckpoints]);

  const selected = runs.find((r) => r.id === runId) ?? null;

  return (
    <div className="grid-2">
      <div>
        <div className="card-title">Runs</div>
        {runs.map((r) => (
          <div
            className="card"
            key={r.id}
            style={{ borderColor: selected?.id === r.id ? "var(--primary)" : undefined }}
          >
            <div className="row" style={{ justifyContent: "space-between" }}>
              <div style={{ fontWeight: 600 }}>{r.id}</div>
              <span style={{ color: STATUS_COLOR[r.status], fontWeight: 600, fontSize: 12 }}>
                {r.status}
              </span>
            </div>
            <div className="muted" style={{ fontSize: 12 }}>
              {r.flowId} @ {r.flowVersionId} · node: {r.currentNode}
            </div>
            <div className="muted" style={{ fontSize: 12 }}>
              started {r.startedAt} · updated {r.updatedAt}
            </div>
          </div>
        ))}
      </div>
      <div>
        <div className="card-title">Pending checkpoints</div>
        {checkpoints.length === 0 && <div className="muted">None awaiting approval.</div>}
        {checkpoints.map((c) => (
          <div className="card" key={c.id}>
            <div style={{ fontWeight: 600 }}>
              {c.runId} · {c.nodeId}
            </div>
            <div style={{ margin: "8px 0" }}>{c.prompt}</div>
            <div className="muted" style={{ fontSize: 12 }}>
              deadline: {c.deadline} · status: {c.status}
            </div>
            <div className="row" style={{ gap: 8, marginTop: 8 }}>
              <button
                className="primary"
                onClick={() => {
                  // TODO(generated): executionClient.resolveCheckpoint({
                  //   checkpointId: c.id, decision: APPROVED })
                  window.alert("ResolveCheckpoint not yet wired up; awaiting generated client");
                }}
              >
                Approve
              </button>
              <button
                onClick={() => {
                  // TODO(generated): executionClient.resolveCheckpoint({
                  //   checkpointId: c.id, decision: REJECTED })
                  window.alert("ResolveCheckpoint not yet wired up; awaiting generated client");
                }}
              >
                Reject
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
