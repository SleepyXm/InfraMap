"use client";

import type { SupabaseUsagePoint, WorkspaceMetrics, WorkspaceProviderStatus, WorkspaceResourceTelemetry } from "@/app/components/types/metrics";
import type { VercelDeployment } from "@/app/components/types/deployments";
import type { WorkspaceDetail } from "@/app/components/types/workspaces";
import styles from "./WorkspaceDashboard.module.css";

export type WorkspaceTab = "metrics" | "deployments" | "environments" | "services";

type WorkspaceDashboardProps = {
  detail: WorkspaceDetail;
  metrics: WorkspaceMetrics | null;
  loading: boolean;
  error: string | null;
  tab: WorkspaceTab;
  onRefresh: () => void;
};

export function WorkspaceDashboard({ detail, metrics, loading, error, tab, onRefresh }: WorkspaceDashboardProps) {
  if (tab === "deployments") return <DeploymentsView deployments={metrics?.deployments ?? []} loading={loading} error={error} onRefresh={onRefresh} />;
  if (tab === "environments") return <EnvironmentsView metrics={metrics} services={detail.services} />;
  if (tab === "services") return <ResourcesView resources={metrics?.resources ?? []} loading={loading} error={error} onRefresh={onRefresh} />;
  return <ObservabilityView detail={detail} metrics={metrics} loading={loading} error={error} onRefresh={onRefresh} />;
}

function ResourcesView({ resources, loading, error, onRefresh }: { resources: WorkspaceResourceTelemetry[]; loading: boolean; error: string | null; onRefresh: () => void; }) {
  return (
    <section className={styles.resourceTelemetry}>
      <div className={styles.viewHeader}>
        <div><p>Resource observability</p><h2>Analytics, attributed.</h2><span>Every signal below belongs to the resource ID shown with it.</span></div>
        <button type="button" onClick={onRefresh} disabled={loading}>{loading ? "Refreshing…" : "Refresh analytics"}</button>
      </div>
      {error && <div className={styles.warnings}><strong>Resource telemetry unavailable</strong><p>{error}</p></div>}
      <div className={styles.resourceTelemetryGrid}>
        {resources.length ? resources.map((resource) => <ResourceTelemetryCard key={resource.resource_id} resource={resource} />) : <EmptyState title="No resources attached" detail="Attach a source or provider resource and its attributed signals will appear here." />}
      </div>
    </section>
  );
}

function ResourceTelemetryCard({ resource }: { resource: WorkspaceResourceTelemetry; }) {
  const requests = totalRequests(resource.supabase_usage);
  const healthy = resource.supabase_services.filter((service) => service.healthy).length;
  return (
    <article className={styles.resourceTelemetryCard}>
      <header><span>{resource.provider}</span><strong>{resource.name}</strong><em>{statusLabel(resource.status)}</em></header>
      <div className={styles.resourceTelemetryMetrics}>
        <Metric label="Requests" value={formatNumber(requests)} detail={resource.supabase_usage.length ? "Observed activity" : "No request telemetry"} />
        <Metric label="Deployments" value={String(resource.deployments.length)} detail={resource.deployments[0] ? formatDeploymentState(resource.deployments[0]) : "No deployment telemetry"} />
        <Metric label="Services" value={String(resource.service_ids.length)} detail={resource.service_ids.length ? "Mapped by EZDeploy" : "No service mapping"} />
        <Metric label="Health" value={resource.supabase_services.length ? `${healthy}/${resource.supabase_services.length}` : "—"} detail={resource.supabase_services.length ? "Provider services healthy" : "No health telemetry"} />
      </div>
      <footer><span>{resource.resource_type}</span><code>{resource.resource_id}</code></footer>
      {resource.warnings.map((warning) => <p className={styles.resourceWarning} key={warning}>{warning}</p>)}
    </article>
  );
}

