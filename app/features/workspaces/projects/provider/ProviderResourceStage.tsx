"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { getGitHubRepositories, getVercelProjects } from "@/app/features/workspaces/connections";
import type { Connection } from "@/app/components/types/users";
import { attachGitHubResource, attachVercelResource } from "../api";
import type { ProjectResource, WorkspaceProjectDetail } from "../types";
import styles from "./ProviderResourceStage.module.css";

type Provider = "github" | "vercel";
type ProviderOption = { id: string; name: string; detail: string; status: string };

type ProviderResourceStageProps = {
  accountID: string;
  projectID: string;
  provider: Provider;
  role: WorkspaceProjectDetail["role"];
  resource?: ProjectResource;
  connections: Connection[];
  connectionsLoading: boolean;
  connectionsError: string | null;
  onAttached: (resource: ProjectResource) => void;
};

const providerCopy = {
  github: {
    sectionID: "source-layer",
    eyebrow: "Source layer",
    title: "Connect GitHub.",
    description: "Choose the repository that supplies this project. InfraMap verifies access again before binding it.",
    mark: "GH",
    resourceLabel: "repository",
  },
  vercel: {
    sectionID: "deployment-layer",
    eyebrow: "Deployment layer",
    title: "Connect Vercel.",
    description: "Bind the Vercel project that builds and serves this application so deployments can be managed from one place.",
    mark: "▲",
    resourceLabel: "project",
  },
} as const;

export function ProviderResourceStage({ accountID, projectID, provider, role, resource, connections, connectionsLoading, connectionsError, onAttached }: ProviderResourceStageProps) {
  const copy = providerCopy[provider];
  const [connectionID, setConnectionID] = useState("");
  const [options, setOptions] = useState<ProviderOption[]>([]);
  const [selectedID, setSelectedID] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const selectedConnection = connections.find((connection) => connection.id === connectionID) ?? connections[0];
  const canAttach = role !== "viewer";

  useEffect(() => {
    if (!selectedConnection || resource) return;
    let active = true;
    void Promise.resolve().then(async () => {
      if (!active) return;
      setLoading(true);
      setError("");
      setOptions([]);
      setSelectedID("");
      try {
        const available = await loadProviderOptions(provider, accountID, selectedConnection.id);
        if (active) {
          setOptions(available);
          setSelectedID(available[0]?.id ?? "");
        }
      } catch (cause) {
        if (active) setError(cause instanceof Error ? cause.message : `${copy.title.replace("Connect ", "")} resources could not be loaded.`);
      } finally {
        if (active) setLoading(false);
      }
    });

    return () => { active = false; };
  }, [accountID, copy.title, provider, resource, selectedConnection]);

  async function attachResource() {
    if (!selectedConnection || !selectedID || !canAttach) return;
    setSaving(true);
    setError("");
    try {
      const attached = provider === "github"
        ? await attachGitHubResource(accountID, projectID, selectedConnection.id, Number(selectedID))
        : await attachVercelResource(accountID, projectID, selectedConnection.id, selectedID);
      onAttached(attached);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : `${copy.resourceLabel} could not be attached.`);
    } finally {
      setSaving(false);
    }
  }

  if (resource) return <ConnectedProviderResource provider={provider} resource={resource} />;

  return (
    <section id={copy.sectionID} className={styles.section}>
      <header className={styles.heading}>
        <div><p className={styles.eyebrow}>{copy.eyebrow}</p><h2>{copy.title}</h2><p>{copy.description}</p></div>
        <span className={styles.providerMark}>{copy.mark}</span>
      </header>

      {connectionsError && <p className={styles.error}>{connectionsError}</p>}
      {connectionsLoading ? (
        <p className={styles.state}>Loading {providerLabel(provider)} connections…</p>
      ) : connections.length === 0 ? (
        <div className={styles.emptyConnection}>
          <div><h3>Connect {providerLabel(provider)} first.</h3><p>The workspace needs an authorised {providerLabel(provider)} account before a {copy.resourceLabel} can be attached.</p></div>
          <Link href="/workspaces?tab=connections">Open connections →</Link>
        </div>
      ) : (
        <div className={styles.setup}>
          {connections.length > 1 && (
            <label className={styles.field}><span>{providerLabel(provider)} connection</span><select value={selectedConnection?.id ?? ""} onChange={(event) => setConnectionID(event.target.value)}>{connections.map((connection) => <option key={connection.id} value={connection.id}>{connection.name}</option>)}</select></label>
          )}
          <div className={styles.attachPanel}>
            <div><h3>Choose a {copy.resourceLabel}</h3><p>Only resources currently accessible through this workspace connection are shown.</p></div>
            {loading ? <p className={styles.state}>Loading available {copy.resourceLabel}s…</p> : options.length ? (
              <>
                <div className={styles.resourceChoices}>
                  {options.map((option) => (
                    <label key={option.id} className={selectedID === option.id ? styles.resourceChoiceSelected : styles.resourceChoice}>
                      <input type="radio" name={`${provider}-resource`} value={option.id} checked={selectedID === option.id} onChange={() => setSelectedID(option.id)} />
                      <span><strong>{option.name}</strong><small>{option.detail}</small></span>
                      <em>{option.status}</em>
                    </label>
                  ))}
                </div>
                <button type="button" className={styles.primaryAction} disabled={saving || !selectedID || !canAttach} onClick={() => void attachResource()}>{saving ? "Attaching…" : `Attach selected ${copy.resourceLabel}`}</button>
                {!canAttach && <p className={styles.permissionNote}>Viewers can inspect this layer but cannot change its attachment.</p>}
              </>
            ) : <p className={styles.state}>No {providerLabel(provider)} {copy.resourceLabel}s are available on this connection.</p>}
          </div>
          {error && <p className={styles.error}>{error}</p>}
        </div>
      )}
    </section>
  );
}

