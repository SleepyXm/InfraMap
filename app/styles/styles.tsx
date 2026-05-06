import { weightColor } from "../types/styles";

export function Stat({ label, value, color }: { label: string; value: string | number; color?: string }) {
  return (
    <div style={{ textAlign: "center" }}>
      <div style={{ fontSize: 9, color: "#4a6080", letterSpacing: "0.08em" }}>{label}</div>
      <div style={{ fontSize: 13, color: color || "#c8d4e0", fontWeight: 500, marginTop: 1 }}>{value}</div>
    </div>
  );
}

export function MiniBar({ label, value, color }: { label: string; value: number; color?: string }) {
  return (
    <div style={{ marginBottom: 10 }}>
      <div style={{ display: "flex", justifyContent: "space-between", fontSize: 9, color: "#4a6080", marginBottom: 3 }}>
        <span>{label}</span>
        <span style={{ color: color || weightColor(value) }}>{(value * 100).toFixed(0)}%</span>
      </div>
      <div style={{ height: 2, background: "#0e2033", borderRadius: 2 }}>
        <div style={{ height: "100%", width: `${value * 100}%`, background: color || weightColor(value), borderRadius: 2, transition: "width 0.8s ease" }} />
      </div>
    </div>
  );
}

export function MetricPill({ label, value, color }: { label: string; value: string | number; color?: string }) {
  return (
    <div style={{ padding: "6px 10px", background: "#0a1422", border: `1px solid ${color || "#1e3050"}30`, borderRadius: 2, minWidth: 80 }}>
      <div style={{ fontSize: 9, color: "#4a6080", marginBottom: 2 }}>{label}</div>
      <div style={{ fontSize: 12, color: color || "#c8d4e0", fontWeight: 500 }}>{value}</div>
    </div>
  );
}