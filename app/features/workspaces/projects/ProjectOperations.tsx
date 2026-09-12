"use client";

import Link from "next/link";
import { useState } from "react";
import { useWorkspaceConnections } from "@/app/features/workspaces/connections";
import { ProjectControlDashboard, type ProjectControlTab } from "./control/ProjectControlDashboard";
import { ProviderResourceStage } from "./provider/ProviderResourceStage";
import type { ProjectResource } from "./types";
import { useProjectControl } from "./useProjectControl";
import { useWorkspaceProject } from "./useWorkspaceProject";
import { SupabaseProjectStage } from "./supabase/SupabaseProjectStage";
import styles from "./ProjectOperations.module.css";

export function ProjectOperations({ accountID, projectID }: { accountID: string; projectID: string }) {
  const [tab, setTab] = useState<ProjectControlTab>("observability");
  const { detail, loading, error, refresh, setResource } = useWorkspaceProject(accountID, projectID);
  const { control, loading: controlLoading, error: controlError, refresh: refreshControl } = useProjectControl(accountID, projectID);
  const { connections, loading: connectionsLoading, error: connectionsError } = useWorkspaceConnections(accountID);

  if (loading) return <main className={styles.state}>Loading the project control plane…</main>;
  if (error || !detail) {
    return (
      <main className={styles.state}>
        <div><span className={styles.stateMark}>!</span><h1>This project could not be opened.</h1><p>{error}</p><button type="button" onClick={() => void refresh()}>Try again</button></div>
      </main>
    );
  }

  const storedSupabase = detail.resources.find((resource) => resource.provider === "supabase" && resource.resource_type === "project");
  const liveSupabase = control?.providers.find((provider) => provider.provider === "supabase");
  const supabase = storedSupabase && liveSupabase ? { ...storedSupabase, status: liveSupabase.status } : storedSupabase;
  const source = detail.resources.find((resource) => resource.provider === "github");
  const deployment = detail.resources.find((resource) => resource.provider === "vercel");
  const runtime = detail.resources.find((resource) => resource.provider === "aws");

  function handleAttached(resource: ProjectResource) {
    setResource(resource);
    void refreshControl();
  }

  return (
    <main className={styles.page}>
      <div className={styles.shell}>
        <header className={styles.controlHeader}>
          <div>
            <Link href="/workspaces?tab=workspaces" className={styles.backLink}>← Workspaces</Link>
            <p className={styles.eyebrow}>Project control</p>
            <h1>{detail.project.name}</h1>
            <p>{detail.project.description || "A live view of this project's source, delivery, data, and runtime."}</p>
          </div>
          <div className={styles.controlMeta}>
            <div className={styles.headerMeta}><span className={control?.status === "operational" ? styles.operational : styles.attention}>{control?.status ?? detail.project.status}</span><span>{detail.role}</span></div>
            <span>Project ID</span><strong>{detail.project.slug}</strong><small>{control ? `Observed ${formatObservation(control.observed_at)}` : "Loading provider state"}</small>
          </div>
        </header>

        <nav className={styles.tabs} aria-label="Project controls">
          {(["observability", "deployments", "environments", "resources"] as const).map((item) => <button key={item} type="button" className={tab === item ? styles.tabActive : undefined} aria-current={tab === item ? "page" : undefined} onClick={() => setTab(item)}>{item}</button>)}
        </nav>

        {tab === "resources" ? (
          <div className={styles.resourcesView}>
            <section className={styles.stages} aria-label="Project layers">
              <Stage number="01" label="Source" name={source?.name} action="Connect GitHub" href="#source-layer" />
              <Stage number="02" label="Deployment" name={deployment?.name} action="Connect Vercel" href="#deployment-layer" />
              <Stage number="03" label="Data" name={supabase?.name} action="Configure Supabase" href="#data-layer" />
              <Stage number="04" label="Runtime" name={runtime?.name} action="Connect AWS" />
            </section>

            <ProviderResourceStage
              accountID={accountID}
              projectID={projectID}
              provider="github"
              role={detail.role}
              resource={source}
              connections={connections.filter((connection) => connection.provider === "github")}
              connectionsLoading={connectionsLoading}
              connectionsError={connectionsError}
              onAttached={handleAttached}
            />
            <ProviderResourceStage
              accountID={accountID}
              projectID={projectID}
              provider="vercel"
              role={detail.role}
              resource={deployment}
              connections={connections.filter((connection) => connection.provider === "vercel")}
              connectionsLoading={connectionsLoading}
              connectionsError={connectionsError}
              onAttached={handleAttached}
            />
            <SupabaseProjectStage
              accountID={accountID}
              projectID={projectID}
              role={detail.role}
              resource={supabase}
              connections={connections.filter((connection) => connection.provider === "supabase")}
              connectionsLoading={connectionsLoading}
              connectionsError={connectionsError}
              onAttached={handleAttached}
            />
          </div>
        ) : (
          <ProjectControlDashboard detail={detail} control={control} loading={controlLoading} error={controlError} tab={tab} onRefresh={() => void refreshControl()} />
        )}
      </div>
    </main>
  );
}

function Stage({ number, label, name, action, href = "/workspaces?tab=connections" }: { number: string; label: string; name?: string; action: string; href?: string }) {
  return (
    <article className={styles.stage}>
      <div><span className={styles.stageNumber}>{number}</span><span className={name ? styles.stageLive : styles.stageEmpty}>{name ? "Attached" : "Open"}</span></div>
      <div><small>{label}</small><h3>{name || "Not connected"}</h3></div>
      <Link href={href}>{name ? "View layer" : action} <span>→</span></Link>
    </article>
  );
}

function formatObservation(value: string): string {
  return new Intl.DateTimeFormat(undefined, { hour: "2-digit", minute: "2-digit" }).format(new Date(value));
}
