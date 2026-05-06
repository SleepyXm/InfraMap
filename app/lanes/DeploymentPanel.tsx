import { MOCK_DEPLOYMENTS } from "../content/mockdata";

export default function DeploymentsPanel() {
  const statusColors = { success: "#00e5a0", running: "#4fc3f7", failed: "#ff3b5c" };
  return (
    <div style={{ flex: 1, overflow: "auto", padding: 20 }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 16 }}>
        <div style={{ fontSize: 11, color: "#4a6080", letterSpacing: "0.1em" }}>RECENT DEPLOYMENTS</div>
        <button style={{ padding: "5px 14px", background: "#00e5a010", border: "1px solid #00e5a030", color: "#00e5a0", fontFamily: "inherit", fontSize: 10, cursor: "pointer", borderRadius: 2, letterSpacing: "0.08em" }}>+ DEPLOY</button>
      </div>
      <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 11 }}>
        <thead>
          <tr style={{ borderBottom: "1px solid #0e2033" }}>
            {["VERSION", "ENVIRONMENT", "NODE", "STATUS", "DURATION", "TIME", "COMMIT"].map(h => (
              <th key={h} style={{ padding: "6px 10px", textAlign: "left", fontSize: 9, color: "#4a6080", letterSpacing: "0.08em", fontWeight: 400 }}>{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {MOCK_DEPLOYMENTS.map(d => (
            <tr key={d.id} style={{ borderBottom: "1px solid #090f18" }}
              onMouseEnter={e => e.currentTarget.style.background = "#0a1422"}
              onMouseLeave={e => e.currentTarget.style.background = "transparent"}>
              <td style={{ padding: "8px 10px", color: "#c8d4e0" }}>{d.version}</td>
              <td style={{ padding: "8px 10px", color: "#4a6080" }}>{d.env}</td>
              <td style={{ padding: "8px 10px", color: "#4fc3f7" }}>{d.node}</td>
              <td style={{ padding: "8px 10px" }}>
                <span style={{ color: statusColors[d.status], display: "flex", alignItems: "center", gap: 5 }}>
                  {d.status === "running" && <span style={{ display: "inline-block", width: 5, height: 5, borderRadius: "50%", background: "#4fc3f7", animation: "pulse 0.8s ease-in-out infinite" }} />}
                  {d.status.toUpperCase()}
                </span>
              </td>
              <td style={{ padding: "8px 10px", color: "#4a6080" }}>{d.duration}</td>
              <td style={{ padding: "8px 10px", color: "#4a6080" }}>{d.ts}</td>
              <td style={{ padding: "8px 10px", color: "#4a6080", fontFamily: "monospace" }}>{d.commit}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}