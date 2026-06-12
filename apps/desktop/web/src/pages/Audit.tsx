import { useParams } from "react-router-dom";

interface AuditRow {
  ts: string;
  runId: string;
  actor: string;
  action: string;
  detail: string;
}

// TODO(generated): auditClient.listEvents({ runId, page }) and
// auditClient.tailEvents({}) for the live tail (feature-flagged: auditTail).
const MOCK_AUDIT: AuditRow[] = [
  {
    ts: "2026-06-10T09:02:11Z",
    runId: "run_001",
    actor: "orchestrator",
    action: "checkpoint.requested",
    detail: "node=approve approvers=role:reviewer timeout=24h",
  },
  {
    ts: "2026-06-10T08:44:02Z",
    runId: "run_001",
    actor: "agt_dev",
    action: "segment.completed",
    detail: "segment=2 nodes=fix artifacts=1",
  },
  {
    ts: "2026-06-10T08:15:33Z",
    runId: "run_001",
    actor: "agt_triage",
    action: "segment.started",
    detail: "segment=1 nodes=triage,severity",
  },
  {
    ts: "2026-06-09T16:21:40Z",
    runId: "run_002",
    actor: "orchestrator",
    action: "run.completed",
    detail: "duration=21m segments=1",
  },
];

export function Audit(): JSX.Element {
  const { runId } = useParams();
  const rows = runId ? MOCK_AUDIT.filter((r) => r.runId === runId) : MOCK_AUDIT;

  return (
    <div>
      <div className="row" style={{ justifyContent: "space-between", marginBottom: 16 }}>
        <div className="muted">
          {rows.length} events{runId ? ` for ${runId}` : ""} · Temporal history and
          acpx trace bundles link from each row once wired up
        </div>
        <button
          onClick={() => {
            // TODO(generated): trigger export of acpx trace bundle for run
            window.alert("ExportTrace not yet wired up; awaiting generated client");
          }}
        >
          Export trace
        </button>
      </div>
      <div className="card" style={{ padding: 0 }}>
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 13 }}>
          <thead>
            <tr style={{ textAlign: "left", borderBottom: "1px solid var(--border)" }}>
              <th style={{ padding: 8 }}>Time</th>
              <th style={{ padding: 8 }}>Run</th>
              <th style={{ padding: 8 }}>Actor</th>
              <th style={{ padding: 8 }}>Action</th>
              <th style={{ padding: 8 }}>Detail</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((r, i) => (
              <tr key={i} style={{ borderBottom: "1px solid var(--border)" }}>
                <td style={{ padding: 8, whiteSpace: "nowrap" }}>{r.ts}</td>
                <td style={{ padding: 8 }}>{r.runId}</td>
                <td style={{ padding: 8 }}>{r.actor}</td>
                <td style={{ padding: 8, fontWeight: 600 }}>{r.action}</td>
                <td style={{ padding: 8 }} className="muted">
                  {r.detail}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
