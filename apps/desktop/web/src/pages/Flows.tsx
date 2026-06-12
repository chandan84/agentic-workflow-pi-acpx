import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { useAppStore, type FlowSummary } from "../state/store";

const MOCK_FLOWS: FlowSummary[] = [
  {
    id: "flw_bug_triage",
    name: "Bug triage & fix",
    description: "Triage incoming bug, implement fix, QA review, ship notes",
    createdAt: "2026-05-22T09:00:00Z",
  },
  {
    id: "flw_release_notes",
    name: "Release notes",
    description: "Summarize merged PRs into release notes with approval gate",
    createdAt: "2026-06-01T14:30:00Z",
  },
];

export function Flows(): JSX.Element {
  const flows = useAppStore((s) => s.flows);
  const setFlows = useAppStore((s) => s.setFlows);
  const [prompt, setPrompt] = useState("");
  const [authoring, setAuthoring] = useState(false);

  useEffect(() => {
    // TODO(generated): flowsClient.listFlows({ page: { pageSize: 50 } })
    if (flows.length === 0) setFlows(MOCK_FLOWS);
  }, [flows.length, setFlows]);

  return (
    <div>
      <div className="card" style={{ marginBottom: 24 }}>
        <div className="card-title">New flow from prompt</div>
        <div className="muted" style={{ marginBottom: 8 }}>
          Describe the workflow in plain language; the flow-author agent drafts
          a Flow IR you can refine on the canvas.
        </div>
        <textarea
          rows={3}
          style={{ width: "100%", boxSizing: "border-box" }}
          placeholder="e.g. When a bug report arrives, have the triage agent classify it, then…"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
        />
        <div className="row" style={{ justifyContent: "flex-end", marginTop: 8 }}>
          <button
            className="primary"
            disabled={prompt.trim().length === 0 || authoring}
            onClick={() => {
              // TODO(generated): flowsClient.authorFlow({ prompt }) — server
              // streaming; updates land as draft IR revisions.
              setAuthoring(true);
              window.setTimeout(() => {
                setAuthoring(false);
                window.alert("AuthorFlow not yet wired up; awaiting generated client");
              }, 300);
            }}
          >
            {authoring ? "Drafting…" : "Draft flow"}
          </button>
        </div>
      </div>

      <div className="row" style={{ justifyContent: "space-between", marginBottom: 16 }}>
        <div className="muted">{flows.length} flows</div>
      </div>
      {flows.map((f) => (
        <div className="card" key={f.id}>
          <div className="row" style={{ justifyContent: "space-between" }}>
            <div>
              <div style={{ fontWeight: 600 }}>{f.name}</div>
              <div className="muted">{f.description}</div>
            </div>
            <div className="row" style={{ gap: 8 }}>
              <Link to={`/flows/${f.id}/edit`}>
                <button>Open canvas</button>
              </Link>
              <button
                className="primary"
                onClick={() => {
                  // TODO(generated): executionClient.startRun({ flowId: f.id })
                  window.alert("StartRun not yet wired up; awaiting generated client");
                }}
              >
                Run
              </button>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
