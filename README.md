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

## License

Apache 2.0
