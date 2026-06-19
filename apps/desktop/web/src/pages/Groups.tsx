import { useEffect, useState } from "react";
import { agentsClient } from "../client/clients";
import { useAppStore, type Agent, type AgentGroup } from "../state/store";

const MOCK_GROUPS: AgentGroup[] = [
  {
    id: "grp_payments",
    name: "Payments",
    description: "Bug triage and release notes",
    createdAt: "2026-05-12T10:00:00Z",
  },
  {
    id: "grp_support",
    name: "Support",
    description: "Ticket triage",
    createdAt: "2026-05-20T10:00:00Z",
  },
];

const MOCK_AGENTS: Agent[] = [
  {
    id: "agt_qa",
    groupId: "grp_payments",
    name: "qa",
    role: "reviewer",
    piWorkspace: ".pi/qa",
    env: {},
  },
  {
    id: "agt_dev",
    groupId: "grp_payments",
    name: "dev",
    role: "implementer",
    piWorkspace: ".pi/dev",
    env: {},
  },
  {
    id: "agt_triage",
    groupId: "grp_support",
    name: "triage",
    role: "classifier",
    piWorkspace: ".pi/triage",
    env: {},
  },
];

export function Groups(): JSX.Element {
  const groups = useAppStore((s) => s.groups);
  const agents = useAppStore((s) => s.agents);
  const setGroups = useAppStore((s) => s.setGroups);
  const setAgents = useAppStore((s) => s.setAgents);
  const [selectedGroup, setSelectedGroup] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    // Call the real gateway first; fall back to mocks if the gateway is
    // unreachable (e.g. demo mode, no backend running). The gateway must
    // serve Connect protocol — see services/gateway/README.md.
    Promise.all([
      agentsClient.listGroups({ page: { limit: 50, cursor: "" } }),
      agentsClient.listAgents({ groupId: "", page: { limit: 100, cursor: "" } }),
    ])
      .then(([gRes, aRes]) => {
        if (cancelled) return;
        setGroups(
          gRes.groups.map((g) => ({
            id: g.id,
            name: g.name,
            description: g.description,
            createdAt: g.createdAt ? new Date(Number(g.createdAt.seconds) * 1000).toISOString() : "",
          })),
        );
        setAgents(
          aRes.agents.map((a) => ({
            id: a.id,
            groupId: a.groupId,
            name: a.name,
            role: a.role,
            piWorkspace: a.piWorkspace,
            env: a.env,
          })),
        );
      })
      .catch(() => {
        if (cancelled) return;
        if (groups.length === 0) setGroups(MOCK_GROUPS);
        if (agents.length === 0) setAgents(MOCK_AGENTS);
      });
    return () => {
      cancelled = true;
    };
  }, [groups.length, agents.length, setGroups, setAgents]);

  const visibleAgents = selectedGroup
    ? agents.filter((a) => a.groupId === selectedGroup)
    : agents;

  return (
    <div>
      <div className="row" style={{ justifyContent: "space-between", marginBottom: 16 }}>
        <div className="muted">{groups.length} groups</div>
        <button
          className="primary"
          onClick={async () => {
            const name = window.prompt("Group name?") ?? "";
            if (!name) return;
            const description = window.prompt("Description?") ?? "";
            try {
              const res = await agentsClient.createGroup({ name, description });
              if (res.group) {
                setGroups([
                  ...groups,
                  {
                    id: res.group.id,
                    name: res.group.name,
                    description: res.group.description,
                    createdAt: new Date().toISOString(),
                  },
                ]);
              }
            } catch (e) {
              window.alert(`Gateway unreachable: ${String(e)}`);
            }
          }}
        >
          New group
        </button>
      </div>

      <div className="grid-2">
        <div>
          <div className="card-title">Groups</div>
          {groups.map((g) => (
            <div
              key={g.id}
              className="card"
              style={{
                cursor: "pointer",
                borderColor: selectedGroup === g.id ? "var(--primary)" : undefined,
              }}
              onClick={() => setSelectedGroup(g.id)}
            >
              <div style={{ fontWeight: 600 }}>{g.name}</div>
              <div className="muted">{g.description}</div>
              <div className="muted" style={{ fontSize: 12, marginTop: 4 }}>
                {agents.filter((a) => a.groupId === g.id).length} agents
              </div>
            </div>
          ))}
        </div>
        <div>
          <div className="row" style={{ justifyContent: "space-between" }}>
            <div className="card-title">
              Agents{selectedGroup ? ` in ${selectedGroup}` : " (all)"}
            </div>
            <button
              onClick={() => {
                // TODO(generated): agentsClient.createAgent({ groupId, name, role, env })
                window.alert("CreateAgent not yet wired up; awaiting generated client");
              }}
            >
              New agent
            </button>
          </div>
          {visibleAgents.map((a) => (
            <div className="card" key={a.id}>
              <div style={{ fontWeight: 600 }}>{a.name}</div>
              <div className="muted">role: {a.role}</div>
              <div className="muted" style={{ fontSize: 12 }}>
                workspace: {a.piWorkspace}
              </div>
              {/* TODO(generated): agentsClient.attachResource({ agentId, resourceId, mountPath }) */}
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
