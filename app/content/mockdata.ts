// ── Mock Data ─────────────────────────────────────────────────────────────────

export const INITIAL_NODES = [
  { id: "lb-lon", type: "loadbalancer", label: "LB-LON-01", city: "London", country: "UK", lat: 51.5, lng: -0.1, status: "healthy", connections: 3200, maxConnections: 5000, weight: 0.64, p95: 28, p99: 41, memory: 0.51, spawnedAt: Date.now() - 86400000 },
  { id: "lb-nyc", type: "loadbalancer", label: "LB-NYC-01", city: "New York", country: "US", lat: 40.7, lng: -74.0, status: "warning", connections: 4100, maxConnections: 5000, weight: 0.82, p95: 34, p99: 67, memory: 0.74, spawnedAt: Date.now() - 72000000 },
  { id: "lb-sin", type: "loadbalancer", label: "LB-SIN-01", city: "Singapore", country: "SG", lat: 1.35, lng: 103.8, status: "healthy", connections: 1800, maxConnections: 5000, weight: 0.36, p95: 22, p99: 35, memory: 0.38, spawnedAt: Date.now() - 50000000 },
  { id: "redis-lon-1", type: "redis", label: "RDS-LON-01", city: "London", country: "UK", lat: 51.52, lng: -0.08, status: "healthy", connections: 2100, maxConnections: 5000, weight: 0.42, p95: 12, p99: 19, memory: 0.45, spawnedAt: Date.now() - 80000000, redis: { version: "7.2.3", mode: "standalone", role: "master", connectedClients: 48, usedMemoryHuman: "312.4M", maxmemoryHuman: "2.00G", hitRate: 0.94, opsPerSec: 12400, evictedKeys: 0, keyspaceHits: 892341, keyspaceMisses: 54231 } },
  { id: "redis-lon-2", type: "redis", label: "RDS-LON-02", city: "London", country: "UK", lat: 51.48, lng: -0.12, status: "healthy", connections: 1800, maxConnections: 5000, weight: 0.36, p95: 11, p99: 17, memory: 0.39, spawnedAt: Date.now() - 60000000, redis: { version: "7.2.3", mode: "standalone", role: "replica", connectedClients: 31, usedMemoryHuman: "278.1M", maxmemoryHuman: "2.00G", hitRate: 0.91, opsPerSec: 9800, evictedKeys: 2, keyspaceHits: 712000, keyspaceMisses: 71200 } },
  { id: "redis-nyc-1", type: "redis", label: "RDS-NYC-01", city: "New York", country: "US", lat: 40.72, lng: -73.98, status: "critical", connections: 4700, maxConnections: 5000, weight: 0.94, p95: 89, p99: 142, memory: 0.91, spawnedAt: Date.now() - 70000000, redis: { version: "7.2.3", mode: "standalone", role: "master", connectedClients: 312, usedMemoryHuman: "1.82G", maxmemoryHuman: "2.00G", hitRate: 0.71, opsPerSec: 48200, evictedKeys: 8821, keyspaceHits: 2100000, keyspaceMisses: 860000 } },
  { id: "redis-nyc-2", type: "redis", label: "RDS-NYC-02", city: "New York", country: "US", lat: 40.68, lng: -74.02, status: "spawning", connections: 120, maxConnections: 5000, weight: 0.02, p95: 8, p99: 12, memory: 0.04, spawnedAt: Date.now() - 45000, redis: { version: "7.2.3", mode: "standalone", role: "master", connectedClients: 2, usedMemoryHuman: "18.2M", maxmemoryHuman: "2.00G", hitRate: 0.0, opsPerSec: 12, evictedKeys: 0, keyspaceHits: 0, keyspaceMisses: 0 } },
  { id: "redis-sin-1", type: "redis", label: "RDS-SIN-01", city: "Singapore", country: "SG", lat: 1.37, lng: 103.82, status: "healthy", connections: 900, maxConnections: 5000, weight: 0.18, p95: 9, p99: 14, memory: 0.22, spawnedAt: Date.now() - 48000000, redis: { version: "7.2.3", mode: "standalone", role: "master", connectedClients: 19, usedMemoryHuman: "98.4M", maxmemoryHuman: "2.00G", hitRate: 0.96, opsPerSec: 3100, evictedKeys: 0, keyspaceHits: 412000, keyspaceMisses: 17200 } },
  { id: "app-lon", type: "app", label: "APP-LON-01", city: "London", country: "UK", lat: 51.505, lng: -0.05, status: "healthy", connections: 3200, maxConnections: 10000, weight: 0.32, p95: 27, p99: 38, memory: 0.41, spawnedAt: Date.now() - 82000000 },
  { id: "app-nyc", type: "app", label: "APP-NYC-01", city: "New York", country: "US", lat: 40.74, lng: -74.05, status: "warning", connections: 6800, maxConnections: 10000, weight: 0.68, p95: 45, p99: 78, memory: 0.63, spawnedAt: Date.now() - 71000000 },
  { id: "app-sin", type: "app", label: "APP-SIN-01", city: "Singapore", country: "SG", lat: 1.33, lng: 103.78, status: "healthy", connections: 2100, maxConnections: 10000, weight: 0.21, p95: 19, p99: 28, memory: 0.29, spawnedAt: Date.now() - 49000000 },
  { id: "db-lon", type: "database", label: "PG-LON-01", city: "London", country: "UK", lat: 51.51, lng: -0.15, status: "healthy", connections: 45, maxConnections: 200, weight: 0.225, p95: 4, p99: 8, memory: 0.55, spawnedAt: Date.now() - 90000000 },
  { id: "db-nyc", type: "database", label: "PG-NYC-01", city: "New York", country: "US", lat: 40.66, lng: -73.96, status: "healthy", connections: 82, maxConnections: 200, weight: 0.41, p95: 5, p99: 9, memory: 0.62, spawnedAt: Date.now() - 90000000 },
];

