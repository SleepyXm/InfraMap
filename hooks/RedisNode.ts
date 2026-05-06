import { useState, useEffect, useRef } from "react";


// ── Redis Integration Hook ────────────────────────────────────────────────────
// Swap REDIS_ENDPOINT to your backend route e.g. "http://localhost:3001/api/redis/info"
// Your backend should run: redis-cli INFO and return the parsed fields as JSON.
// Shape: { connectedClients, usedMemoryHuman, maxmemoryHuman, opsPerSec, hitRate,
//          evictedKeys, keyspaceHits, keyspaceMisses, version, role, mode }

const REDIS_ENDPOINT = null; // set to your endpoint to go live

export default function useRedisNode(nodeId: string, mockData: any) {
  const [data, setData] = useState(mockData || null);
  const [live, setLive] = useState(false);

  useEffect(() => {
    if (!REDIS_ENDPOINT) return; // mock mode
    let cancelled = false;
    const poll = async () => {
      try {
        const res = await fetch(`${REDIS_ENDPOINT}?node=${nodeId}`);
        if (!res.ok) return;
        const json = await res.json();
        if (!cancelled) { setData(json); setLive(true); }
      } catch (_) {}
    };
    poll();
    const iv = setInterval(poll, 2000);
    return () => { cancelled = true; clearInterval(iv); };
  }, [nodeId]);

  return { data, live };
}



