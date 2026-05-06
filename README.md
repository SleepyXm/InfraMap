# InfraMap

Real-time infrastructure topology visualisation and operational control plane. InfraMap gives you a live geographic view of your entire infrastructure — nodes, traffic, health state, and latency — with the controls to provision, test, and manage it from a single interface.

-----

## Overview

Most observability tools tell you a number is bad. InfraMap tells you where it’s bad, what’s around it, and what’s trending toward failure before it gets there. It’s built for infrastructure that needs to stay alive under real load, not just look healthy in a dashboard.

The map is the interface. Everything — provisioning, chaos testing, deployment tracking, alerting — is anchored to the physical and logical topology of your system.

-----

## Features

### Infrastructure Topology Map

- Geographic world map with equirectangular projection
- Nodes plotted at actual datacenter coordinates
- Animated traffic edges between nodes, line weight and colour scale with throughput
- Layer toggles — Redis, Load Balancer, App, Database — independently shown or hidden
- Spawning animation on new instances coming online
- Critical pulse on nodes exceeding 90% weight threshold

### Node Health & Metrics

Each node exposes:

- `INSTANCE.WEIGHT` — a 0–1 composite load score derived from connection count, memory pressure, and command latency
- Weight arc rendered around the node on the map — green to amber to red
- p95 and p99 latency
- Connection count against max connections
- Memory utilisation
- Uptime since spawn

Click any node to open the detail panel with live metrics and traffic edges in and out.

### Predictive Alerting

InfraMap alerts on trajectory, not just threshold breach. If a node’s weight is trending toward ceiling, the alert fires with an estimated time to breach — giving the system or the engineer time to act before anything degrades.

Alert emission targets:

- Native InfraMap alerts
- Grafana
- AWS CloudWatch
- AWS SNS

### Provisioning & Control

Directly from the map:

- Spawn a new node into any region
- Shard a dummy Redis instance with synthetic data
- Replicate a dummy database shard
- Shadow prod — provision a full synthetic replica of the current production topology for isolated testing
- All provisioning actions logged to the provisioning history timeline

### Chaos Testing

InfraMap includes a dedicated chaos layer — visually separate from production, operated from the same interface.

Right-click any node in the chaos layer to:

- Introduce artificial latency
- Kill the node
- Saturate connections
- Invalidate cache

Results feed back into the metrics panel in real time. The chaos layer runs entirely against dummy infrastructure — production is never the target.

Red team mode locks the session to the chaos environment, logs every action taken, and generates a report on completion covering what was attacked, what was detected, and what wasn’t.

### Deployment Visibility

- Active deployments shown as an overlay on the topology
- Which nodes are receiving the deploy, in what order
- Rollback controls per deployment
- Deployment history timeline — correlate a deploy with a latency spike directly on the map

### Overhead Analyser

Identifies where compute is being wasted across the topology — underutilised nodes, over-provisioned regions, and traffic imbalances that could be redistributed for better efficiency.

### Analytics Aggregator

Aggregates metrics across all nodes and regions into a unified view — p95/p99 distribution across the full topology, regional traffic breakdown, connection count trends, and weight distribution over time.

-----

## Architecture

InfraMap is a frontend service. The backend is a Go service that:

- Maintains a WebSocket connection to each infrastructure node
- Streams node state, edge throughput, and metric updates to InfraMap in real time
- Exposes provisioning and chaos controls as authenticated internal endpoints
- Integrates with the Redis observability module for native `INSTANCE.WEIGHT` scoring

### Frontend Structure

```
inframap/
├── hooks/        # Data fetching, WebSocket connection, state management
├── lanes/        # Layout and routing
├── styles/       # Design tokens and global styles
├── map/          # Topology map, node rendering, edge animation
└── content/      # Panels, detail views, control surfaces
```

### Data Flow

```
Infrastructure nodes
        ↓
Go backend (WebSocket aggregator)
        ↓
InfraMap frontend (live map + controls)
        ↓
Grafana / CloudWatch / SNS (alert emission)
```

-----

## Node Types

|Symbol|Type         |Description                       |
|------|-------------|----------------------------------|
|⬡     |Redis        |Cache and pub/sub layer           |
|◈     |Load Balancer|Internal and host balancer nodes  |
|▣     |App          |Application service instances     |
|⬟     |Database     |Postgres primary and replica nodes|

