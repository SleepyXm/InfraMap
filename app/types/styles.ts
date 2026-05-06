export function statusColor(status: string) {
  switch (status) {
    case "healthy":  return "#00e5a0";
    case "warning":  return "#f5a623";
    case "critical": return "#ff3b5c";
    case "spawning": return "#4fc3f7";
    case "draining": return "#9e9e9e";
  }
}

export function weightColor(w: number) {
  if (w < 0.5) return "#00e5a0";
  if (w < 0.8) return "#f5a623";
  return "#ff3b5c";
}

export function nodeIcon(type: string) {
  switch (type) {
    case "redis":        return "⬡";
    case "loadbalancer": return "◈";
    case "app":          return "▣";
    case "database":     return "⬟";
  }
}

export function throughputColor(t: number) {
  if (t < 0.5) return "rgba(0,229,160,0.6)";
  if (t < 0.8) return "rgba(245,166,35,0.6)";
  return "rgba(255,59,92,0.8)";
}


