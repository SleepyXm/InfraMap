"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import type { CreateWorkspace, Workspace } from "@/app/components/types/workspaces";
import styles from "./WorkspaceList.module.css";

type WorkspaceListProps = {
  workspaces: Workspace[];
  loading: boolean;
  error: string | null;
  canCreate: boolean;
  canDelete: boolean;
  onDelete: (workspaceID: string) => Promise<void>;
  composerOpen: boolean;
  onComposerChange: (open: boolean) => void;
  onCreate: (input: CreateWorkspace) => Promise<Workspace>;
  onRetry: () => Promise<void>;
  onCreated: (workspace: Workspace) => void;
};

export function WorkspaceList({ workspaces, loading, error, canCreate, canDelete, onDelete, composerOpen, onComposerChange, onCreate, onRetry, onCreated }: WorkspaceListProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [saving, setSaving] = useState(false);
  const [formError, setFormError] = useState("");

  function setComposer(open: boolean) {
    setFormError("");
    onComposerChange(open);
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setFormError("");
    try {
      const workspace = await onCreate({ name, description });
      setName("");
      setDescription("");
      setComposer(false);
      onCreated(workspace);
    } catch (cause) {
      setFormError(cause instanceof Error ? cause.message : "Workspace could not be created.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className={styles.workspaces}>
      <header className={styles.heading}>
        <div>
          <p className={styles.eyebrow}>Workspaces</p>
          <h1>Workspaces, clearly in view.</h1>
          <p>Create the workspace first, then attach its sources, discovered services, environments, and provider resources.</p>
        </div>
        <button type="button" className={styles.primaryAction} disabled={!canCreate} onClick={() => setComposer(true)}>
          {canCreate ? "New workspace" : "Viewer access"}
        </button>
      </header>

      {composerOpen && canCreate && (
        <form className={styles.composer} onSubmit={submit}>
          <div className={styles.composerHeading}>
            <div>
              <p className={styles.eyebrow}>Create workspace</p>
              <h2>Give the control plane a name.</h2>
            </div>
            <button type="button" className={styles.closeButton} aria-label="Close workspace form" onClick={() => setComposer(false)}>×</button>
          </div>
          <label>
            <span>Workspace name</span>
            <input value={name} onChange={(event) => setName(event.target.value)} maxLength={80} placeholder="payment-api" autoFocus required />
          </label>
          <label>
            <span>Description <small>Optional</small></span>
            <textarea value={description} onChange={(event) => setDescription(event.target.value)} maxLength={280} rows={3} placeholder="The source, runtime, data, and infrastructure for this service." />
          </label>
          {formError && <p className={styles.error}>{formError}</p>}
          <footer className={styles.composerActions}>
            <button type="button" className={styles.secondaryAction} onClick={() => setComposer(false)}>Cancel</button>
            <button type="submit" className={styles.primaryAction} disabled={saving || !name.trim()}>{saving ? "Creating…" : "Create workspace"}</button>
          </footer>
        </form>
      )}

      {error && (
        <div className={styles.errorState}>
          <p>{error}</p>
          <button type="button" className={styles.secondaryAction} onClick={() => void onRetry()}>Retry</button>
        </div>
      )}

      {!error && loading ? (
        <p className={styles.state}>Loading workspaces…</p>
      ) : !error && workspaces.length === 0 ? (
        <div className={styles.empty}>
          <span className={styles.emptyMark}>01</span>
          <h2>Your first workspace starts here.</h2>
          <p>Create the workspace now. Sources and resources can be attached afterwards without blocking setup.</p>
          {canCreate && <button type="button" className={styles.primaryAction} onClick={() => setComposer(true)}>Create a workspace</button>}
        </div>
      ) : !error ? (
        <div className={styles.grid}>
          {workspaces.map((workspace) => <WorkspaceCard key={workspace.id} workspace={workspace} onDelete={canDelete ? onDelete : undefined} />)}
        </div>
      ) : null}
    </section>
  );
}

export function WorkspacePreviewList({ workspaces, onOpen }: { workspaces: Workspace[]; onOpen: () => void; }) {
  if (workspaces.length === 0) {
    return (
      <div className={styles.previewEmpty}>
        <p>No workspaces yet. Create one before attaching infrastructure.</p>
        <button type="button" className={styles.secondaryAction} onClick={onOpen}>Open workspaces</button>
      </div>
    );
  }
  return (
    <div className={styles.previewList}>
      {workspaces.slice(0, 4).map((workspace) => (
        <Link key={workspace.id} className={styles.previewRow} href={`/workspaces/${workspace.account_id}/${workspace.id}`}>
          <span><strong>{workspace.name}</strong><small>{workspace.description || "Ready for connections"}</small></span>
          <em>{workspace.status}</em>
        </Link>
      ))}
      <button type="button" className={styles.previewMore} onClick={onOpen}>View all workspaces →</button>
    </div>
  );
}

function WorkspaceCard({ workspace, onDelete }: { workspace: Workspace; onDelete?: (workspaceID: string) => Promise<void>; }) {
  const [confirming, setConfirming] = useState(false);
  const [confirmation, setConfirmation] = useState("");
  const [deleting, setDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState("");

  async function remove(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!onDelete || deleting || confirmation !== workspace.name) return;
    setDeleting(true);
    setDeleteError("");
    try {
      await onDelete(workspace.id);
    } catch (cause) {
      setDeleteError(cause instanceof Error ? cause.message : "Workspace could not be deleted.");
    } finally {
      setDeleting(false);
    }
  }

  return (
    <article className={styles.card}>
      <div className={styles.cardTop}>
        <span className={styles.workspaceMark}>{workspace.name.slice(0, 2).toUpperCase()}</span>
        <span className={styles.status}>{workspace.status}</span>
      </div>
      <div>
        <h2>{workspace.name}</h2>
        <p>{workspace.description || "Ready for sources, services, deployments, and metrics."}</p>
      </div>
      <div className={styles.providers} aria-label="Workspace providers">
        <span>Source</span><span>Deploy</span><span>Data</span><span>Cloud</span>
      </div>
      <Link className={styles.openWorkspace} href={`/workspaces/${workspace.account_id}/${workspace.id}`}>Open workspace <span>→</span></Link>
      {onDelete && (confirming ? (
        <form className={styles.composer} onSubmit={remove} aria-label={`Delete ${workspace.name}`}>
          <strong>Delete {workspace.name}?</strong>
          <p>This permanently removes the InfraMap workspace, its attachments, and saved scan data. Hosted services, repositories, and account connections remain intact. This cannot be undone.</p>
          <label>
            <span>Type {workspace.name} to confirm</span>
            <input value={confirmation} onChange={(event) => setConfirmation(event.target.value)} disabled={deleting} autoComplete="off" autoFocus required />
          </label>
          {deleteError && <p className={styles.error} role="alert">{deleteError}</p>}
          <div className={styles.composerActions}>
            <button type="button" className={styles.secondaryAction} disabled={deleting} onClick={() => { setConfirming(false); setConfirmation(""); setDeleteError(""); }}>Cancel</button>
            <button type="submit" className={styles.primaryAction} disabled={deleting || confirmation !== workspace.name}>{deleting ? "Deleting…" : "Delete workspace"}</button>
          </div>
        </form>
      ) : (
        <button type="button" className={styles.secondaryAction} onClick={() => setConfirming(true)} aria-label={`Delete ${workspace.name}`}>Delete workspace</button>
      ))}
    </article>
  );
}
