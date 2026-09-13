"use client";

import Link from "next/link";
import { FormEvent, useEffect, useMemo, useState } from "react";
import { getSupabaseServices } from "@/app/components/handlers/connections";
import { attachSupabaseService, createSupabaseService } from "@/app/components/handlers/services";
import type { Connection } from "@/app/components/types/connections";
import type { SupabaseService } from "@/app/components/types/services";
import type { WorkspaceDetail, WorkspaceResource } from "@/app/components/types/workspaces";
import styles from "@/app/UI/WorkspaceControls.module.css";

type SetupMode = "attach" | "create";
type SupabaseOrganization = { id: string; slug: string; name: string; };

type SupabaseServiceStageProps = {
  accountID: string;
  workspaceID: string;
  role: WorkspaceDetail["access"];
  resource?: WorkspaceResource;
  connections: Connection[];
  connectionsLoading: boolean;
  connectionsError: string | null;
  onAttached: (resource: WorkspaceResource) => void;
};

const regions = [
  { value: "emea", label: "Europe, Middle East and Africa" },
  { value: "americas", label: "Americas" },
  { value: "apac", label: "Asia Pacific" },
] as const;

export function SupabaseServiceStage({ accountID, workspaceID, role, resource, connections, connectionsLoading, connectionsError, onAttached }: SupabaseServiceStageProps) {
  const [mode, setMode] = useState<SetupMode>("attach");
  const [connectionID, setConnectionID] = useState("");
  const [services, setServices] = useState<SupabaseService[]>([]);
  const [servicesLoading, setServicesLoading] = useState(false);
  const [selectedProjectRef, setSelectedProjectRef] = useState("");
  const [name, setName] = useState("");
  const [organizationSlug, setOrganizationSlug] = useState("");
  const [regionGroup, setRegionGroup] = useState<"americas" | "emea" | "apac">("emea");
  const [databasePassword, setDatabasePassword] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const selectedConnection = connections.find((connection) => connection.id === connectionID) ?? connections[0];
  const organizations = useMemo(() => getOrganizations(selectedConnection), [selectedConnection]);
  const canCreate = role === "owner" || role === "admin";

  useEffect(() => {
    if (!selectedConnection) return;
    let active = true;
    void Promise.resolve().then(async () => {
      if (!active) return;
      setConnectionID(selectedConnection.id);
      setOrganizationSlug(getOrganizations(selectedConnection)[0]?.slug || "");
      setServicesLoading(true);
      setError("");
      try {
        const available = await getSupabaseServices(accountID, selectedConnection.id);
        if (active) {
          setServices(available);
          setSelectedProjectRef(available[0]?.ref || "");
        }
      } catch (cause) {
        if (active) setError(cause instanceof Error ? cause.message : "Supabase projects could not be loaded.");
      } finally {
        if (active) setServicesLoading(false);
      }
    });
    return () => { active = false; };
  }, [accountID, selectedConnection]);

  async function attachProject() {
    if (!selectedConnection || !selectedProjectRef) return;
    setSaving(true);
    setError("");
    try {
      onAttached(await attachSupabaseService(accountID, workspaceID, selectedConnection.id, selectedProjectRef));
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Supabase project could not be attached.");
    } finally {
      setSaving(false);
    }
  }

  async function createProject(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selectedConnection) return;
    setSaving(true);
    setError("");
    try {
      const attached = await createSupabaseService(accountID, workspaceID, {
        connection_id: selectedConnection.id,
        name,
        organization_slug: organizationSlug,
        region_group: regionGroup,
        database_password: databasePassword,
      });
      setDatabasePassword("");
      onAttached(attached);
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : "Supabase project could not be created.");
    } finally {
      setSaving(false);
    }
  }

  function generatePassword() {
    const bytes = crypto.getRandomValues(new Uint8Array(24));
    const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789!@#$%^&*";
    setDatabasePassword(Array.from(bytes, (byte) => alphabet[byte % alphabet.length]).join(""));
  }

  if (resource) return <ConnectedSupabase resource={resource} />;

  return (
    <section id="data-layer" className={styles.section}>
      <header className={styles.heading}>
        <div><p className={styles.eyebrow}>Data layer</p><h2>Connect Supabase.</h2><p>Attach an existing database project or create a fresh one without leaving this operational workspace.</p></div>
        <span className={styles.providerMark}>S</span>
      </header>

      {connectionsError && <p className={styles.error}>{connectionsError}</p>}
      {connectionsLoading ? (
        <p className={styles.state}>Loading Supabase connections…</p>
      ) : connections.length === 0 ? (
        <div className={styles.emptyConnection}>
          <div><h3>Connect Supabase first.</h3><p>The workspace needs an authorised Supabase account before a project can be attached or created.</p></div>
          <Link href="/workspaces?tab=connections">Open connections →</Link>
        </div>
      ) : (
        <div className={styles.setup}>
          <nav className={styles.modeTabs} aria-label="Supabase setup method">
            <button type="button" className={mode === "attach" ? styles.modeActive : ""} onClick={() => { setMode("attach"); setError(""); }}>Attach existing</button>
            <button type="button" className={mode === "create" ? styles.modeActive : ""} disabled={!canCreate} onClick={() => { setMode("create"); setError(""); }}>Create new</button>
          </nav>

          {connections.length > 1 && (
            <label className={styles.field}><span>Supabase connection</span><select value={selectedConnection?.id || ""} onChange={(event) => setConnectionID(event.target.value)}>{connections.map((connection) => <option key={connection.id} value={connection.id}>{connection.name}</option>)}</select></label>
          )}

          {mode === "attach" ? (
            <div className={styles.attachPanel}>
              <div><h3>Use an existing Supabase project</h3><p>The selection is verified against Supabase before it is attached.</p></div>
              {servicesLoading ? <p className={styles.state}>Loading available projects…</p> : services.length ? (
                <>
                  <div className={styles.projectChoices}>
                    {services.map((project) => (
                      <label key={project.ref} className={selectedProjectRef === project.ref ? styles.projectChoiceSelected : styles.projectChoice}>
                        <input type="radio" name="supabase-project" value={project.ref} checked={selectedProjectRef === project.ref} onChange={() => setSelectedProjectRef(project.ref)} />
                        <span><strong>{project.name}</strong><small>{project.organization_slug} · {project.region || "Automatic region"}</small></span>
                        <em>{formatStatus(project.status)}</em>
                      </label>
                    ))}
                  </div>
                  <button type="button" className={styles.primaryAction} disabled={saving || !selectedProjectRef} onClick={() => void attachProject()}>{saving ? "Attaching…" : "Attach selected project"}</button>
                </>
              ) : <p className={styles.state}>No Supabase projects are available on this connection.</p>}
            </div>
          ) : (
            <form className={styles.createPanel} onSubmit={createProject}>
              <div><h3>Create a Supabase project</h3><p>Supabase will create the database inside the selected organisation. Its plan may incur charges.</p></div>
              <div className={styles.formGrid}>
                <label className={styles.field}><span>Project name</span><input value={name} onChange={(event) => setName(event.target.value)} maxLength={80} placeholder="payment-api-data" required /></label>
                <label className={styles.field}><span>Organisation</span><select value={organizationSlug} onChange={(event) => setOrganizationSlug(event.target.value)} required>{organizations.map((organization) => <option key={organization.id} value={organization.slug}>{organization.name}</option>)}</select></label>
                <label className={styles.field}><span>Region group</span><select value={regionGroup} onChange={(event) => setRegionGroup(event.target.value as typeof regionGroup)}>{regions.map((region) => <option key={region.value} value={region.value}>{region.label}</option>)}</select></label>
                <label className={styles.field}><span>Database password</span><div className={styles.passwordControl}><input type="password" value={databasePassword} onChange={(event) => setDatabasePassword(event.target.value)} minLength={16} maxLength={128} autoComplete="new-password" required /><button type="button" onClick={generatePassword}>Generate</button></div><small>Encrypted on the backend and never returned to this browser.</small></label>
              </div>
              <button type="submit" className={styles.primaryAction} disabled={saving || !name.trim() || !organizationSlug || databasePassword.length < 16}>{saving ? "Creating in Supabase…" : "Create and attach project"}</button>
            </form>
          )}
          {error && <p className={styles.error}>{error}</p>}
        </div>
      )}
    </section>
  );
}

