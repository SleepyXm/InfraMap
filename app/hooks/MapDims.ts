import { useState, useEffect, useRef } from "react";

export function useMapDims() {
  const [dims, setDims] = useState({ w: 900, h: 480 });
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const ro = new ResizeObserver(entries => {
      for (const e of entries) setDims({ w: e.contentRect.width, h: e.contentRect.height });
    });
    if (containerRef.current) ro.observe(containerRef.current);
    return () => ro.disconnect();
  }, []);

  return { dims, containerRef };
}