import { NavLink } from "react-router-dom";

interface NavItem {
  to: string;
  label: string;
}

const NAV: NavItem[] = [
  { to: "/groups", label: "Groups and Agents" },
  { to: "/resources", label: "Resource Bundles" },
  { to: "/flows", label: "Flows" },
  { to: "/runs", label: "Runs" },
  { to: "/audit", label: "Audit and Replay" },
  { to: "/settings", label: "Settings" },
];

const sidebarStyle: React.CSSProperties = {
  borderRight: "1px solid var(--border)",
  background: "var(--card)",
  padding: "20px 12px",
  display: "flex",
  flexDirection: "column",
  gap: "4px",
};

const brandStyle: React.CSSProperties = {
  fontWeight: 700,
  padding: "0 12px 16px 12px",
  fontSize: "16px",
  color: "var(--text-strong)",
};

const linkBase: React.CSSProperties = {
  display: "block",
  padding: "8px 12px",
  borderRadius: "6px",
  color: "var(--text-strong)",
  fontWeight: 500,
};

const linkActive: React.CSSProperties = {
  ...linkBase,
  background: "var(--primary-soft)",
  color: "var(--primary)",
};

export function Sidebar(): JSX.Element {
  return (
    <aside style={sidebarStyle}>
      <div style={brandStyle}>AWPA</div>
      {NAV.map((item) => (
        <NavLink
          key={item.to}
          to={item.to}
          style={({ isActive }) => (isActive ? linkActive : linkBase)}
        >
          {item.label}
        </NavLink>
      ))}
    </aside>
  );
}