function ConnectedSupabase({ resource }: { resource: WorkspaceResource; }) {
  const region = typeof resource.metadata.region === "string" ? resource.metadata.region : "Automatic region";
  const organization = typeof resource.metadata.organization_slug === "string" ? resource.metadata.organization_slug : "Supabase";
  return (
    <section id="data-layer" className={styles.section}>
      <header className={styles.heading}><div><p className={styles.eyebrow}>Data service</p><h2>{resource.name}</h2><p>This Supabase project is attached as a workspace service and ready for operational controls.</p></div><span className={styles.providerMark}>S</span></header>
      <div className={styles.connectedGrid}>
        <div><span>Status</span><strong>{formatStatus(resource.status)}</strong></div>
        <div><span>Region</span><strong>{region}</strong></div>
        <div><span>Organisation</span><strong>{organization}</strong></div>
        <div><span>Project reference</span><strong>{resource.external_id}</strong></div>
      </div>
      <div className={styles.connectedActions}><a href={`https://supabase.com/dashboard/project/${encodeURIComponent(resource.external_id)}`} target="_blank" rel="noreferrer">Open in Supabase ↗</a><span>Configuration, health, migrations, and environments will live here next.</span></div>
    </section>
  );
}

function getOrganizations(connection?: Connection): SupabaseOrganization[] {
  const raw = connection?.metadata?.organizations;
  if (!Array.isArray(raw)) return [];
  return raw.filter((value): value is SupabaseOrganization => {
    if (!value || typeof value !== "object") return false;
    const organization = value as Record<string, unknown>;
    return typeof organization.id === "string" && typeof organization.slug === "string" && typeof organization.name === "string";
  });
}

function formatStatus(status: string): string {
  if (status.toLowerCase() === "inactive") return "Paused";
  return status.toLowerCase().split("_").map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join(" ");
}
