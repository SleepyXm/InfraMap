export function CostPanel() {
  const rows = [
    { label: "Compute (App nodes)", monthly: 312, daily: 10.4, trend: "+2.1%" },
    { label: "Redis (self-hosted)", monthly: 48, daily: 1.6, trend: "0.0%" },
    { label: "Database (Postgres)", monthly: 94, daily: 3.1, trend: "+0.8%" },
    { label: "Egress / bandwidth", monthly: 67, daily: 2.2, trend: "+5.4%" },
    { label: "Load balancers", monthly: 29, daily: 0.97, trend: "0.0%" },
  ];
  const total = rows.reduce((a, r) => a + r.monthly, 0);
  return (
    <div style={{ flex: 1, overflow: "auto", padding: 20 }}>
      <div style={{ display: "flex", gap: 12, marginBottom: 20 }}>
        {[{ label: "MTD COST", value: `$${total}`, color: "#c8d4e0" }, { label: "PROJECTED", value: `$${Math.round(total * 1.08)}`, color: "#f5a623" }, { label: "BUDGET", value: "$800", color: "#00e5a0" }, { label: "REMAINING", value: `$${800 - total}`, color: "#00e5a0" }].map(s => (
          <div key={s.label} style={{ flex: 1, padding: "12px 14px", background: "#0a1422", border: "1px solid #0e2033", borderRadius: 2 }}>
            <div style={{ fontSize: 9, color: "#4a6080", marginBottom: 4 }}>{s.label}</div>
            <div style={{ fontSize: 20, color: s.color }}>{s.value}</div>
          </div>
        ))}
      </div>
      <table style={{ width: "100%", borderCollapse: "collapse", fontSize: 11 }}>
        <thead>
          <tr style={{ borderBottom: "1px solid #0e2033" }}>
            {["RESOURCE", "MONTHLY", "DAILY AVG", "TREND"].map(h => (
              <th key={h} style={{ padding: "6px 10px", textAlign: "left", fontSize: 9, color: "#4a6080", letterSpacing: "0.08em", fontWeight: 400 }}>{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map(r => (
            <tr key={r.label} style={{ borderBottom: "1px solid #090f18" }}>
              <td style={{ padding: "8px 10px", color: "#c8d4e0" }}>{r.label}</td>
              <td style={{ padding: "8px 10px", color: "#c8d4e0" }}>${r.monthly}</td>
              <td style={{ padding: "8px 10px", color: "#4a6080" }}>${r.daily}</td>
              <td style={{ padding: "8px 10px", color: r.trend.startsWith("+") ? "#f5a623" : "#00e5a0" }}>{r.trend}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}