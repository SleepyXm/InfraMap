"use client";

import { FormEvent, useState } from "react";
import Link from "next/link";
import type { CreateWorkspaceProject, WorkspaceProject } from "./types";
import styles from "./ProjectWorkspace.module.css";

type ProjectWorkspaceProps = {
  projects: WorkspaceProject[];
  loading: boolean;
  error: string | null;
  canCreate: boolean;
  composerOpen: boolean;
  onComposerChange: (open: boolean) => void;
  onCreate: (input: CreateWorkspaceProject) => Promise<WorkspaceProject>;
  onRetry: () => Promise<void>;
  onCreated: (project: WorkspaceProject) => void;
};

export function ProjectWorkspace({ projects, loading, error, canCreate, composerOpen, onComposerChange, onCreate, onRetry, onCreated }: ProjectWorkspaceProps) {
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
      const project = await onCreate({ name, description });
      setName("");
      setDescription("");
      setComposer(false);
      onCreated(project);
    } catch (cause) {
      setFormError(cause instanceof Error ? cause.message : "Project could not be created.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <section className={styles.projects}>
      <header className={styles.heading}>
        <div>
          <p className={styles.eyebrow}>Workspaces</p>
          <h1>Projects, clearly in view.</h1>
          <p>Create the project first, then attach its GitHub source, Vercel deployment, Supabase data, and AWS runtime.</p>
        </div>
        <button type="button" className={styles.primaryAction} disabled={!canCreate} onClick={() => setComposer(true)}>
          {canCreate ? "New project" : "Viewer access"}
        </button>
      </header>

      {composerOpen && canCreate && (
        <form className={styles.composer} onSubmit={submit}>
          <div className={styles.composerHeading}>
            <div>
              <p className={styles.eyebrow}>Create project</p>
              <h2>Give the control plane a name.</h2>
            </div>
            <button type="button" className={styles.closeButton} aria-label="Close project form" onClick={() => setComposer(false)}>×</button>
          </div>
          <label>
            <span>Project name</span>
            <input value={name} onChange={(event) => setName(event.target.value)} maxLength={80} placeholder="payment-api" autoFocus required />
          </label>
          <label>
            <span>Description <small>Optional</small></span>
            <textarea value={description} onChange={(event) => setDescription(event.target.value)} maxLength={280} rows={3} placeholder="The source, runtime, data, and infrastructure for this service." />
          </label>
          {formError && <p className={styles.error}>{formError}</p>}
          <footer className={styles.composerActions}>
            <button type="button" className={styles.secondaryAction} onClick={() => setComposer(false)}>Cancel</button>
            <button type="submit" className={styles.primaryAction} disabled={saving || !name.trim()}>{saving ? "Creating…" : "Create project"}</button>
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
        <p className={styles.state}>Loading projects…</p>
      ) : !error && projects.length === 0 ? (
        <div className={styles.empty}>
          <span className={styles.emptyMark}>01</span>
          <h2>Your first project starts here.</h2>
          <p>Create the project now. Connections can be attached afterwards without blocking the setup.</p>
          {canCreate && <button type="button" className={styles.primaryAction} onClick={() => setComposer(true)}>Create a project</button>}
        </div>
      ) : !error ? (
        <div className={styles.grid}>
          {projects.map((project) => <ProjectCard key={project.id} project={project} />)}
        </div>
      ) : null}
    </section>
  );
}

export function ProjectPreviewList({ projects, onOpen }: { projects: WorkspaceProject[]; onOpen: () => void }) {
  if (projects.length === 0) {
    return (
      <div className={styles.previewEmpty}>
        <p>No projects yet. Create one before attaching infrastructure.</p>
        <button type="button" className={styles.secondaryAction} onClick={onOpen}>Open workspaces</button>
      </div>
    );
  }
  return (
    <div className={styles.previewList}>
      {projects.slice(0, 4).map((project) => (
        <Link key={project.id} className={styles.previewRow} href={`/workspaces/${project.account_id}/projects/${project.id}`}>
          <span><strong>{project.name}</strong><small>{project.description || "Ready for connections"}</small></span>
          <em>{project.status}</em>
        </Link>
      ))}
      <button type="button" className={styles.previewMore} onClick={onOpen}>View all projects →</button>
    </div>
  );
}

function ProjectCard({ project }: { project: WorkspaceProject }) {
  return (
    <article className={styles.card}>
      <div className={styles.cardTop}>
        <span className={styles.projectMark}>{project.name.slice(0, 2).toUpperCase()}</span>
        <span className={styles.status}>{project.status}</span>
      </div>
      <div>
        <h2>{project.name}</h2>
        <p>{project.description || "Ready for source, deployment, data, and infrastructure connections."}</p>
      </div>
      <div className={styles.providers} aria-label="Project providers">
        <span>Source</span><span>Deploy</span><span>Data</span><span>Cloud</span>
      </div>
      <Link className={styles.openProject} href={`/workspaces/${project.account_id}/projects/${project.id}`}>Open project <span>→</span></Link>
    </article>
  );
}