export const INITIAL_EDGES = [
  { from: "lb-lon", to: "app-lon", throughput: 0.64, active: true },
  { from: "lb-nyc", to: "app-nyc", throughput: 0.82, active: true },
  { from: "lb-sin", to: "app-sin", throughput: 0.36, active: true },
  { from: "app-lon", to: "redis-lon-1", throughput: 0.55, active: true },
  { from: "app-lon", to: "redis-lon-2", throughput: 0.42, active: true },
  { from: "app-nyc", to: "redis-nyc-1", throughput: 0.94, active: true },
  { from: "app-nyc", to: "redis-nyc-2", throughput: 0.08, active: true },
  { from: "app-sin", to: "redis-sin-1", throughput: 0.28, active: true },
  { from: "app-lon", to: "db-lon", throughput: 0.22, active: true },
  { from: "app-nyc", to: "db-nyc", throughput: 0.41, active: true },
  { from: "lb-lon", to: "lb-nyc", throughput: 0.15, active: true },
  { from: "lb-nyc", to: "lb-sin", throughput: 0.11, active: true },
];

export const MOCK_LOGS = [
  { id: 1, ts: "14:32:01.441", level: "ERROR", node: "RDS-NYC-01", msg: "maxmemory threshold exceeded — evicting keys (policy: allkeys-lru)" },
  { id: 2, ts: "14:32:00.112", level: "WARN", node: "LB-NYC-01", msg: "upstream latency spike detected: p99 crossed 100ms threshold" },
  { id: 3, ts: "14:31:58.903", level: "INFO", node: "RDS-NYC-02", msg: "replica sync complete — ready to accept connections" },
  { id: 4, ts: "14:31:55.210", level: "INFO", node: "APP-NYC-01", msg: "horizontal scale event triggered: spawning RDS-NYC-02" },
  { id: 5, ts: "14:31:48.003", level: "WARN", node: "RDS-NYC-01", msg: "connected clients approaching limit: 312/500" },
  { id: 6, ts: "14:31:32.881", level: "INFO", node: "LB-LON-01", msg: "health check passed — all upstreams healthy" },
  { id: 7, ts: "14:31:20.004", level: "INFO", node: "APP-SIN-01", msg: "deploy v2.4.1 complete — zero downtime rollout finished" },
  { id: 8, ts: "14:30:55.772", level: "DEBUG", node: "PG-NYC-01", msg: "slow query detected: 340ms on users.sessions (seq scan)" },
];

export const MOCK_DEPLOYMENTS = [
  { id: "d1", version: "v2.4.1", env: "production", node: "APP-SIN-01", status: "success", duration: "1m 42s", ts: "14:31:20", commit: "a3f9c12" },
  { id: "d2", version: "v2.4.1", env: "production", node: "APP-LON-01", status: "success", duration: "1m 38s", ts: "14:29:40", commit: "a3f9c12" },
  { id: "d3", version: "v2.4.1", env: "production", node: "APP-NYC-01", status: "running", duration: "0m 54s", ts: "14:32:08", commit: "a3f9c12" },
  { id: "d4", version: "v2.4.0", env: "staging", node: "APP-LON-01", status: "success", duration: "2m 01s", ts: "11:12:05", commit: "b8d2e41" },
  { id: "d5", version: "v2.3.9", env: "production", node: "APP-NYC-01", status: "failed", duration: "0m 28s", ts: "09:44:31", commit: "c1a0f88" },
];