function ObservabilityView({ detail, metrics, loading, error, onRefresh }: Omit<WorkspaceDashboardProps, "tab">) {
  const providers = metrics?.providers ?? fallbackProviders(detail);
  const connected = providers.filter((provider) => provider.connected).length;
  const requestTotal = totalRequests(metrics?.supabase_usage ?? []);
  const latestDeployment = metrics?.deployments[0];
  const buildTime = deploymentDuration(latestDeployment);
  const serviceHealth = metrics?.supabase_services ?? [];
  const healthyServices = serviceHealth.filter((service) => service.healthy).length;
  const attention = metrics?.status === "attention";
  const partial = metrics?.status === "partial";

  return (
    <div className={styles.dashboard}>
      <section className={attention || partial ? styles.healthAttention : styles.healthOperational}>
        <span className={styles.healthDot} aria-hidden="true" />
        <div>
          <strong>{attention ? "System needs attention" : partial ? "Telemetry partially available" : "Connected systems operational"}</strong>
          <span>{attention ? "One or more attached providers need action." : partial ? metrics?.warnings[0] ?? "Attached systems are reachable, but some live signals could not be read." : `Last checked ${formatObservedAt(metrics?.observed_at)}`}</span>
        </div>
        <button type="button" onClick={onRefresh} disabled={loading}>{loading ? "Checking…" : "Refresh status"}</button>
      </section>

      {error ? (
        <section className={styles.warnings} aria-label="Telemetry notices">
          <strong>Telemetry notice</strong>
          <div><p>{error}</p></div>
        </section>
      ) : null}

      <section className={styles.metrics} aria-label="Workspace metrics">
        <Metric label="API requests observed" value={formatNumber(requestTotal)} detail={requestTotal ? "Supabase Analytics logs" : "No activity in the last 24 hours"} />
        <Metric label="Recent deployments" value={String(metrics?.deployments.length ?? 0)} detail={latestDeployment ? formatDeploymentState(latestDeployment) : "No deployment history"} />
        <Metric label="Latest build time" value={buildTime} detail={latestDeployment?.target ? `${capitalize(latestDeployment.target)} deployment` : "No completed build"} />
        <Metric label="System coverage" value={`${connected}/4`} detail={`${healthyServices}/${serviceHealth.length || 5} Supabase services healthy`} />
      </section>

      <div className={styles.primaryGrid}>
        <section className={styles.systemPanel}>
          <PanelHeader eyebrow="Live topology" title="System map" detail="The attached path from source to runtime." />
          <div className={styles.systemFlow}>
            {providers.map((provider, index) => (
              <div className={styles.flowItem} key={provider.provider}>
                <ProviderNode provider={provider} />
                {index < providers.length - 1 && <span className={styles.flowConnector} aria-hidden="true" />}
              </div>
            ))}
          </div>
          <div className={styles.systemFoot}>
            <span>Application path</span>
            <strong>{connected === 4 ? "Fully attached" : `${4 - connected} layer${4 - connected === 1 ? "" : "s"} remaining`}</strong>
          </div>
        </section>

        <section className={styles.providerPanel}>
          <PanelHeader eyebrow="Provider state" title="Health" detail="Live provider and resource condition." />
          <div className={styles.providerList}>{providers.map((provider) => <ProviderHealth key={provider.provider} provider={provider} />)}</div>
        </section>
      </div>

      <div className={styles.secondaryGrid}>
        <section className={styles.chartPanel}>
          <PanelHeader eyebrow="Supabase" title="Request activity" detail="OAuth Analytics logs across the last 24 hours." />
          <UsageChart points={metrics?.supabase_usage ?? []} />
        </section>
        <section className={styles.servicesPanel}>
          <PanelHeader eyebrow="Database platform" title="Supabase services" detail="Health reported by the linked project." />
          {serviceHealth.length ? (
            <div className={styles.serviceList}>{serviceHealth.map((service) => (
              <div key={service.name}><span className={service.healthy ? styles.dotHealthy : styles.dotAttention} /><strong>{formatServiceName(service.name)}</strong><em>{service.healthy ? "Healthy" : service.status || "Unavailable"}</em></div>
            ))}</div>
          ) : <EmptyState title="Service health unavailable" detail="Refresh the control plane to retry the linked Supabase project." />}
        </section>
      </div>

      <section className={styles.activityPanel}>
        <PanelHeader eyebrow="Operational history" title="Recent activity" detail="Latest deployment and provider events." />
        <ActivityList deployments={metrics?.deployments ?? []} providers={providers} />
      </section>
    </div>
  );
}

function DeploymentsView({ deployments, loading, error, onRefresh }: { deployments: VercelDeployment[]; loading: boolean; error: string | null; onRefresh: () => void; }) {
  return (
    <div className={styles.dashboard}>
      <section className={styles.viewHeader}>
        <div><p>Vercel delivery</p><h2>Deployments</h2><span>Build state, target, duration, and destination for the attached Vercel project.</span></div>
        <button type="button" onClick={onRefresh} disabled={loading}>{loading ? "Refreshing…" : "Refresh deployments"}</button>
      </section>
      {error && <section className={styles.warnings}><strong>Deployment telemetry unavailable</strong><p>{error}</p></section>}
      <section className={styles.tablePanel}>
        <div className={styles.tableHeader}><span>Deployment</span><span>Environment</span><span>State</span><span>Duration</span><span>Created</span></div>
        {deployments.length ? deployments.map((deployment) => (
          <a className={styles.tableRow} key={deployment.uid} href={deployment.inspectorUrl || `https://${deployment.url}`} target="_blank" rel="noreferrer">
            <span><strong>{deployment.name}</strong><small>{deployment.url}</small></span>
            <span>{capitalize(deployment.target || "preview")}</span>
            <span><i className={isHealthyDeployment(deployment) ? styles.dotHealthy : styles.dotAttention} />{formatDeploymentState(deployment)}</span>
            <span>{deploymentDuration(deployment)}</span>
            <span>{formatDate(deployment.created)}</span>
          </a>
        )) : <EmptyState title="No deployments reported" detail="Deploy the attached Vercel project and its build history will appear here." />}
      </section>
    </div>
  );
}

