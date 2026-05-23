# Roadmap

InfraMap ships as an open interface. The plugin system is where the value compounds — each new plugin extends what the map can see and control without touching the frontend.

---

## Phase 1 — Live Backend Integration

Wire the frontend to real infrastructure data, replacing mock state.

- [ ] WebSocket data feed from Go backend replacing mock data
- [ ] Live node and edge state rendering from real topology
- [ ] `useTopology` reconnection with exponential backoff
- [ ] Per-node metric subscriptions via `useMetrics`
- [ ] Authentication via shared secret header on all backend endpoints

---

## Phase 2 — Provisioning & Chaos Controls

Connect the control surfaces to real backend actions.

- [ ] Spawn node wired to `/provision/node`
- [ ] Dummy Redis shard via `/provision/redis/shard`
- [ ] Dummy DB shard via `/provision/db/shard`
- [ ] Chaos layer: latency, kill, saturate, invalidate wired to backend
- [ ] Red team session logging and report generation
- [ ] Shadow prod — full synthetic topology replica on demand

---

## Phase 3 — Plugin System

The plugin system is the core extension surface. The interface consumes plugin output through the same WebSocket feed regardless of source — adding a plugin requires no frontend changes.

- [ ] Plugin interface defined in Go backend
- [ ] Secure access layer — credential declaration and scoping per plugin
- [ ] `supabase` plugin — Management API + scoped service role (first-party)
- [ ] `postgres` plugin — `pg_stat` and `information_schema` introspection
- [ ] `redis` plugin — `INSTANCE.WEIGHT` scoring, connection metrics, pub/sub topology
- [ ] `mongodb` plugin — collection-level introspection and health
- [ ] `mysql` plugin — `information_schema` introspection
- [ ] Plugin registry — list, enable, and configure plugins from the interface
- [ ] Plugin documentation and authoring guide

---

## Phase 4 — Deployment Visibility

Surface deployment state directly on the topology map.

- [ ] Active deployment overlay on the map
- [ ] Per-node deploy status and ordering
- [ ] Rollback controls per deployment
- [ ] Deployment history timeline — correlate deploys with metric spikes

---

## Phase 5 — Analysis & Alerting

- [ ] Overhead analyser — underutilised nodes, over-provisioned regions, traffic imbalances
- [ ] Analytics aggregator — p95/p99 across full topology, regional breakdown, connection trends
- [ ] Predictive alerting wired to live data (trajectory-based, not threshold-based)
- [ ] Alert emission to Grafana, AWS CloudWatch, AWS SNS

---

## Future Considerations

- Multi-tenant support for external infrastructure customers
- simulacrum integration — chaos layer and shadow prod seeded directly from simulacrum scenarios
- Additional plugins: Elasticsearch, Kafka, RabbitMQ, ClickHouse
- Historical topology replay — step back through infrastructure state over time
- Mobile-optimised map view