async function loadProviderOptions(provider: Provider, accountID: string, connectionID: string): Promise<ProviderOption[]> {
  if (provider === "github") {
    const repositories = await getGitHubRepositories(accountID, connectionID);
    return repositories.map((repository) => ({
      id: String(repository.id),
      name: repository.full_name,
      detail: `${repository.private ? "Private" : "Public"} · ${repository.default_branch || "Default branch"}${repository.language ? ` · ${repository.language}` : ""}`,
      status: repository.selected ? "Selected" : "Available",
    }));
  }
  const projects = await getVercelProjects(accountID, connectionID);
  return projects.map((project) => ({
    id: project.id,
    name: project.name,
    detail: project.framework || "Framework not detected",
    status: project.selected ? "Selected" : "Available",
  }));
}

function ConnectedProviderResource({ provider, resource }: { provider: Provider; resource: ProjectResource }) {
  const copy = providerCopy[provider];
  const details = provider === "github"
    ? [
      ["Status", formatStatus(resource.status)],
      ["Visibility", resource.metadata.private === true ? "Private" : "Public"],
      ["Default branch", stringMetadata(resource, "default_branch", "Not reported")],
      ["Language", stringMetadata(resource, "language", "Not reported")],
    ]
    : [
      ["Status", formatStatus(resource.status)],
      ["Framework", stringMetadata(resource, "framework", "Not detected")],
      ["Vercel project ID", resource.external_id],
      ["Account", stringMetadata(resource, "account_id", "Current connection")],
    ];
  const githubURL = stringMetadata(resource, "html_url", "");
  const dashboardURL = provider === "github" && githubURL ? githubURL : "https://vercel.com/dashboard";

  return (
    <section id={copy.sectionID} className={styles.section}>
      <header className={styles.heading}><div><p className={styles.eyebrow}>{copy.eyebrow}</p><h2>{resource.name}</h2><p>This {providerLabel(provider)} {copy.resourceLabel} is attached and ready for project-level operations.</p></div><span className={styles.providerMark}>{copy.mark}</span></header>
      <div className={styles.connectedGrid}>{details.map(([label, value]) => <div key={label}><span>{label}</span><strong>{value}</strong></div>)}</div>
      <div className={styles.connectedActions}><a href={dashboardURL} target="_blank" rel="noreferrer">Open in {providerLabel(provider)} ↗</a><span>Deployment controls and live status will build on this attachment.</span></div>
    </section>
  );
}

function stringMetadata(resource: ProjectResource, key: string, fallback: string): string {
  return typeof resource.metadata[key] === "string" && resource.metadata[key] ? resource.metadata[key] as string : fallback;
}

function providerLabel(provider: Provider): string {
  return provider === "github" ? "GitHub" : "Vercel";
}

function formatStatus(status: string): string {
  return status.toLowerCase().split("_").map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join(" ");
}
