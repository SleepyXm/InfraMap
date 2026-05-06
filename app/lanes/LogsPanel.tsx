import { useState } from "react";
import { MOCK_LOGS } from "../content/mockdata";

export default function LogsPanel() {
  const levelColors = { ERROR: "#ff3b5c", WARN: "#f5a623", INFO: "#00e5a0", DEBUG: "#4a6080" };
  const [filter, setFilter] = useState("ALL");
  const levels = ["ALL", "ERROR", "WARN", "INFO", "DEBUG"];
  const logs = filter === "ALL" ? MOCK_LOGS : MOCK_LOGS.filter(l => l.level === filter);
  return (
    <div style={{ flex: 1, overflow: "hidden", display: "flex", flexDirection: "column" }}>
      <div style={{ display: "flex", gap: 6, padding: "12px 20px", borderBottom: "1px solid #0e2033", flexShrink: 0 }}>
        {levels.map(l => (
          <button key={l} onClick={() => setFilter(l)} style={{ padding: "3px 10px", background: filter === l ? "#0e2033" : "transparent", border: `1px solid ${filter === l ? "#1e3050" : "#0e2033"}`, color: filter === l ? (levelColors[l] || "#c8d4e0") : "#4a6080", fontFamily: "inherit", fontSize: 9, cursor: "pointer", borderRadius: 2, letterSpacing: "0.08em" }}>{l}</button>
        ))}
        <div style={{ marginLeft: "auto", fontSize: 9, color: "#2a3d50", display: "flex", alignItems: "center", gap: 5 }}>
          <div style={{ width: 5, height: 5, borderRadius: "50%", background: "#00e5a0", animation: "pulse 1s ease-in-out infinite" }} />
          STREAMING
        </div>
      </div>
      <div style={{ flex: 1, overflow: "auto", padding: "8px 0" }}>
        {logs.map(log => (
          <div key={log.id} style={{ display: "flex", gap: 14, padding: "6px 20px", fontSize: 10, borderBottom: "1px solid #0a0f18", fontFamily: "monospace" }}
            onMouseEnter={e => e.currentTarget.style.background = "#080e18"}
            onMouseLeave={e => e.currentTarget.style.background = "transparent"}>
            <span style={{ color: "#2a3d50", minWidth: 90 }}>{log.ts}</span>
            <span style={{ color: levelColors[log.level], minWidth: 44 }}>{log.level}</span>
            <span style={{ color: "#4fc3f7", minWidth: 100 }}>{log.node}</span>
            <span style={{ color: "#8090a0" }}>{log.msg}</span>
          </div>
        ))}
      </div>
    </div>
  );
}