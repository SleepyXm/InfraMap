export function TestingPanel() {
  return (
    <div style={{ flex: 1, overflow: "auto", padding: 20 }}>
      <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: 12, marginBottom: 20 }}>
        {[
          { label: "SMOKE TEST", desc: "1 VU · 10s · sanity check", status: "READY", color: "#00e5a0" },
          { label: "LOAD TEST", desc: "100 VU · 30s · websocket", status: "READY", color: "#4fc3f7" },
          { label: "STRESS TEST", desc: "500 VU · 2m · ramp up", status: "READY", color: "#f5a623" },
          { label: "SOAK TEST", desc: "50 VU · 30m · endurance", status: "READY", color: "#a78bfa" },
          { label: "SPIKE TEST", desc: "0→800 VU · 10s burst", status: "READY", color: "#f5a623" },
          { label: "CHAOS TEST", desc: "kill random nodes · 60s", status: "SOON", color: "#ff3b5c" },
        ].map(t => (
          <div key={t.label} style={{ padding: 14, background: "#0a1422", border: `1px solid ${t.color}20`, borderRadius: 2 }}>
            <div style={{ display: "flex", justifyContent: "space-between", marginBottom: 8 }}>
              <span style={{ fontSize: 11, color: t.color, fontWeight: 600 }}>{t.label}</span>
              <span style={{ fontSize: 8, color: t.status === "SOON" ? "#2a3d50" : t.color, padding: "1px 6px", border: `1px solid ${t.status === "SOON" ? "#1e3050" : t.color}40`, borderRadius: 2 }}>{t.status}</span>
            </div>
            <div style={{ fontSize: 10, color: "#4a6080", marginBottom: 12 }}>{t.desc}</div>
            <button disabled={t.status === "SOON"} style={{ width: "100%", padding: "6px", background: t.status === "SOON" ? "transparent" : `${t.color}10`, border: `1px solid ${t.status === "SOON" ? "#1e3050" : t.color}40`, color: t.status === "SOON" ? "#2a3d50" : t.color, fontFamily: "inherit", fontSize: 10, cursor: t.status === "SOON" ? "not-allowed" : "pointer", borderRadius: 2, letterSpacing: "0.08em" }}>
              {t.status === "SOON" ? "COMING SOON" : "▶ RUN"}
            </button>
          </div>
        ))}
      </div>
      <div style={{ fontSize: 9, color: "#2a3d50", letterSpacing: "0.08em" }}>LAST RUN · LOAD TEST · 14:01:32 · 98.4% success rate · p95 34ms · p99 67ms</div>
    </div>
  );
}