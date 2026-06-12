import { useState } from "react";
import { useAppStore, type FeatureFlags } from "../state/store";

const FLAG_LABELS: Record<keyof FeatureFlags, string> = {
  streamingFlowAuthor: "Stream flow-author drafts as they generate",
  liveRunReplay: "Live run replay (server-streaming watch)",
  auditTail: "Live audit tail",
};

export function Settings(): JSX.Element {
  const gatewayUrl = useAppStore((s) => s.gatewayUrl);
  const flags = useAppStore((s) => s.flags);
  const setGatewayUrl = useAppStore((s) => s.setGatewayUrl);
  const setFlag = useAppStore((s) => s.setFlag);
  const [draftUrl, setDraftUrl] = useState(gatewayUrl);

  return (
    <div style={{ maxWidth: 560 }}>
      <div className="card">
        <div className="card-title">Gateway</div>
        <div className="muted" style={{ marginBottom: 8 }}>
          Connect-Web endpoint of the gateway-service. Overridable at build time
          via VITE_GATEWAY_URL.
        </div>
        <div className="row" style={{ gap: 8 }}>
          <input
            style={{ flex: 1 }}
            value={draftUrl}
            onChange={(e) => setDraftUrl(e.target.value)}
          />
          <button className="primary" onClick={() => setGatewayUrl(draftUrl)}>
            Apply
          </button>
        </div>
        <div className="muted" style={{ fontSize: 12, marginTop: 8 }}>
          active: {gatewayUrl}
        </div>
      </div>

      <div className="card">
        <div className="card-title">Feature flags</div>
        {(Object.keys(FLAG_LABELS) as Array<keyof FeatureFlags>).map((name) => (
          <label
            key={name}
            className="row"
            style={{ gap: 8, padding: "6px 0", cursor: "pointer" }}
          >
            <input
              type="checkbox"
              checked={flags[name]}
              onChange={(e) => setFlag(name, e.target.checked)}
            />
            <span>{FLAG_LABELS[name]}</span>
          </label>
        ))}
      </div>

      <div className="card">
        <div className="card-title">Authentication</div>
        <div className="muted">
          Auth mode is configured on the gateway (none | static-token). Token
          entry lands here when the gateway enables it.
        </div>
        {/* TODO(generated): persist token via Tauri secure storage IPC command */}
      </div>
    </div>
  );
}
