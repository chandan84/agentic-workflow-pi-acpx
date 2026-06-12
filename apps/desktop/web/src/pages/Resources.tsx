import { useEffect } from "react";
import { useAppStore, type ResourceBundle } from "../state/store";

const MOCK_RESOURCES: ResourceBundle[] = [
  {
    id: "res_mcp_fs",
    name: "filesystem-mcp",
    kind: "mcp",
    spec: '{"command":"npx","args":["@modelcontextprotocol/server-filesystem"]}',
  },
  {
    id: "res_secret_slack",
    name: "slack-bot-token",
    kind: "secret",
    spec: '{"ref":"vault:secret/slack#token"}',
  },
  {
    id: "res_tool_curl",
    name: "curl",
    kind: "tool",
    spec: '{"binary":"/usr/bin/curl"}',
  },
];

export function Resources(): JSX.Element {
  const resources = useAppStore((s) => s.resources);
  const setResources = useAppStore((s) => s.setResources);

  useEffect(() => {
    // TODO(generated): agentsClient.listResources({ page: { pageSize: 50 } })
    if (resources.length === 0) setResources(MOCK_RESOURCES);
  }, [resources.length, setResources]);

  return (
    <div>
      <div className="row" style={{ justifyContent: "space-between", marginBottom: 16 }}>
        <div className="muted">{resources.length} bundles</div>
        <button
          className="primary"
          onClick={() => {
            // TODO(generated): agentsClient.createResource({ name, kind, spec })
            window.alert("CreateResource not yet wired up; awaiting generated client");
          }}
        >
          New bundle
        </button>
      </div>
      {resources.map((r) => (
        <div className="card" key={r.id}>
          <div className="row" style={{ justifyContent: "space-between" }}>
            <div style={{ fontWeight: 600 }}>{r.name}</div>
            <span
              style={{
                background: "var(--primary-soft)",
                color: "var(--primary)",
                padding: "2px 8px",
                borderRadius: 999,
                fontSize: 12,
                fontWeight: 600,
              }}
            >
              {r.kind}
            </span>
          </div>
          <pre
            style={{
              background: "#f9fafb",
              padding: 8,
              borderRadius: 4,
              fontSize: 12,
              marginTop: 8,
              overflow: "auto",
            }}
          >
            {r.spec}
          </pre>
        </div>
      ))}
    </div>
  );
}