-----

## Health States

|State     |Description                                    |
|----------|-----------------------------------------------|
|`healthy` |Node operating within normal parameters        |
|`warning` |Weight between 75–90%, trending toward pressure|
|`critical`|Weight above 90%, intervention likely required |
|`spawning`|New instance coming online, not yet at capacity|
|`draining`|Instance being taken out of rotation gracefully|

-----

## Alerting Philosophy

InfraMap does not alert when things break. It alerts when things are trending toward breaking. By the time a threshold breach fires in a traditional monitoring tool, degradation has already started. InfraMap’s predictive model gives you — or the system itself — a window to act first.

The target outcome is that engineers are notified of resolutions, not incidents.

-----

## Roadmap

- [ ] WebSocket data feed from Go backend replacing mock data
- [ ] Live provisioning controls wired to infrastructure spawner
- [ ] Chaos layer with right-click controls
- [ ] Red team session logging and report generation
- [ ] Deployment overlay and rollback controls
- [ ] Overhead analyser
- [ ] Analytics aggregator with historical timeline
- [ ] Multi-tenant support for external infrastructure customers

-----

## Related Projects

- **Trading Platform** — the primary consumer of the infrastructure suite
- **Redis Observability Module** — native `INSTANCE.WEIGHT` scoring and alert emission
- **Redis Spawner** — session-aware instance routing and dynamic provisioning
- **Internal Load Balancer** — service discovery and health-based routing
- **Host Balancer** — customer-facing LB communication and failover coordination

-----

# InfraMap

Real-time infrastructure topology visualisation and operational control plane built with Next.js.

-----

## Prerequisites

- Node.js 18+
- npm or pnpm

-----

## Getting Started

Clone the repository:

```bash
git clone https://github.com/your-org/inframap.git
cd inframap
```

Install dependencies:

```bash
npm install
# or
pnpm install
```

Start the development server:

```bash
npm run dev
# or
pnpm dev
```

Open <http://localhost:3000>.

-----

## Environment Variables

Create a `.env.local` file in the root of the project:

```env
# Email — required for alert notifications
RESEND_API_KEY=

# Backend service URL — required for live infrastructure data
NEXT_PUBLIC_BACKEND_URL=

# Internal identifier distinguishing frontend from backend in shared logging or tracing
NEXT_PUBLIC_SERVICE_NAME=inframap-frontend
```

All variables prefixed with `NEXT_PUBLIC_` are exposed to the browser. Variables without the prefix are server-side only.

-----

## Project Structure

```
inframap/
├── app/                  # Next.js app router — pages and layouts
├── hooks/                # Custom React hooks
│   ├── useTopology       # WebSocket connection and live node/edge state
│   ├── useMetrics        # Per-node metric subscriptions
│   └── useProvisioning   # Provisioning and chaos control actions
├── lanes/                # Layout components and route-level structure
├── styles/               # Global styles and design tokens
├── map/                  # Topology map
│   ├── InfraMap          # Root map component
│   ├── Node              # Individual node rendering — health state, weight arc, animations
│   ├── Edge              # Traffic edge rendering — throughput scaling, flow animation
│   └── Projection        # Lat/lng to SVG coordinate mapping
└── content/              # Panels and control surfaces
    ├── NodePanel         # Node detail view — metrics, traffic, uptime
    ├── ProvisionPanel    # Spawn, shard, and replicate controls
    ├── ChaosPanel        # Chaos layer controls and red team session management
    ├── DeployPanel       # Deployment overlay and rollback controls
    └── AlertPanel        # Alert feed and trajectory warnings
```

-----

## Map

The topology map renders infrastructure nodes at their actual geographic coordinates using an equirectangular projection. Nodes are typed, health-state coloured, and carry a weight arc showing live load.

### Node Types

|Symbol|Type         |Description                        |
|------|-------------|-----------------------------------|
|⬡     |Redis        |Cache and pub/sub instances        |
|◈     |Load Balancer|Internal LB and host balancer nodes|
|▣     |App          |Application service instances      |
|⬟     |Database     |Postgres primary and replica nodes |

### Node Health States

|State     |Description                                |
|----------|-------------------------------------------|
|`healthy` |Operating within normal parameters         |
|`warning` |Weight between 75–90%                      |
|`critical`|Weight above 90%                           |
|`spawning`|Instance coming online, not yet at capacity|
|`draining`|Instance being taken out of rotation       |

