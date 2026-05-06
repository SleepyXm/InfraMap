"use client";

import { useState } from "react";
import { statusColor } from "./types/styles";
import DeploymentsPanel from "./lanes/DeploymentPanel";
import LogsPanel from "./lanes/LogsPanel";
import { CostPanel } from "./lanes/CostPanel";
import { TestingPanel } from "./lanes/TestingPanel";
import { INITIAL_NODES, INITIAL_EDGES } from "./content/mockdata";
import { NodeList } from "./lanes/InfraNodes";
import { TYPE_COLORS, TYPE_LABELS, TYPE_MAP } from "./types/type";
import { Stat } from "./styles/styles";
import { NodeDrawer } from "./map/nodedrawer";
import { useInfraNodes } from "./hooks/InfraNodes";
import { useWorldMap } from "./hooks/WorldMap";
import { InfraMap } from "./map/inframap";
import { useMapDims } from "./hooks/MapDims";


const NAV_TABS = ["MAP", "DEPLOYMENTS", "LOGS", "TESTING", "COST"];
const NODE_FILTERS = ["ALL", "LB", "APP", "REDIS", "DB"];



// ── Main Component ────────────────────────────────────────────────────────────

export default function InfraDashboard() {
  const { nodes, edges } = useInfraNodes();
  const { dims, containerRef } = useMapDims();
  const mapData = useWorldMap(dims);

  const [navTab, setNavTab] = useState<"MAP" | "DEPLOYMENTS" | "LOGS" | "TESTING" | "COST">("MAP");
  const [nodeFilter, setNodeFilter] = useState<"ALL" | "LB" | "APP" | "REDIS" | "DB">("ALL");
  const [selected, setSelected] = useState<InfraNode | null>(null);
  const [hoveredEdge, setHoveredEdge] = useState<string | null>(null);

  const totalConns = nodes.reduce((a, n) => a + n.connections, 0);
  const criticalCount = nodes.filter(n => n.status === "critical").length;
  const warningCount = nodes.filter(n => n.status === "warning").length;
  const spawnCount = nodes.filter(n => n.status === "spawning").length;

  const visibleNodes = nodeFilter === "ALL" ? nodes : nodes.filter(n => n.type === TYPE_MAP[nodeFilter]);
  const visibleEdges = edges.filter(e => {
    const fn = nodes.find(x => x.id === e.from);
    const tn = nodes.find(x => x.id === e.to);
    if (!fn || !tn) return false;
    return (nodeFilter === "ALL" || fn.type === TYPE_MAP[nodeFilter]) &&
           (nodeFilter === "ALL" || tn.type === TYPE_MAP[nodeFilter]);
  });

  return (
    <div style={{ fontFamily: "'IBM Plex Mono', 'Courier New', monospace", background: "#080c12", color: "#c8d4e0", height: "100vh", display: "flex", flexDirection: "column", overflow: "hidden" }}>
      <style>{`
        @import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@300;400;500;600&family=Syne:wght@700;800&display=swap');
        * { box-sizing: border-box; margin: 0; padding: 0; }
        ::-webkit-scrollbar { width: 3px; height: 3px; } ::-webkit-scrollbar-track { background: #0d1520; } ::-webkit-scrollbar-thumb { background: #1e3050; border-radius: 2px; }
        .node-dot { cursor: pointer; }
        .node-dot:hover { filter: brightness(1.4); }
        .edge-hit { cursor: pointer; }
        @keyframes pulse { 0%,100% { opacity:1; } 50% { opacity:0.4; } }
        @keyframes spawn-ring { 0% { r:8; opacity:0.8; } 100% { r:28; opacity:0; } }
        @keyframes flow { 0% { stroke-dashoffset: 40; } 100% { stroke-dashoffset: 0; } }
        .spawning-ring { animation: spawn-ring 1.4s ease-out infinite; }
        .critical-pulse { animation: pulse 0.8s ease-in-out infinite; }
        .flow-line { animation: flow 1.2s linear infinite; }
        @keyframes slideup { from { transform: translateY(20px); opacity:0; } to { transform: translateY(0); opacity:1; } }
        @keyframes fadein { from { opacity:0; } to { opacity:1; } }
        .nav-tab { cursor: pointer; background: none; border: none; font-family: inherit; font-size: 11px; letter-spacing: 0.1em; padding: 10px 16px; transition: all 0.15s; border-bottom: 2px solid transparent; }
        .nav-tab:hover { color: #c8d4e0; }
        .filter-btn { cursor: pointer; background: transparent; border: 1px solid #0e2033; font-family: inherit; font-size: 10px; letter-spacing: 0.08em; padding: 3px 10px; border-radius: 2px; transition: all 0.15s; }
        .filter-btn:hover { border-color: #1e3050; }
      `}</style>

      {/* Top bar */}
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "0 20px", borderBottom: "1px solid #0e2033", background: "#080c12", zIndex: 10, flexShrink: 0 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 0 }}>
          <div style={{ fontFamily: "'Syne', sans-serif", fontSize: 16, fontWeight: 800, letterSpacing: "0.12em", color: "#00e5a0", paddingRight: 20 }}>NOVA<span style={{ color: "#4fc3f7" }}>MAP</span></div>
          {NAV_TABS.map(tab => (
            <button key={tab} className="nav-tab" onClick={() => setNavTab(tab)}
              style={{ color: navTab === tab ? "#c8d4e0" : "#4a6080", borderBottomColor: navTab === tab ? "#00e5a0" : "transparent" }}>
              {tab}
            </button>
          ))}
        </div>
        <div style={{ display: "flex", gap: 20, fontSize: 11 }}>
          <Stat label="NODES" value={nodes.length.toString()} />
          <Stat label="CONNS" value={totalConns.toLocaleString()} />
          <Stat label="CRITICAL" value={criticalCount.toString()} color={criticalCount > 0 ? "#ff3b5c" : "#4a6080"} />
          <Stat label="WARN" value={warningCount.toString()} color={warningCount > 0 ? "#f5a623" : "#4a6080"} />
          <Stat label="SPAWN" value={spawnCount.toString()} color={spawnCount > 0 ? "#4fc3f7" : "#4a6080"} />
        </div>
      </div>

      {/* Main content */}
      <div style={{ display: "flex", flex: 1, overflow: "hidden" }}>

        {/* Sidebar */}
        <div style={{ width: 220, background: "#060a10", borderRight: "1px solid #0e2033", display: "flex", flexDirection: "column", flexShrink: 0, overflow: "hidden" }}>
          {/* Node filter tabs */}
          <div style={{ display: "flex", flexWrap: "wrap", gap: 4, padding: "10px 10px 8px", borderBottom: "1px solid #0e2033", flexShrink: 0 }}>
            {NODE_FILTERS.map(f => {
              const color = f === "ALL" ? "#c8d4e0" : TYPE_COLORS[TYPE_MAP[f]];
              const active = nodeFilter === f;
              return (
                <button key={f} className="filter-btn" onClick={() => setNodeFilter(f)}
                  style={{ color: active ? color : "#4a6080", borderColor: active ? `${color}50` : "#0e2033", background: active ? `${color}10` : "transparent" }}>
                  {f}
                </button>
              );
            })}
          </div>
          <NodeList nodes={nodes} selected={selected} onSelect={n => setSelected(selected?.id === n.id ? null : n)} filter={nodeFilter} />
          {/* Bottom ping status */}
          <div style={{ padding: "8px 12px", borderTop: "1px solid #0e2033", fontSize: 9, color: "#2a3d50", display: "flex", justifyContent: "space-between", flexShrink: 0 }}>
            <span>LIVE · 1.2s</span>
            <span>{new Date().toLocaleTimeString()}</span>
          </div>
        </div>

        {/* Main panel */}
        <div style={{ flex: 1, display: "flex", flexDirection: "column", overflow: "hidden", position: "relative" }}>

          {navTab === "MAP" && (
            <div ref={containerRef} style={{ flex: 1, position: "relative", overflow: "hidden" }}>
              <InfraMap
                nodes={nodes}
                visibleNodes={visibleNodes}
                visibleEdges={visibleEdges}
                mapData={mapData}
                dims={dims}
                selected={selected}
                hoveredEdge={hoveredEdge}
                onSelectNode={setSelected}
                onHoverEdge={setHoveredEdge}
              />
              <div style={{ position: "absolute", bottom: selected ? "43vh" : 12, left: 12, display: "flex", gap: 14, fontSize: 9, color: "#4a6080", letterSpacing: "0.08em", transition: "bottom 0.22s ease" }}>
                {["healthy", "warning", "critical", "spawning"].map(s => (
                  <div key={s} style={{ display: "flex", alignItems: "center", gap: 5 }}>
                    <div style={{ width: 5, height: 5, borderRadius: "50%", background: statusColor(s) }} />
                    {s.toUpperCase()}
                  </div>
                ))}
              </div>
            {selected && <NodeDrawer node={selected} edges={edges} nodes={nodes} onClose={() => setSelected(null)} />}
          </div>
          )}
          {navTab === "DEPLOYMENTS" && <DeploymentsPanel />}
          {navTab === "LOGS" && <LogsPanel />}
          {navTab === "TESTING" && <TestingPanel />}
          {navTab === "COST" && <CostPanel />}
        </div>
      </div>
    </div>
  );
}