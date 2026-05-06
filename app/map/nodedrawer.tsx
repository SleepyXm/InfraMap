import { statusColor, nodeIcon, throughputColor } from "../types/styles";
import { MiniBar } from "../styles/styles";
import RedisPanel from "../lanes/Redis";

function formatUptime(spawnedAt: number) {
  const ms = Date.now() - spawnedAt;
  const h = Math.floor(ms / 3600000);
  const m = Math.floor((ms % 3600000) / 60000);
  if (h > 24) return `${Math.floor(h / 24)}d ${h % 24}h`;
  return `${h}h ${m}m`;
}

export function NodeDrawer({ node, edges, nodes, onClose }: { node: any; edges: any[]; nodes: any[]; onClose: () => void }) {
  if (!node) return null;
  const color = statusColor(node.status);

  return (
    <div style={{
      position: "absolute", bottom: 0, left: 0, right: 0,
      background: "#080e18", borderTop: `1px solid ${color}30`,
      zIndex: 20, animation: "slideup 0.22s ease-out",
      maxHeight: "42vh", display: "flex", flexDirection: "column"
    }}>
      {/* Drawer handle + header */}
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "10px 20px", borderBottom: "1px solid #0e2033", flexShrink: 0 }}>
        <div style={{ display: "flex", alignItems: "center", gap: 12 }}>
          <span style={{ fontSize: 16, color }}>{nodeIcon(node.type)}</span>
          <div>
            <span style={{ fontSize: 13, color, fontWeight: 600, letterSpacing: "0.06em" }}>{node.label}</span>
            <span style={{ fontSize: 10, color: "#4a6080", marginLeft: 10 }}>{node.city}, {node.country}</span>
          </div>
          <div style={{ display: "inline-flex", alignItems: "center", gap: 5, padding: "2px 8px", borderRadius: 2, background: `${color}15`, border: `1px solid ${color}35` }}>
            <div style={{ width: 4, height: 4, borderRadius: "50%", background: color }} />
            <span style={{ fontSize: 9, color, letterSpacing: "0.1em" }}>{node.status.toUpperCase()}</span>
          </div>
          <span style={{ fontSize: 9, color: "#2a3d50" }}>UP {formatUptime(node.spawnedAt)}</span>
        </div>
        <button onClick={onClose} style={{ background: "none", border: "1px solid #1e3050", color: "#4a6080", fontFamily: "inherit", fontSize: 10, padding: "4px 10px", cursor: "pointer", borderRadius: 2, letterSpacing: "0.08em" }}>ESC</button>
      </div>

      {/* Drawer body */}
      <div style={{ flex: 1, overflowY: "auto", padding: "14px 20px", display: "grid", gridTemplateColumns: "200px 200px 1fr 180px", gap: 20 }}>

        {/* Col 1 — Resource bars */}
        <div>
          <div style={{ fontSize: 9, color: "#4a6080", letterSpacing: "0.1em", marginBottom: 10 }}>RESOURCES</div>
          <MiniBar label="WEIGHT" value={node.weight} />
          <MiniBar label="MEMORY" value={node.memory} />
          <MiniBar label={`CONNECTIONS (${node.connections.toLocaleString()}/${node.maxConnections.toLocaleString()})`} value={node.connections / node.maxConnections} />
        </div>

        {/* Col 2 — Latency */}
        <div>
          <div style={{ fontSize: 9, color: "#4a6080", letterSpacing: "0.1em", marginBottom: 10 }}>LATENCY</div>
          <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
            <div style={{ flex: 1, padding: "8px", background: "#060a10", borderRadius: 2, textAlign: "center" }}>
              <div style={{ fontSize: 9, color: "#4a6080", marginBottom: 4 }}>P95</div>
              <div style={{ fontSize: 18, color: node.p95 > 60 ? "#ff3b5c" : node.p95 > 30 ? "#f5a623" : "#00e5a0" }}>{node.p95}<span style={{ fontSize: 9, color: "#4a6080" }}>ms</span></div>
            </div>
            <div style={{ flex: 1, padding: "8px", background: "#060a10", borderRadius: 2, textAlign: "center" }}>
              <div style={{ fontSize: 9, color: "#4a6080", marginBottom: 4 }}>P99</div>
              <div style={{ fontSize: 18, color: node.p99 > 100 ? "#ff3b5c" : node.p99 > 50 ? "#f5a623" : "#00e5a0" }}>{node.p99}<span style={{ fontSize: 9, color: "#4a6080" }}>ms</span></div>
            </div>
          </div>
        </div>

        {/* Col 3 — Redis or Traffic */}
        <div>
          {node.type === "redis" ? (
            <>
              <div style={{ fontSize: 9, color: "#4a6080", letterSpacing: "0.1em", marginBottom: 10 }}>REDIS METRICS</div>
              <RedisPanel node={node} />
            </>
          ) : (
            <>
              <div style={{ fontSize: 9, color: "#4a6080", letterSpacing: "0.1em", marginBottom: 10 }}>TRAFFIC</div>
              <div style={{ display: "flex", flexDirection: "column", gap: 4 }}>
                {edges.filter(e => e.from === node.id || e.to === node.id).map(e => {
                  const otherId = e.from === node.id ? e.to : e.from;
                  const other = nodes.find(n => n.id === otherId);
                  if (!other) return null;
                  const dir = e.from === node.id ? "→" : "←";
                  const tc = throughputColor(e.throughput);
                  return (
                    <div key={`${e.from}-${e.to}`} style={{ display: "flex", justifyContent: "space-between", alignItems: "center", padding: "5px 8px", background: "#060a10", borderRadius: 2, fontSize: 10 }}>
                      <span style={{ color: "#4a6080" }}>{dir} {other.label}</span>
                      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
                        <div style={{ width: 60, height: 2, background: "#0e2033", borderRadius: 2 }}>
                          <div style={{ width: `${e.throughput * 100}%`, height: "100%", background: tc, borderRadius: 2 }} />
                        </div>
                        <span style={{ color: tc, minWidth: 30, textAlign: "right" }}>{Math.round(e.throughput * 100)}%</span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </>
          )}
        </div>

        {/* Col 4 — Quick actions */}
        <div>
          <div style={{ fontSize: 9, color: "#4a6080", letterSpacing: "0.1em", marginBottom: 10 }}>ACTIONS</div>
          <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
            {[
              { label: "PING NODE", color: "#00e5a0" },
              { label: "VIEW LOGS", color: "#4fc3f7" },
              { label: "RESTART", color: "#f5a623" },
              { label: "DRAIN", color: "#9e9e9e" },
              ...(node.type === "redis" ? [{ label: "FLUSH CACHE", color: "#ff3b5c" }] : []),
            ].map(a => (
              <button key={a.label} style={{ padding: "6px 10px", background: "transparent", border: `1px solid ${a.color}30`, color: a.color, fontFamily: "inherit", fontSize: 10, cursor: "pointer", borderRadius: 2, letterSpacing: "0.08em", textAlign: "left", transition: "background 0.15s" }}
                onMouseEnter={e => e.currentTarget.style.background = `${a.color}10`}
                onMouseLeave={e => e.currentTarget.style.background = "transparent"}>
                {a.label}
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}