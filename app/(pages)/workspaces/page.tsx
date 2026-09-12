"use client";

import Link from "next/link";
import { useMemo, useState, Suspense } from "react";
import { useSearchParams } from "next/navigation";
import {
  beginWorkspaceOAuth,
  getGitHubRepositories,
  getSupabaseProjects,
  getVercelProjects,
  updateGitHubRepositories,
  updateSupabaseProjects,
  updateVercelProjects,
  useWorkspaceConnections,
} from "@/app/features/workspaces/connections";
import { ProjectPreviewList, ProjectWorkspace, useWorkspaceProjects } from "@/app/features/workspaces/projects";
import { useUser } from "@/app/components/provider/UserProvider";
import type {
  Connection,
  ConnectionProvider,
  GitHubRepository,
  SupabaseProject,
  VercelProject,
} from "@/app/components/types/users";
import { ConnectionActions } from "@/app/UI/ConnectionActions";
import { ResourceSelector } from "@/app/UI/ResourceSelector";
import styles from "@/app/UI/Workspace.module.css";

type Provider = {
  id: ConnectionProvider;
  name: string;
  mark: string;
  category: string;
  description: string;
  pendingLabel: string;
};

const providers: Provider[] = [
  {
    id: "github",
    name: "GitHub",
    mark: "GH",
    category: "Source",
    description: "Choose repositories and track the commits that become deployments.",
    pendingLabel: "Connect GitHub repositories",
  },
  {
    id: "vercel",
    name: "Vercel",
    mark: "▲",
    category: "Deployment",
    description: "Bring projects, deployments, domains, logs, and environment configuration together.",
    pendingLabel: "Connect Vercel projects",
  },
  {
    id: "supabase",
    name: "Supabase",
    mark: "S",
    category: "Data",
    description: "Attach organisations and projects through the Management API.",
    pendingLabel: "Connect Supabase projects",
  },
  {
    id: "cloudflare",
    name: "Cloudflare",
    mark: "CF",
    category: "Cloud edge",
    description: "Bring zones, DNS, Workers, Pages, and edge configuration into project context.",
    pendingLabel: "API token setup next",
  },
  {
    id: "aws",
    name: "AWS",
    mark: "AWS",
    category: "Cloud",
    description: "Discover the services behind a project through a scoped cross-account role.",
    pendingLabel: "Role connection next",
  },
];

type WorkspaceTab = "overview" | "workspaces" | "connections";

export default function WorkspacesPage() {
  return (
    <Suspense fallback={<main className={styles.state}>Loading your workspaces…</main>}>
      <WorkspaceContent />
    </Suspense>
  );
}

