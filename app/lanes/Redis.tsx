import useRedisNode from "../hooks/RedisNode";
import { MiniBar, MetricPill } from "../styles/styles";

export default function RedisPanel({ node }: { node: any }) {
  const { data, live } = useRedisNode(node.id, node.redis);
  if (!data) return <div style={{ fontSize: 10, color: "#4a6080", padding: 8 }}>No Redis data available</div>;

  const memUsed = parseFloat(data.usedMemoryHuman);
  const memMax = parseFloat(data.maxmemoryHuman);
  const memPct = memMax > 0 ? memUsed / memMax : 0;

  return (
    <div>
      <div style={{ display: "flex", alignItems: "center", gap: 8, marginBottom: 12 }}>
        <div style={{ fontSize: 9, color: "#4a6080" }}>REDIS {data.version} · {data.mode?.toUpperCase()} · {data.role?.toUpperCase()}</div>
        {!live && <div style={{ fontSize: 8, color: "#2a3d50", padding: "1px 5px", border: "1px solid #1e3050", borderRadius: 2 }}>MOCK</div>}
        {live && <div style={{ fontSize: 8, color: "#00e5a0", padding: "1px 5px", border: "1px solid #00e5a040", borderRadius: 2 }}>LIVE</div>}
      </div>

      <div style={{ display: "flex", flexWrap: "wrap", gap: 6, marginBottom: 14 }}>
        <MetricPill label="OPS/SEC" value={data.opsPerSec?.toLocaleString()} color="#4fc3f7" />
        <MetricPill label="CLIENTS" value={data.connectedClients} color="#a78bfa" />
        <MetricPill label="HIT RATE" value={`${((data.hitRate || 0) * 100).toFixed(1)}%`} color={data.hitRate > 0.9 ? "#00e5a0" : data.hitRate > 0.7 ? "#f5a623" : "#ff3b5c"} />
        <MetricPill label="EVICTED" value={data.evictedKeys?.toLocaleString()} color={data.evictedKeys > 0 ? "#f5a623" : "#4a6080"} />
      </div>

      <MiniBar label="MEMORY" value={memPct} color={memPct > 0.85 ? "#ff3b5c" : memPct > 0.7 ? "#f5a623" : "#00e5a0"} />

      <div style={{ display: "flex", justifyContent: "space-between", fontSize: 9, color: "#4a6080", marginBottom: 12, marginTop: -4 }}>
        <span>{data.usedMemoryHuman} used</span>
        <span>{data.maxmemoryHuman} max</span>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 6, fontSize: 9 }}>
        <div style={{ padding: "6px 8px", background: "#060a10", borderRadius: 2 }}>
          <div style={{ color: "#4a6080", marginBottom: 2 }}>KEYSPACE HITS</div>
          <div style={{ color: "#00e5a0" }}>{data.keyspaceHits?.toLocaleString()}</div>
        </div>
        <div style={{ padding: "6px 8px", background: "#060a10", borderRadius: 2 }}>
          <div style={{ color: "#4a6080", marginBottom: 2 }}>KEYSPACE MISSES</div>
          <div style={{ color: "#f5a623" }}>{data.keyspaceMisses?.toLocaleString()}</div>
        </div>
      </div>
    </div>
  );
}