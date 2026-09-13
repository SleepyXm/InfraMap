"use client";

import Link from "next/link";
import { useState } from "react";
import { useWorkspaceConnections } from "@/app/features/workspaces/connections";
import { WorkspaceDashboard, type WorkspaceTab } from "../observability/WorkspaceDashboard";
import { WorkspaceSetupGuide, type WorkspaceLayer } from "./WorkspaceSetupGuide";
import { EZDeployAnalysisStage } from "../scans/EZDeployAnalysisStage";
import { ProviderAttachmentStage } from "./ProviderAttachmentStage";
import type { WorkspaceResource } from "@/app/components/types/workspaces";
import { useWorkspaceMetrics } from "./useWorkspaceMetrics";
import { useWorkspace } from "./useWorkspace";
import { SupabaseServiceStage } from "../services/SupabaseServiceStage";
import { AWSRuntimeStage } from "../services/AWSRuntimeStage";
import styles from "./WorkspaceControl.module.css";

export function WorkspaceControl({ accountID, workspaceID, initialSetup = false }: { accountID: string; workspaceID: string; initialSetup?: boolean; }) {
  const [tab, setTab] = useState<WorkspaceTab>(initialSetup ? "services" : "metrics");
  const [resourceLayer, setResourceLayer] = useState<WorkspaceLayer>("source");
  const { detail, loading, error, refresh, setResource } = useWorkspace(accountID, workspaceID);
  const { metrics, loading: metricsLoading, error: metricsError, refresh: refreshMetrics } = useWorkspaceMetrics(accountID, workspaceID);
  const { connections, loading: connectionsLoading, error: connectionsError } = useWorkspaceConnections(accountID);

  if (loading) return <main className={styles.state}>Loading the workspace control plane…</main>;
  if (error || !detail) {
    return (
      <main className={styles.state}>
        <div><span className={styles.stateMark}>!</span><h1>This workspace could not be opened.</h1><p>{error}</p><button type="button" onClick={() => void refresh()}>Try again</button></div>
      </main>
    );
  }

  const sources = detail.sources;
  const attachedResources = [...sources, ...detail.services];
  const storedSupabase = detail.services.find((service) => service.provider === "supabase");
  const liveSupabase = metrics?.providers.find((provider) => provider.provider === "supabase");
  const supabase = storedSupabase && liveSupabase ? { ...storedSupabase, status: liveSupabase.status } : storedSupabase;
  const source = sources[0];
  const deployment = detail.services.find((service) => service.provider === "vercel");
  const runtime = detail.services.find((service) => service.provider === "aws");

  function handleAttached(resource: WorkspaceResource) {
    setResource(resource);
    void refreshMetrics();
    if (resource.provider === "github") setResourceLayer("analysis");
    if (resource.provider === "vercel") setResourceLayer("data");
    if (resource.provider === "supabase") setResourceLayer("runtime");
  }

  function handleSourceUpdated(resource: WorkspaceResource) {
    setResource(resource);
    void refreshMetrics();
  }

  return (
    <main className={styles.page}>
      <div className={styles.shell}>
        <header className={styles.controlHeader}>
          <div>
            <Link href="/workspaces?tab=workspaces" className={styles.backLink}>← Workspaces</Link>
            <p className={styles.eyebrow}>Workspace control</p>
            <h1>{detail.workspace.name}</h1>
            <p>{detail.workspace.description || "A live view of this workspace's sources, services, deployments, and metrics."}</p>
          </div>
          <div className={styles.controlMeta}>
            <div className={styles.headerMeta}><span className={metrics?.status === "operational" ? styles.operational : styles.attention}>{metrics?.status ?? detail.workspace.status}</span><span>{detail.access}</span></div>
            <span>Workspace ID</span><strong>{detail.workspace.slug}</strong><small>{metrics ? `Observed ${formatObservation(metrics.observed_at)}` : "Loading provider state"}</small>
          </div>
        </header>

        <nav className={styles.tabs} aria-label="Workspace controls">
          {(["metrics", "deployments", "environments", "services"] as const).map((item) => <button key={item} type="button" className={tab === item ? styles.tabActive : undefined} aria-current={tab === item ? "page" : undefined} onClick={() => setTab(item)}>{item}</button>)}
        </nav>

        {tab === "services" ? (
          <div className={styles.resourcesView}>
            <WorkspaceSetupGuide workspaceName={detail.workspace.name} resources={attachedResources} activeLayer={resourceLayer} onSelect={setResourceLayer} />
            {resourceLayer === "source" && <ProviderAttachmentStage
              accountID={accountID}
              workspaceID={workspaceID}
              provider="github"
              role={detail.access}
              resource={source}
              connections={connections.filter((connection) => connection.provider === "github")}
              connectionsLoading={connectionsLoading}
              connectionsError={connectionsError}
              onAttached={handleAttached}
            />}
            {resourceLayer === "analysis" && source && <EZDeployAnalysisStage accountID={accountID} workspaceID={workspaceID} role={detail.access} source={source} resources={attachedResources} onUpdated={handleSourceUpdated} onNavigate={setResourceLayer} />}
            {resourceLayer === "deployment" && <ProviderAttachmentStage
              accountID={accountID}
              workspaceID={workspaceID}
              provider="vercel"
              role={detail.access}
              resource={deployment}
              connections={connections.filter((connection) => connection.provider === "vercel")}
              connectionsLoading={connectionsLoading}
              connectionsError={connectionsError}
              onAttached={handleAttached}
            />}
            {resourceLayer === "data" && <SupabaseServiceStage
              accountID={accountID}
              workspaceID={workspaceID}
              role={detail.access}
              resource={supabase}
              connections={connections.filter((connection) => connection.provider === "supabase")}
              connectionsLoading={connectionsLoading}
              connectionsError={connectionsError}
              onAttached={handleAttached}
            />}
            {resourceLayer === "runtime" && <AWSRuntimeStage accountID={accountID} workspaceID={workspaceID} role={detail.access} resource={runtime} source={source} onAttached={handleAttached} onNavigate={setResourceLayer} />}
            <WorkspaceDashboard detail={detail} metrics={metrics} loading={metricsLoading} error={metricsError} tab="services" onRefresh={() => void refreshMetrics()} />
          </div>
        ) : (
          <WorkspaceDashboard detail={detail} metrics={metrics} loading={metricsLoading} error={metricsError} tab={tab} onRefresh={() => void refreshMetrics()} />
        )}
      </div>
    </main>
  );
}

function formatObservation(value: string): string {
  return new Intl.DateTimeFormat(undefined, { hour: "2-digit", minute: "2-digit" }).format(new Date(value));
}