function WorkspaceContent() {
  const searchParams = useSearchParams();
  const { user, resolved } = useUser();
  const [tab, setTab] = useState<WorkspaceTab>(() => getWorkspaceTab(searchParams.get("tab")));
  const [selectedAccountID, setSelectedAccountID] = useState("");
  const [notice, setNotice] = useState(() => getConnectionNotice(searchParams));
  const [projectComposerOpen, setProjectComposerOpen] = useState(false);
  const [repositoryConnectionID, setRepositoryConnectionID] = useState("");
  const [repositories, setRepositories] = useState<GitHubRepository[]>([]);
  const [repositorySearch, setRepositorySearch] = useState("");
  const [selectedRepositoryIDs, setSelectedRepositoryIDs] = useState<Set<number>>(new Set());
  const [repositoryLoading, setRepositoryLoading] = useState(false);
  const [repositorySaving, setRepositorySaving] = useState(false);
  const [repositoryError, setRepositoryError] = useState("");
  const [supabaseConnectionID, setSupabaseConnectionID] = useState("");
  const [supabaseProjects, setSupabaseProjects] = useState<SupabaseProject[]>([]);
  const [supabaseSearch, setSupabaseSearch] = useState("");
  const [selectedSupabaseRefs, setSelectedSupabaseRefs] = useState<Set<string>>(new Set());
  const [supabaseLoading, setSupabaseLoading] = useState(false);
  const [supabaseSaving, setSupabaseSaving] = useState(false);
  const [supabaseError, setSupabaseError] = useState("");
  const [vercelConnectionID, setVercelConnectionID] = useState("");
  const [vercelProjects, setVercelProjects] = useState<VercelProject[]>([]);
  const [vercelSearch, setVercelSearch] = useState("");
  const [selectedVercelProjectIDs, setSelectedVercelProjectIDs] = useState<Set<string>>(new Set());
  const [vercelLoading, setVercelLoading] = useState(false);
  const [vercelSaving, setVercelSaving] = useState(false);
  const [vercelError, setVercelError] = useState("");
  const accounts = useMemo(() => user?.accounts ?? [], [user?.accounts]);
  const selectedAccount = useMemo(
    () => accounts.find((account) => account.id === selectedAccountID) ?? accounts[0],
    [accounts, selectedAccountID],
  );
  const { connections, loading, error, refresh } = useWorkspaceConnections(selectedAccount?.id);
  const {
    projects,
    loading: projectsLoading,
    error: projectsError,
    refresh: refreshProjects,
    createProject,
  } = useWorkspaceProjects(selectedAccount?.id);
  const githubIdentity = user?.identities?.find((identity) => identity.provider === "github");

  function changeWorkspace(accountID: string) {
    setSelectedAccountID(accountID);
    setProjectComposerOpen(false);
    setRepositoryConnectionID("");
    setRepositories([]);
    setSelectedRepositoryIDs(new Set());
    setRepositoryError("");
    setSupabaseConnectionID("");
    setSupabaseProjects([]);
    setSelectedSupabaseRefs(new Set());
    setSupabaseError("");
    setVercelConnectionID("");
    setVercelProjects([]);
    setSelectedVercelProjectIDs(new Set());
    setVercelError("");
  }

  const connectedProviders = useMemo(
    () => new Set(connections.map((connection) => connection.provider)),
    [connections],
  );
  const activeConnections = connections.filter((connection) => connection.status === "active").length;
  const selectedRepositoryCount = connections.reduce(
    (total, connection) => total + getSelectedRepositoryCount(connection),
    0,
  );
  const filteredRepositories = repositories.filter((repository) =>
    repository.full_name.toLowerCase().includes(repositorySearch.trim().toLowerCase()),
  );
  const filteredSupabaseProjects = supabaseProjects.filter((project) =>
    [project.name, project.ref, project.organization_slug]
      .some((value) => value.toLowerCase().includes(supabaseSearch.trim().toLowerCase())),
  );
  const filteredVercelProjects = vercelProjects.filter((project) =>
    [project.name, project.id, project.framework]
      .some((value) => value.toLowerCase().includes(vercelSearch.trim().toLowerCase())),
  );

  async function openRepositoryPicker(connection: Connection) {
    if (!selectedAccount) return;
    setSupabaseConnectionID("");
    setVercelConnectionID("");
    setRepositoryConnectionID(connection.id);
    setRepositoryLoading(true);
    setRepositoryError("");
    setRepositorySearch("");

    try {
      const available = await getGitHubRepositories(selectedAccount.id, connection.id);
      setRepositories(available);
      setSelectedRepositoryIDs(new Set(available.filter((repository) => repository.selected).map((repository) => repository.id)));
    } catch (caught) {
      setRepositoryError(caught instanceof Error ? caught.message : "Repositories could not be loaded.");
    } finally {
      setRepositoryLoading(false);
    }
  }

  async function saveRepositorySelection() {
    if (!selectedAccount || !repositoryConnectionID) return;
    setRepositorySaving(true);
    setRepositoryError("");

    try {
      await updateGitHubRepositories(
        selectedAccount.id,
        repositoryConnectionID,
        Array.from(selectedRepositoryIDs),
      );
      await refresh();
      setNotice("Repository access saved for this workspace.");
      setRepositoryConnectionID("");
    } catch (caught) {
      setRepositoryError(caught instanceof Error ? caught.message : "Repository selection could not be saved.");
    } finally {
      setRepositorySaving(false);
    }
  }

  function toggleRepository(repositoryID: number) {
    setSelectedRepositoryIDs((current) => {
      const next = new Set(current);
      if (next.has(repositoryID)) next.delete(repositoryID);
      else next.add(repositoryID);
      return next;
    });
  }

  async function openSupabasePicker(connection: Connection) {
    if (!selectedAccount) return;
    setRepositoryConnectionID("");
    setVercelConnectionID("");
    setSupabaseConnectionID(connection.id);
    setSupabaseLoading(true);
    setSupabaseError("");
    setSupabaseSearch("");

    try {
      const available = await getSupabaseProjects(selectedAccount.id, connection.id);
      setSupabaseProjects(available);
      setSelectedSupabaseRefs(new Set(available.filter((project) => project.selected).map((project) => project.ref)));
    } catch (caught) {
      setSupabaseError(caught instanceof Error ? caught.message : "Supabase projects could not be loaded.");
    } finally {
      setSupabaseLoading(false);
    }
  }

  async function saveSupabaseSelection() {
    if (!selectedAccount || !supabaseConnectionID) return;
    setSupabaseSaving(true);
    setSupabaseError("");

    try {
      await updateSupabaseProjects(
        selectedAccount.id,
        supabaseConnectionID,
        Array.from(selectedSupabaseRefs),
      );
      await refresh();
      setNotice("Supabase project access saved for this workspace.");
      setSupabaseConnectionID("");
    } catch (caught) {
      setSupabaseError(caught instanceof Error ? caught.message : "Supabase project selection could not be saved.");
    } finally {
      setSupabaseSaving(false);
    }
  }

  function toggleSupabaseProject(projectRef: string) {
    setSelectedSupabaseRefs((current) => {
      const next = new Set(current);
      if (next.has(projectRef)) next.delete(projectRef);
      else next.add(projectRef);
      return next;
    });
  }

  async function openVercelPicker(connection: Connection) {
    if (!selectedAccount) return;
    setRepositoryConnectionID("");
    setSupabaseConnectionID("");
    setVercelConnectionID(connection.id);
    setVercelLoading(true);
    setVercelError("");
    setVercelSearch("");

    try {
      const available = await getVercelProjects(selectedAccount.id, connection.id);
      setVercelProjects(available);
      setSelectedVercelProjectIDs(new Set(available.filter((project) => project.selected).map((project) => project.id)));
    } catch (caught) {
      setVercelError(caught instanceof Error ? caught.message : "Vercel projects could not be loaded.");
    } finally {
      setVercelLoading(false);
    }
  }

  async function saveVercelSelection() {
    if (!selectedAccount || !vercelConnectionID) return;
    setVercelSaving(true);
    setVercelError("");

    try {
      await updateVercelProjects(selectedAccount.id, vercelConnectionID, Array.from(selectedVercelProjectIDs));
      await refresh();
      setNotice("Vercel project access saved for this workspace.");
      setVercelConnectionID("");
    } catch (caught) {
      setVercelError(caught instanceof Error ? caught.message : "Vercel project selection could not be saved.");
    } finally {
      setVercelSaving(false);
    }
  }

  function toggleVercelProject(projectID: string) {
    setSelectedVercelProjectIDs((current) => {
      const next = new Set(current);
      if (next.has(projectID)) next.delete(projectID);
      else next.add(projectID);
      return next;
    });
  }

  async function connectProvider(provider: ConnectionProvider) {
    if (!selectedAccount) return;
    const result = await beginWorkspaceOAuth(selectedAccount.id, provider);
    if (result !== "completed") return;
    await refresh();
    setNotice("Vercel workspace access is connected. Select the projects InfraMap should manage.");
  }

  if (!resolved) return <main className={styles.state}>Loading your workspaces…</main>;

  if (!user) {
    return (
      <main className={styles.state}>
        <div>
          <span className={styles.stateMark}>IM</span>
          <h1>Sign in to open your control plane.</h1>
          <Link href="/login">Continue to sign in</Link>
        </div>
      </main>
    );
  }

  if (!selectedAccount) return <main className={styles.state}>No workspace is attached to this account yet.</main>;

  return (
    <main className={styles.page}>
      <div className={styles.shell}>
        <header className={styles.workspaceHeader}>
          <div className={styles.workspaceIdentity}>
            <span className={styles.workspaceMark}>{selectedAccount.name.slice(0, 2).toUpperCase()}</span>
            <div>
              <label htmlFor="workspace-select">Workspace</label>
              <select
                id="workspace-select"
                value={selectedAccount.id}
                onChange={(event) => changeWorkspace(event.target.value)}
              >
                {accounts.map((account) => (
                  <option key={account.id} value={account.id}>{account.name}</option>
                ))}
              </select>
            </div>
          </div>
          <div className={styles.headerActions}>
            <div className={styles.headerMeta}>
              <span>{selectedAccount.role}</span>
              <span>{selectedAccount.type}</span>
            </div>
            <button
              type="button"
              className={styles.headerCreate}
              disabled={selectedAccount.role === "viewer"}
              onClick={() => {
                setTab("workspaces");
                setProjectComposerOpen(true);
              }}
            >
              New project
            </button>
          </div>
        </header>

        <nav className={styles.tabs} aria-label="Workspace sections">
          <button type="button" className={tab === "overview" ? styles.activeTab : ""} onClick={() => setTab("overview")}>Overview</button>
          <button type="button" className={tab === "workspaces" ? styles.activeTab : ""} onClick={() => setTab("workspaces")}>Workspaces</button>
          <button type="button" className={tab === "connections" ? styles.activeTab : ""} onClick={() => setTab("connections")}>Connections</button>
        </nav>

        {notice && <p className={styles.notice}>{notice}</p>}

        {tab === "overview" ? (
          <div className={styles.overview}>
            <section className={styles.intro}>
              <div>
                <p className={styles.eyebrow}>Workspace overview</p>
                <h1>{selectedAccount.name}</h1>
                <p>One place to understand the source, deployments, data, and infrastructure behind every project.</p>
              </div>
              <button
                type="button"
                className={styles.primaryAction}
                disabled={selectedAccount.role === "viewer"}
                onClick={() => {
                  setTab("workspaces");
                  setProjectComposerOpen(true);
                }}
              >
                New project
              </button>
            </section>

            <section className={styles.metrics} aria-label="Workspace summary">
              <Metric label="Projects" value={String(projects.length)} detail={projects.length ? "Inside this workspace" : "Ready to create"} />
              <Metric label="Active connections" value={String(activeConnections)} detail={githubIdentity ? "GitHub identity linked" : "No source identity"} />
              <Metric label="Repositories" value={String(selectedRepositoryCount)} detail="Attached to this workspace" />
              <Metric label="Drift" value="—" detail="No environments tracked" />
            </section>

            <div className={styles.overviewGrid}>
              <section className={styles.panel}>
                <div className={styles.panelHeader}>
                  <div>
                    <p className={styles.eyebrow}>Projects</p>
                    <h2>{projects.length ? "Your control planes" : "Nothing attached yet"}</h2>
                  </div>
                  <span className={styles.count}>{projects.length}</span>
                </div>
                <ProjectPreviewList projects={projects} onOpen={() => setTab("workspaces")} />
              </section>

              <section className={styles.panel}>
                <div className={styles.panelHeader}>
                  <div>
                    <p className={styles.eyebrow}>Connection health</p>
                    <h2>Provider access</h2>
                  </div>
                </div>
                <div className={styles.healthList}>
                  {providers.map((provider) => {
                    const connected = connectedProviders.has(provider.id);
                    const identityOnly = provider.id === "github" && githubIdentity && !connected;
                    return (
                      <div key={provider.id} className={styles.healthRow}>
                        <span className={styles.providerMark}>{provider.mark}</span>
                        <div>
                          <strong>{provider.name}</strong>
                          <small>{connected ? "Operational connection" : identityOnly ? "Identity linked · repository access needed" : "Not connected"}</small>
                        </div>
                        <span className={connected ? styles.statusActive : identityOnly ? styles.statusPartial : styles.statusIdle}>
                          {connected ? "Active" : identityOnly ? "Partial" : "Offline"}
                        </span>
                      </div>
                    );
                  })}
                </div>
              </section>
            </div>
          </div>
        ) : tab === "workspaces" ? (
          <ProjectWorkspace
            projects={projects}
            loading={projectsLoading}
            error={projectsError}
            canCreate={selectedAccount.role !== "viewer"}
            composerOpen={projectComposerOpen}
            onComposerChange={setProjectComposerOpen}
            onCreate={createProject}
            onRetry={refreshProjects}
            onCreated={(project) => setNotice(`${project.name} was created. Connections can now be attached.`)}
          />
        ) : (
          <section className={styles.connections}>
            <div className={styles.connectionsHead}>
              <div>
                <p className={styles.eyebrow}>Connections</p>
                <h1>Give this workspace context.</h1>
                <p>Provider access belongs to the workspace. Your profile identities remain separate and can still be used to sign in.</p>
              </div>
              {error && <button type="button" className={styles.retry} onClick={() => void refresh()}>Retry loading</button>}
            </div>

            {error && <p className={styles.error}>{error}</p>}
            <div className={styles.providerGrid} aria-busy={loading}>
              {providers.map((provider) => {
                const connection = connections.find((item) => item.provider === provider.id);
                const identityOnly = provider.id === "github" && githubIdentity && !connection;
                const selectedCount = connection
                  ? provider.id === "github"
                    ? getSelectedRepositoryCount(connection)
                    : provider.id === "supabase" || provider.id === "vercel"
                      ? getSelectedProjectCount(connection)
                      : 0
                  : 0;
                return (
                  <article key={provider.id} className={styles.providerCard}>
                    <div className={styles.providerTop}>
                      <span className={styles.providerMarkLarge}>{provider.mark}</span>
                      <span className={connection ? styles.statusActive : identityOnly ? styles.statusPartial : styles.statusIdle}>
                        {connection ? "Active" : identityOnly ? "Identity linked" : "Not connected"}
                      </span>
                    </div>
                    <div>
                      <small>{provider.category}</small>
                      <h2>{provider.name}</h2>
                      <p>{provider.description}</p>
                    </div>
                    {provider.id === "github" && connection ? (
                      <ConnectionActions
                        primaryLabel={selectedCount ? `${selectedCount} selected` : "Select repositories"}
                        reconnectLabel="Reconnect"
                        onPrimary={() => void openRepositoryPicker(connection)}
                        onReconnect={() => void connectProvider(provider.id)}
                      />
                    ) : provider.id === "supabase" && connection ? (
                      <ConnectionActions
                        primaryLabel={selectedCount ? `${selectedCount} selected` : "Select projects"}
                        reconnectLabel="Reconnect"
                        onPrimary={() => void openSupabasePicker(connection)}
                        onReconnect={() => void connectProvider(provider.id)}
                      />
                    ) : provider.id === "vercel" && connection ? (
                      <ConnectionActions
                        primaryLabel={selectedCount ? `${selectedCount} selected` : "Select projects"}
                        reconnectLabel="Reconnect"
                        onPrimary={() => void openVercelPicker(connection)}
                        onReconnect={() => void connectProvider(provider.id)}
                      />
                    ) : provider.id === "github" || provider.id === "supabase" || provider.id === "vercel" ? (
                      <button
                        type="button"
                        className={styles.connectButton}
                        onClick={() => void connectProvider(provider.id)}
                      >
                        <span>{provider.pendingLabel}</span><span>→</span>
                      </button>
                    ) : connection ? (
                      <div className={styles.connectionDetail}><span>{connection.name}</span><strong>Connected</strong></div>
                    ) : (
                      <button type="button" className={styles.connectButton} disabled>{provider.pendingLabel}</button>
                    )}
                  </article>
                );
              })}
            </div>

            {repositoryConnectionID && (
              <ResourceSelector
                eyebrow="GitHub source"
                title="Select repositories"
                description="Choose which repositories belong in this workspace. You can revise this list at any time."
                permissionNote="GitHub authorises repository access at account level. InfraMap only attaches the repositories selected here to this workspace."
                searchPlaceholder="Find a repository"
                emptyMessage="No accessible repositories match this search."
                options={filteredRepositories.map((repository) => ({
                  id: repository.id,
                  title: repository.full_name,
                  description: repository.description || [repository.language, repository.default_branch].filter(Boolean).join(" · "),
                  tag: repository.private ? "Private" : "Public",
                }))}
                selected={selectedRepositoryIDs}
                search={repositorySearch}
                loading={repositoryLoading}
                saving={repositorySaving}
                error={repositoryError}
                onSearchChange={setRepositorySearch}
                onToggle={toggleRepository}
                onClose={() => setRepositoryConnectionID("")}
                onSave={() => void saveRepositorySelection()}
              />
            )}

            {supabaseConnectionID && (
              <ResourceSelector
                eyebrow="Supabase data"
                title="Select Supabase projects"
                description="Attach the Supabase projects that support applications in this workspace."
                permissionNote="InfraMap stores the short-lived access and refresh tokens encrypted. Project credentials are not sent to the browser by this connection."
                searchPlaceholder="Find a Supabase project"
                emptyMessage="No authorised Supabase projects match this search."
                options={filteredSupabaseProjects.map((project) => ({
                  id: project.ref,
                  title: project.name,
                  description: [project.organization_slug, project.region, project.database?.version].filter(Boolean).join(" · "),
                  tag: formatSupabaseStatus(project.status),
                }))}
                selected={selectedSupabaseRefs}
                search={supabaseSearch}
                loading={supabaseLoading}
                saving={supabaseSaving}
                error={supabaseError}
                onSearchChange={setSupabaseSearch}
                onToggle={toggleSupabaseProject}
                onClose={() => setSupabaseConnectionID("")}
                onSave={() => void saveSupabaseSelection()}
              />
            )}

            {vercelConnectionID && (
              <ResourceSelector
                eyebrow="Vercel deployment"
                title="Select Vercel projects"
                description="Attach the deployed projects that InfraMap should place inside this workspace's control layer."
                permissionNote="InfraMap keeps the Vercel installation token encrypted on the backend. The browser receives project metadata, never the provider credential."
                searchPlaceholder="Find a Vercel project"
                emptyMessage="No authorised Vercel projects match this search. Check the integration's project scope in Vercel."
                options={filteredVercelProjects.map((project) => ({
                  id: project.id,
                  title: project.name,
                  description: `Updated ${formatVercelTimestamp(project.updatedAt)} · ${project.id}`,
                  tag: formatVercelFramework(project.framework),
                }))}
                selected={selectedVercelProjectIDs}
                search={vercelSearch}
                loading={vercelLoading}
                saving={vercelSaving}
                error={vercelError}
                onSearchChange={setVercelSearch}
                onToggle={toggleVercelProject}
                onClose={() => setVercelConnectionID("")}
                onSave={() => void saveVercelSelection()}
              />
            )}

            <p className={styles.connectionNote}>Cloudflare is staged for scoped API tokens, while AWS is staged for a cross-account role.</p>
          </section>
        )}
      </div>
    </main>
  );
}

