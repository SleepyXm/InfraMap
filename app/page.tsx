import { FeatureCard, LinkButton, SignalPreview } from "@/app/UI";
import styles from "@/app/UI/Landing.module.css";

const capabilities = [
  { id: "01", title: "Infrastructure topology map", detail: "A geographic world map with live nodes, traffic edges, layer controls, spawning instances, and critical-state pulses.", state: "Live", featured: true },
  { id: "02", title: "Node health & metrics", detail: "INSTANCE.WEIGHT, p95 and p99 latency, connections, memory utilisation, and uptime at each node.", state: "Live" },
  { id: "03", title: "Predictive alerting", detail: "Find infrastructure trending toward a ceiling before a threshold becomes an incident.", state: "Watching" },
  { id: "04", title: "Provisioning & control", detail: "Spawn nodes, shard Redis, replicate database shards, and shadow production from the topology.", state: "Ready" },
  { id: "05", title: "Chaos testing", detail: "Introduce latency, kill nodes, saturate connections, or invalidate cache in a dedicated chaos layer.", state: "Ready" },
  { id: "06", title: "Deployment visibility", detail: "Track active deployments, ordering, rollback controls, and the history behind a latency spike.", state: "Watching" },
] as const;

export default function HomePage() {
  return (
    <main className={styles.page}>
      <section className={`${styles.shell} ${styles.hero}`}>
        <div>
          <p className={styles.eyebrow}>Infrastructure, made legible</p>
          <h1>Move from cloud complexity to clear control.</h1>
          <p className={styles.heroText}>InfraMap gives you a live geographic view of your infrastructure — nodes, traffic, health state, and latency — with the controls to provision, test, and manage it from a single place.</p>
          <div className={styles.heroActions}><LinkButton href="/map">Open the live map <span aria-hidden="true">→</span></LinkButton><LinkButton href="#capabilities" variant="outline">Explore the platform</LinkButton></div>
          <p className={styles.heroNote}>The interface is open. Plugins connect it to your infrastructure.</p>
        </div>
        <SignalPreview><strong>Live topology</strong><br />12 nodes · 21 routes · 1 alert</SignalPreview>
      </section>

      <div className={`${styles.shell} ${styles.logoStrip}`}><span>Built to understand the systems behind</span><span>Supabase</span><span>Postgres</span><span>Redis</span><span>MySQL</span><span>MongoDB</span></div>

      <section id="capabilities" className={`${styles.shell} ${styles.capabilities}`}>
        <div className={styles.sectionHeading}><h2>Everything happens from the map.</h2><p>Infrastructure state comes through the Go backend in real time. Plugins discover, inspect, and report health; the interface turns that depth into something you can act on.</p></div>
        <div className={styles.featureGrid}>{capabilities.map(capability => <FeatureCard key={capability.id} {...capability} />)}</div>
      </section>

      <section className={`${styles.shell} ${styles.closing}`}>
        <div><h2>Start with the system you have. Understand what it needs next.</h2><p>From a single node under pressure to a deployment changing the shape of your topology, InfraMap keeps the important context in one operational surface.</p></div>
        <div className={styles.closingAside}><span>Control plane</span><strong>Observe. Decide. Act.</strong><p>Provisioning and chaos controls remain authenticated through the backend — never in the interface itself.</p><LinkButton href="/map" variant="brand">View infrastructure <span aria-hidden="true">→</span></LinkButton></div>
      </section>
    </main>
  );
}