function EnvironmentsView({ metrics, services }: { metrics: WorkspaceMetrics | null; services: WorkspaceDetail["services"]; }) {
  const providers = metrics?.providers ?? [];
  return (
    <div className={styles.dashboard}>
      <section className={styles.viewHeader}><div><p>Environment control</p><h2>Environments</h2><span>Provider coverage for production, preview, and development.</span></div></section>
      <section className={styles.environmentPanel}>
        <div className={styles.environmentRow}><span><i className={styles.environmentProduction} />Production</span><strong>{providers.filter((provider) => provider.connected).length} providers attached</strong><em>Provider mappings not configured</em></div>
        <div className={styles.environmentRow}><span><i />Preview</span><strong>{services.some((service) => service.provider === "vercel") ? "Vercel available" : "No deployment service"}</strong><em>Variable sync not configured</em></div>
        <div className={styles.environmentRow}><span><i />Development</span><strong>Local workflow</strong><em>Remote environment not configured</em></div>
      </section>
      <section className={styles.emptyConfiguration}>
        <p>Next control layer</p><h3>Environment mappings are not configured yet.</h3><span>This view is intentionally showing the real state. Variable storage, provider-specific targets, and controlled sync can be added once the operational boundaries are defined.</span>
      </section>
    </div>
  );
}

function Metric({ label, value, detail }: { label: string; value: string; detail: string; }) {
  return <article className={styles.metric}><span>{label}</span><strong>{value}</strong><small>{detail}</small></article>;
}

function PanelHeader({ eyebrow, title, detail }: { eyebrow: string; title: string; detail: string; }) {
  return <header className={styles.panelHeader}><div><p>{eyebrow}</p><h2>{title}</h2></div><span>{detail}</span></header>;
}

function ProviderNode({ provider }: { provider: WorkspaceProviderStatus; }) {
  return (
    <article className={provider.connected ? styles.providerNode : styles.providerNodeEmpty}>
      <span className={styles.providerInitial}>{provider.name.slice(0, 1).toUpperCase()}</span>
      <div><small>{providerLabel(provider.provider)}</small><strong>{provider.connected ? provider.name : "Not attached"}</strong><em>{statusLabel(provider.status)}</em></div>
    </article>
  );
}

function ProviderHealth({ provider }: { provider: WorkspaceProviderStatus; }) {
  const needsAttention = provider.status === "paused" || provider.status === "error";
  const href = provider.provider === "supabase" && provider.connected ? "https://supabase.com/dashboard/projects" : undefined;
  const content = <><span className={provider.connected && !needsAttention ? styles.dotHealthy : needsAttention ? styles.dotAttention : styles.dotMuted} /><div><strong>{providerLabel(provider.provider)}</strong><small>{provider.detail}</small></div><em>{statusLabel(provider.status)}</em></>;
  return href ? <a className={styles.providerHealth} href={href} target="_blank" rel="noreferrer">{content}</a> : <div className={styles.providerHealth}>{content}</div>;
}

function UsageChart({ points }: { points: SupabaseUsagePoint[]; }) {
  const values = points.slice(-18).map(pointRequests);
  const maximum = Math.max(...values, 1);
  if (!values.length) return <EmptyState title="No request activity in this window" detail="Analytics access is connected; new Supabase requests will appear here." />;
  return (
    <div className={styles.chart} aria-label="Supabase request activity">
      <div className={styles.chartScale}><span>{formatNumber(maximum)}</span><span>{formatNumber(Math.round(maximum / 2))}</span><span>0</span></div>
      <div className={styles.chartBars}>{values.map((value, index) => <span key={`${value}-${index}`} style={{ height: `${Math.max(3, (value / maximum) * 100)}%` }} title={`${formatNumber(value)} requests`} />)}</div>
      <div className={styles.chartAxis}><span>{formatUsageTime(points.at(-values.length)?.timestamp)}</span><span>Latest</span></div>
    </div>
  );
}