### Layer Toggles

Each node type can be shown or hidden independently. Toggling a layer also hides traffic edges where both endpoints are in the hidden layer.

-----

## Hooks

### `useTopology`

Manages the WebSocket connection to the Go backend and maintains live node and edge state. Handles reconnection on drop with exponential backoff. Returns:

```ts
{
  nodes: InfraNode[]
  edges: TrafficEdge[]
  connected: boolean
}
```

### `useMetrics`

Subscribes to per-node metric updates. Accepts a node ID and returns live p95, p99, memory, connection count, and `INSTANCE.WEIGHT` score.

### `useProvisioning`

Exposes provisioning and chaos control actions — spawn node, shard dummy Redis, replicate dummy DB, kill node, introduce latency. All actions are sent to the backend via authenticated internal endpoints and logged to the provisioning history.

-----

## Node Metrics

Each node exposes the following metrics via the backend WebSocket feed:

|Metric          |Type       |Description                                        |
|----------------|-----------|---------------------------------------------------|
|`weight`        |`float 0–1`|Composite load score — connections, memory, latency|
|`connections`   |`int`      |Current active connections                         |
|`maxConnections`|`int`      |Configured connection ceiling                      |
|`memory`        |`float 0–1`|Memory utilisation                                 |
|`p95`           |`ms`       |95th percentile request latency                    |
|`p99`           |`ms`       |99th percentile request latency                    |
|`status`        |`string`   |Current health state                               |
|`spawnedAt`     |`timestamp`|Instance start time                                |

-----

## Chaos Layer

The chaos layer is a separate topology environment running entirely against dummy infrastructure. Production is never a target.

Dummy environments are provisioned on demand:

- **Dummy Redis shard** — Redis instance pre-loaded with synthetic data mirroring production key structure
- **Dummy DB shard** — Postgres replica populated with synthetic data
- **Shadow prod** — full synthetic replica of the current production topology

Chaos controls available per node:

- Introduce latency — adds a configurable artificial delay to all commands on that node
- Kill node — sends SIGKILL, triggers failover and replica promotion
- Saturate connections — fills the connection pool to ceiling
- Invalidate cache — clears keys to trigger cache miss path

Red team mode locks the session to the chaos environment, records every action with a timestamp and actor, and generates a session report on completion.

-----

## Alerting

Alerts are trajectory-based. A warning fires when a node’s weight trend will breach the configured ceiling within the lookahead window — before degradation starts, not after.

Alert targets are configured per deployment:

- Native InfraMap alert feed
- Grafana
- AWS CloudWatch
- AWS SNS

Alert types:

|Type               |Trigger                                                |
|-------------------|-------------------------------------------------------|
|`weight_trajectory`|Node weight trending to ceiling within lookahead window|
|`weight_breach`    |Node weight exceeded configured threshold              |
|`node_critical`    |Node entered critical state                            |
|`spawn_complete`   |New instance came online                               |
|`chaos_detected`   |Anomaly detected in chaos layer session                |

-----

## Deployments

Active deployments are rendered as an overlay on the topology map. The deployment panel shows:

- Which nodes are receiving the deploy and in what order
- Current deploy status per node
- Rollback controls per deployment
- Deployment history — timestamped and correlated with the metric timeline

-----

## Backend

InfraMap connects to a Go backend service that aggregates infrastructure state and streams it over WebSocket. The backend is maintained in a separate repository.

The frontend expects the backend to expose:

- `WS /ws/topology` — live node and edge state stream
- `POST /provision/node` — spawn a new node
- `POST /provision/redis/shard` — create a dummy Redis shard
- `POST /provision/db/shard` — create a dummy DB shard
- `POST /chaos/latency` — introduce latency on a node
- `POST /chaos/kill` — kill a node
- `POST /chaos/saturate` — saturate connections on a node
- `POST /chaos/invalidate` — invalidate cache on a node

All endpoints require authentication via shared secret header.

-----

## Contributing

1. Fork the repository
1. Create a branch from `main` — `git checkout -b your-feature`
1. Make your changes
1. Open a pull request against `main` with a clear description of what changed and why

There are no contribution templates yet. Just be clear about what the change does.

-----

## License

MIT




## License

Apache 2.0
