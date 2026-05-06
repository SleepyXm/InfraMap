import * as d3 from "d3";
import * as topojson from "topojson-client";
import { InfraNode, TrafficEdge } from "@/types/infra";
import { statusColor, weightColor, throughputColor, nodeIcon } from "../types/styles";
import { useWorldMap } from "../hooks/WorldMap";

type Props = {
  nodes: InfraNode[];
  visibleNodes: InfraNode[];
  visibleEdges: TrafficEdge[];
  mapData: ReturnType<typeof useWorldMap>; // or inline the type
  dims: { w: number; h: number };
  selected: InfraNode | null;
  hoveredEdge: string | null;
  onSelectNode: (node: InfraNode | null) => void;
  onHoverEdge: (id: string | null) => void;
};

export function InfraMap({ nodes, visibleNodes, visibleEdges, mapData, dims, selected, hoveredEdge, onSelectNode, onHoverEdge }: Props) {
  if (!mapData) return null;
  const { pathGen, land, borders, projectLatLng } = mapData;
  const getPos = (id: string) => {
    const n = nodes.find(x => x.id === id);
    return n ? projectLatLng(n.lat, n.lng) : null;
  };

  return (
    <svg width="100%" height="100%" style={{ display: "block", position: "absolute", inset: 0 }}>
      <defs>
        <pattern id="grid" width="40" height="40" patternUnits="userSpaceOnUse">
          <path d="M 40 0 L 0 0 0 40" fill="none" stroke="#0d1a26" strokeWidth="0.5" />
        </pattern>
      </defs>
      <rect width="100%" height="100%" fill="url(#grid)" />

      <path d={pathGen(land as any)} fill="#0d1e2f" stroke="#1a3550" strokeWidth={0.5} />
      <path d={pathGen(borders as any)} fill="none" stroke="#0e2033" strokeWidth={0.3} />

      {visibleEdges.map(edge => {
        const fp = getPos(edge.from);
        const tp = getPos(edge.to);
        if (!fp || !tp) return null;
        const [x1, y1] = fp;
        const [x2, y2] = tp;
        const edgeId = `${edge.from}-${edge.to}`;
        const isHovered = hoveredEdge === edgeId;
        const strokeW = 0.8 + edge.throughput * 2.5;
        const color = throughputColor(edge.throughput);
        const mx = (x1 + x2) / 2, my = (y1 + y2) / 2 - 30;
        const d = `M ${x1} ${y1} Q ${mx} ${my} ${x2} ${y2}`;
        return (
          <g key={edgeId}>
            <path d={d} fill="none" stroke="transparent" strokeWidth={12} style={{ cursor: "pointer" }}
              onMouseEnter={() => onHoverEdge(edgeId)} onMouseLeave={() => onHoverEdge(null)} />
            <path d={d} fill="none" stroke={color} strokeWidth={strokeW} opacity={0.3} />
            <path d={d} fill="none" stroke={color} strokeWidth={strokeW + 0.5}
              strokeDasharray="8 12" className="flow-line" opacity={isHovered ? 0.95 : 0.6} />
            {isHovered && (
              <text x={mx} y={my - 6} textAnchor="middle" fontSize="9" fill="#c8d4e0" fontFamily="IBM Plex Mono">
                {Math.round(edge.throughput * 100)}%
              </text>
            )}
          </g>
        );
      })}

      {visibleNodes.map(node => {
        const pos = projectLatLng(node.lat, node.lng);
        if (!pos) return null;
        const [x, y] = pos;
        const color = statusColor(node.status);
        const r = node.type === "loadbalancer" ? 9 : node.type === "database" ? 7 : 8;
        const isSelected = selected?.id === node.id;
        const circ = 2 * Math.PI * (r + 3);
        return (
          <g key={node.id} onClick={() => onSelectNode(isSelected ? null : node)} style={{ cursor: "pointer" }}>
            {node.status === "spawning" && (
              <circle cx={x} cy={y} r={r} fill="none" stroke="#4fc3f7" strokeWidth="1.5" className="spawning-ring" />
            )}
            {node.status === "critical" && (
              <circle cx={x} cy={y} r={r + 6} fill="#ff3b5c" opacity={0.15} className="critical-pulse" />
            )}
            {isSelected && (
              <circle cx={x} cy={y} r={r + 10} fill="none" stroke={color} strokeWidth="1" strokeDasharray="3 3" opacity={0.7} />
            )}
            <circle cx={x} cy={y} r={r + 8} fill={color} opacity={0.07} />
            <circle cx={x} cy={y} r={r} fill="#0d1520" stroke={color} strokeWidth={isSelected ? 2 : 1.5} />
            <circle cx={x} cy={y} r={r + 3} fill="none"
              stroke={weightColor(node.weight)} strokeWidth="1.5"
              strokeDasharray={`${node.weight * circ} 999`}
              opacity={0.55} transform={`rotate(-90 ${x} ${y})`} />
            <text x={x} y={y + 3.5} textAnchor="middle" fontSize="8" fill={color} style={{ userSelect: "none" }}>
              {nodeIcon(node.type)}
            </text>
            <text x={x} y={y + r + 12} textAnchor="middle" fontSize="8" fill="#4a6080" fontFamily="IBM Plex Mono" style={{ userSelect: "none" }}>
              {node.label}
            </text>
          </g>
        );
      })}
    </svg>
  );
}