function Metric({ label, value, detail }: { label: string; value: string; detail: string }) {
  return (
    <article className={styles.metric}>
      <span>{label}</span>
      <strong>{value}</strong>
      <small>{detail}</small>
    </article>
  );
}

function getWorkspaceTab(tab: string | null): WorkspaceTab {
  return tab === "connections" || tab === "workspaces" ? tab : "overview";
}

function getSelectedRepositoryCount(connection: Connection): number {
  const repositories = connection.metadata?.repositories;
  return Array.isArray(repositories) ? repositories.length : 0;
}

function getSelectedProjectCount(connection: Connection): number {
  const projects = connection.metadata?.projects;
  return Array.isArray(projects) ? projects.length : 0;
}

function formatSupabaseStatus(status: string): string {
  return status
    .toLowerCase()
    .split("_")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatVercelFramework(framework: string): string {
  if (!framework) return "Other";
  return framework
    .split("-")
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(" ");
}

function formatVercelTimestamp(timestamp: number): string {
  if (!timestamp) return "recently";
  return new Intl.DateTimeFormat("en", { dateStyle: "medium" }).format(new Date(timestamp));
}

function getConnectionNotice(searchParams: Pick<URLSearchParams, "get">): string {
  const requestedProvider = searchParams.get("provider");
  const provider = requestedProvider === "supabase" ? "Supabase" : requestedProvider === "vercel" ? "Vercel" : "GitHub";
  if (searchParams.get("oauth") === "success") return `${provider} workspace access is connected.`;

  switch (searchParams.get("error")) {
    case "connection_failed":
      return `${provider} authorised access, but InfraMap could not save the workspace connection.`;
    case "token_exchange_failed":
      return `${provider} authorisation expired before InfraMap could complete it. Please try again.`;
    case "identity_lookup_failed":
      return `${provider} connected, but its account details could not be loaded.`;
    case "invalid_state":
      return `The ${provider} connection request expired. Please start it again.`;
    default:
      return searchParams.get("error") ? `${provider} could not be connected. Please try again.` : "";
  }
}
