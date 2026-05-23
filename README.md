# inframap
 
Real-time infrastructure topology visualisation and operational control plane.
 
InfraMap gives you a live geographic view of your entire infrastructure — nodes, traffic, health state, and latency — with the controls to provision, test, and manage it from a single interface. The interface is open. The plugins are what make it useful.
 
---
 
## Overview
 
Most observability tools tell you a number is bad. InfraMap tells you where it's bad, what's around it, and what's trending toward failure before it gets there.
 
The map is the interface. Everything — provisioning, chaos testing, deployment tracking, alerting — is anchored to the physical and logical topology of your system.
 
InfraMap is a frontend service. It is intentionally thin: no backend logic lives here. All infrastructure state, metrics, and control surfaces are served by the Go backend and extended through the plugin system.
 
---
 
## Plugin System
 
Plugins are how InfraMap connects to real infrastructure. The interface exposes the topology — plugins are what populate it.
 
Each plugin targets a specific service type and implements a shared contract:
 
- **Discover** — locate instances of this service type in the environment
- **Introspect** — read schema, structure, and state
- **Health** — report reachability and operational status
- **Secure access** — declare credential requirements and scope
Plugins connect through a secure access layer. Credentials are never stored in the interface — they are declared by the plugin and passed through the backend's access control policy.
 
### Available Plugins
 
| Plugin | Service | Notes |
|---|---|---|
| `supabase` | Supabase | First-party plugin. Uses the Supabase Management API and scoped service role. Fills the gap that standard Postgres introspection tools can't reach on managed Supabase instances. |
| `postgres` | PostgreSQL | Standard introspection via `pg_stat` and `information_schema`. Works with self-hosted and most managed deployments. |
| `redis` | Redis | Native `INSTANCE.WEIGHT` scoring, connection metrics, pub/sub topology. |
| `mongodb` | MongoDB | Collection-level introspection and health state. |
| `mysql` | MySQL / MariaDB | Standard introspection via `information_schema`. |
 
More plugins are in progress. See the roadmap.
 
### Writing a Plugin
 
Plugins follow a defined interface in the Go backend. The frontend consumes plugin output through the same WebSocket topology feed regardless of source — adding a plugin does not require frontend changes.
 
Documentation for the plugin interface is maintained in the backend repository.
 
---
 
## Features
 
### Infrastructure Topology Map
 
- Geographic world map with equirectangular projection
- Nodes plotted at actual datacenter coordinates
- Animated traffic edges — line weight and colour scale with throughput
- Layer toggles — Redis, Load Balancer, App, Database — independently shown or hidden
- Spawning animation on new instances coming online
- Critical pulse on nodes exceeding 90% weight threshold
### Node Health & Metrics
 
Each node exposes:
 
- `INSTANCE.WEIGHT` — a 0–1 composite load score derived from connection count, memory pressure, and command latency
- Weight arc rendered around the node — green to amber to red
- p95 and p99 latency
- Connection count against max connections
- Memory utilisation
- Uptime since spawn
### Predictive Alerting
 
InfraMap alerts on trajectory, not threshold breach. If a node's weight is trending toward ceiling, the alert fires with an estimated time to breach — before degradation starts.
 
Alert targets: native InfraMap feed, Grafana, AWS CloudWatch, AWS SNS.
 
### Provisioning & Control
 
Directly from the map:
 
- Spawn a new node into any region
- Shard a dummy Redis instance with synthetic data
- Replicate a dummy database shard
- Shadow prod — provision a full synthetic replica of the current production topology
### Chaos Testing
 
A dedicated chaos layer, visually separate from production, operated from the same interface. Right-click any node to introduce latency, kill the node, saturate connections, or invalidate cache.
 
Red team mode locks the session to the chaos environment, logs every action, and generates a report on completion.
 
### Deployment Visibility
 
- Active deployments shown as an overlay on the topology
- Per-node deploy status and ordering
- Rollback controls per deployment
- Deployment history timeline — correlate a deploy with a latency spike directly on the map
---
 
## Architecture
 
InfraMap is a Next.js frontend. The Go backend maintains WebSocket connections to each infrastructure node, streams state to the frontend in real time, and exposes provisioning and chaos controls as authenticated internal endpoints.
 
```
Infrastructure nodes
        ↓
Go backend (WebSocket aggregator + plugin host)
        ↓
InfraMap frontend (live map + controls)
        ↓
Grafana / CloudWatch / SNS (alert emission)
```
 
### Frontend Structure
 
```
inframap/
├── app/          # Next.js app router — pages and layouts
├── hooks/        # WebSocket connection, metrics, provisioning state
├── lanes/        # Layout and routing
├── styles/       # Design tokens and global styles
├── map/          # Topology map, node rendering, edge animation
└── content/      # Panels, detail views, control surfaces
```
 
### Backend Endpoints
 
| Endpoint | Description |
|---|---|
| `WS /ws/topology` | Live node and edge state stream |
| `POST /provision/node` | Spawn a new node |
| `POST /provision/redis/shard` | Create a dummy Redis shard |
| `POST /provision/db/shard` | Create a dummy DB shard |
| `POST /chaos/latency` | Introduce latency on a node |
| `POST /chaos/kill` | Kill a node |
| `POST /chaos/saturate` | Saturate connections |
| `POST /chaos/invalidate` | Invalidate cache |
 
All endpoints require authentication via shared secret header.
 
---
 
## Node Types
 
| Symbol | Type | Description |
|---|---|---|
| ⬡ | Redis | Cache and pub/sub layer |
| ◈ | Load Balancer | Internal and host balancer nodes |
| ▣ | App | Application service instances |
| ⬟ | Database | Postgres primary and replica nodes |
 
## Health States
 
| State | Description |
|---|---|
| `healthy` | Operating within normal parameters |
| `warning` | Weight between 75–90%, trending toward pressure |
| `critical` | Weight above 90%, intervention likely required |
| `spawning` | New instance coming online, not yet at capacity |
| `draining` | Instance being taken out of rotation gracefully |
 
---
 
## Getting Started
 
### Prerequisites
 
- Node.js 18+
- npm or pnpm
### Installation
 
```bash
git clone https://github.com/your-org/inframap.git
cd inframap
npm install
npm run dev
```
 
Open http://localhost:3000.
 
### Environment Variables
 
```env
RESEND_API_KEY=
NEXT_PUBLIC_BACKEND_URL=
NEXT_PUBLIC_SERVICE_NAME=inframap-frontend
```
 
---
 
## Related Projects
 
- **inframap-simulacrum** — developer environment runtime. Seeds, simulates, and narrates. Produces the dummy infrastructure that inframap's chaos layer and shadow prod features operate against.
- **Redis Observability Module** — native `INSTANCE.WEIGHT` scoring and alert emission
- **Redis Spawner** — session-aware instance routing and dynamic provisioning
- **Internal Load Balancer** — service discovery and health-based routing
- **Host Balancer** — customer-facing LB communication and failover coordination
---
 
## Contributing
 
1. Fork the repository
2. Create a branch from `main`
3. Open a pull request with a clear description of what changed and why
If you're building a plugin, refer to the plugin interface documentation in the backend repository.
 
---
 
## License
 
MIT