function ActivityList({ deployments, providers }: { deployments: VercelDeployment[]; providers: WorkspaceProviderStatus[]; }) {
  const rows = deployments.slice(0, 4).map((deployment) => ({
    key: deployment.uid,
    healthy: isHealthyDeployment(deployment),
    title: `${deployment.name} deployment ${formatDeploymentState(deployment).toLowerCase()}`,
    context: capitalize(deployment.target || "preview"),
    time: formatDate(deployment.created),
  }));
  if (!rows.length) providers.filter((provider) => provider.connected).forEach((provider) => rows.push({ key: provider.provider, healthy: provider.status !== "paused" && provider.status !== "error", title: `${providerLabel(provider.provider)} resource attached`, context: statusLabel(provider.status), time: provider.detail }));
  if (!rows.length) return <EmptyState title="No operational events yet" detail="Attach a provider to begin building the workspace activity history." />;
  return <div className={styles.activityList}>{rows.map((row) => <div key={row.key}><span className={row.healthy ? styles.dotHealthy : styles.dotAttention} /><strong>{row.title}</strong><em>{row.context}</em><time>{row.time}</time></div>)}</div>;
}

function EmptyState({ title, detail }: { title: string; detail: string; }) {
  return <div className={styles.emptyState}><strong>{title}</strong><span>{detail}</span></div>;
}

function fallbackProviders(detail: WorkspaceDetail): WorkspaceProviderStatus[] {
  const resources = [...(detail.sources ?? []), ...detail.services];
  return (["github", "vercel", "supabase", "aws"] as const).map((provider) => {
    const resource = resources.find((item) => item.provider === provider);
    return { provider, name: resource?.name ?? providerLabel(provider), connected: Boolean(resource), status: resource?.status ?? "not_connected", detail: resource?.resource_type ?? "No resource attached" };
  });
}

function pointRequests(point: SupabaseUsagePoint): number {
  return point.total_auth_requests + point.total_realtime_requests + point.total_rest_requests + point.total_storage_requests;
}

function totalRequests(points: SupabaseUsagePoint[]): number {
  return points.reduce((total, point) => total + pointRequests(point), 0);
}

function deploymentDuration(deployment?: VercelDeployment): string {
  if (!deployment?.buildingAt || !deployment.ready || deployment.ready <= deployment.buildingAt) return "—";
  const seconds = Math.round((deployment.ready - deployment.buildingAt) / 1000);
  if (seconds < 60) return `${seconds}s`;
  return `${Math.floor(seconds / 60)}m ${String(seconds % 60).padStart(2, "0")}s`;
}

function formatDeploymentState(deployment: VercelDeployment): string {
  return statusLabel(deployment.readyState || deployment.state || "unknown");
}

function isHealthyDeployment(deployment: VercelDeployment): boolean {
  return ["ready", "succeeded"].includes((deployment.readyState || deployment.state).toLowerCase());
}

function statusLabel(value: string): string {
  if (value === "not_connected") return "Not connected";
  if (value === "paused" || value.toLowerCase() === "inactive") return "Paused";
  return value.toLowerCase().split("_").map(capitalize).join(" ");
}

function providerLabel(provider: WorkspaceProviderStatus["provider"]): string {
  if (provider === "github") return "GitHub";
  if (provider === "vercel") return "Vercel";
  if (provider === "supabase") return "Supabase";
  return "AWS";
}

function formatServiceName(name: string): string {
  if (/^gotrue$/i.test(name)) return "Auth";
  if (/^(postgrest|rest)$/i.test(name)) return "REST API";
  if (/^(postgres|db)$/i.test(name)) return "Database";
  return name.split(/[-_]/).map(capitalize).join(" ");
}

function formatObservedAt(value?: string): string {
  if (!value) return "when refreshed";
  return new Intl.DateTimeFormat(undefined, { hour: "2-digit", minute: "2-digit" }).format(new Date(value));
}

function formatUsageTime(value?: string): string {
  if (!value) return "Earlier";
  return new Intl.DateTimeFormat(undefined, { hour: "2-digit", minute: "2-digit" }).format(new Date(value));
}

function formatDate(value: number): string {
  if (!value) return "Not reported";
  return new Intl.DateTimeFormat(undefined, { month: "short", day: "numeric", hour: "2-digit", minute: "2-digit" }).format(new Date(value));
}

function formatNumber(value: number): string {
  return new Intl.NumberFormat(undefined, { notation: value >= 1000 ? "compact" : "standard", maximumFractionDigits: 1 }).format(value);
}

function capitalize(value: string): string {
  return value ? value.charAt(0).toUpperCase() + value.slice(1).toLowerCase() : value;
}
