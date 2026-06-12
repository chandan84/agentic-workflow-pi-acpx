import { Routes, Route, Navigate, useLocation } from "react-router-dom";
import { Sidebar } from "./components/Sidebar";
import { Groups } from "./pages/Groups";
import { Resources } from "./pages/Resources";
import { Flows } from "./pages/Flows";
import { FlowEditor } from "./pages/FlowEditor";
import { Runs } from "./pages/Runs";
import { Audit } from "./pages/Audit";
import { Settings } from "./pages/Settings";

interface CrumbMap {
  [path: string]: string;
}

const CRUMBS: CrumbMap = {
  "/groups": "Agent Groups",
  "/resources": "Resource Bundles",
  "/flows": "Flows",
  "/runs": "Runs",
  "/audit": "Audit and Replay",
  "/settings": "Settings",
};

function Topbar(): JSX.Element {
  const location = useLocation();
  const top = "/" + location.pathname.split("/")[1];
  const label = CRUMBS[top] ?? "AWPA";
  return <div className="topbar">{label}</div>;
}

export function App(): JSX.Element {
  return (
    <div className="app-shell">
      <Sidebar />
      <div className="main-area">
        <Topbar />
        <div className="page">
          <Routes>
            <Route path="/" element={<Navigate to="/groups" replace />} />
            <Route path="/groups" element={<Groups />} />
            <Route path="/resources" element={<Resources />} />
            <Route path="/flows" element={<Flows />} />
            <Route path="/flows/:flowId/edit" element={<FlowEditor />} />
            <Route path="/runs" element={<Runs />} />
            <Route path="/runs/:runId" element={<Runs />} />
            <Route path="/audit" element={<Audit />} />
            <Route path="/audit/:runId" element={<Audit />} />
            <Route path="/settings" element={<Settings />} />
            <Route path="*" element={<div>Not found</div>} />
          </Routes>
        </div>
      </div>
    </div>
  );
}
