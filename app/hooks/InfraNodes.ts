import { useState, useEffect, useRef } from "react";
import { INITIAL_NODES, INITIAL_EDGES } from "../content/mockdata";
import type { InfraNode } from "@/types/infra";

export function useInfraNodes() {
  const [nodes, setNodes] = useState<InfraNode[]>(INITIAL_NODES);
  const [edges] = useState(INITIAL_EDGES);
  const tickRef = useRef(0);

  useEffect(() => {
    const iv = setInterval(() => {
      tickRef.current++;
      setNodes(prev => prev.map(n => {
        const jitter = (Math.random() - 0.5) * 0.04;
        const newWeight = Math.max(0.01, Math.min(0.99, n.weight + jitter));
        const newConns = Math.max(0, Math.min(n.maxConnections, n.connections + Math.floor((Math.random() - 0.48) * 80)));
        let status = n.status;
        if (n.status === "spawning" && newConns > 500) status = "healthy";
        else if (n.status !== "spawning" && n.status !== "draining") {
          if (newWeight > 0.9) status = "critical";
          else if (newWeight > 0.75) status = "warning";
          else status = "healthy";
        }
        return {
          ...n,
          weight: newWeight,
          connections: newConns,
          status,
          p95: Math.max(5, n.p95 + Math.floor((Math.random() - 0.5) * 6)),
          p99: Math.max(8, n.p99 + Math.floor((Math.random() - 0.5) * 10)),
        };
      }));
    }, 1200);
    return () => clearInterval(iv);
  }, []);

  return { nodes, edges };
}