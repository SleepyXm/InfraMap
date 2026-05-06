import { TYPE_MAP } from "../types/type";
import { weightColor, statusColor } from "../types/styles";

type NodeListProps = {
  nodes: any[];
  selected: any;
  onSelect: (node: any) => void;
  filter: string;
};

export function NodeList({ nodes, selected, onSelect, filter }: NodeListProps) {
  const visible = filter === "ALL" ? nodes : nodes.filter((n) => n.type === TYPE_MAP[filter]);

  return (
    <div style={{ flex: 1, overflowY: "auto", padding: "8px 10px" }}>
      {visible.map((node) => (
        <div
          key={node.id}
          onClick={() => onSelect(node)}
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            padding: "7px 8px",
            marginBottom: 3,
            background: selected?.id === node.id ? "#0e1e30" : "#0a1422",
            border: `1px solid ${selected?.id === node.id ? "#1e3050" : "#0e2033"}`,
            borderRadius: 2,
            cursor: "pointer",
            transition: "all 0.1s",
          }}
          onMouseEnter={(e) => {
            if (selected?.id !== node.id) e.currentTarget.style.borderColor = "#1e3050";
          }}
          onMouseLeave={(e) => {
            if (selected?.id !== node.id) e.currentTarget.style.borderColor = "#0e2033";
          }}
        >
          <div style={{ display: "flex", alignItems: "center", gap: 7 }}>
            <div
              style={{
                width: 5,
                height: 5,
                borderRadius: "50%",
                background: statusColor(node.status),
                flexShrink: 0,
                ...(node.status === "critical" ? { animation: "pulse 0.8s ease-in-out infinite" } : {}),
              }}
            />
            <div>
              <div style={{ fontSize: 10, color: "#c8d4e0" }}>{node.label}</div>
              <div style={{ fontSize: 8, color: "#4a6080" }}>{node.city}</div>
            </div>
          </div>
          <div style={{ textAlign: "right" }}>
            <div style={{ fontSize: 10, color: weightColor(node.weight) }}>
              {(node.weight * 100).toFixed(0)}%
            </div>
            <div style={{ fontSize: 8, color: "#4a6080" }}>{node.p95}ms</div>
          </div>
        </div>
      ))}
    </div>
  );
}