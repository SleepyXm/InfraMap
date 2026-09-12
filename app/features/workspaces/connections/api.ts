import { API_BASE, request } from "@/app/components/handlers/auth";
import type { Connection, ConnectionCategory, ConnectionProvider, GitHubRepository, SupabaseProject, VercelProject } from "@/app/components/types/users";

type ConnectionResponse = {
  connections: Array<Omit<Connection, "category" | "status"> & Partial<Pick<Connection, "category" | "status">>>;
};

const providerCategories: Partial<Record<ConnectionProvider, ConnectionCategory>> = {
  github: "source",
  gitlab: "source",
  bitbucket: "source",
  vercel: "deployment",
  railway: "deployment",
  render: "deployment",
  netlify: "deployment",
  supabase: "data",
  neon: "data",
  planetscale: "data",
  aws: "cloud",
  gcp: "cloud",
  azure: "cloud",
  cloudflare: "cloud",
};

export async function getWorkspaceConnections(accountID: string): Promise<Connection[]> {
  const response = await request<ConnectionResponse>(`/api/auth/${accountID}/connections`, { method: "GET" });
  return response.connections.map((connection) => ({
    ...connection,
    category: connection.category ?? providerCategories[connection.provider] ?? "account",
    status: connection.status ?? "active",
  }));
}

export async function deleteWorkspaceConnection(accountID: string, connectionID: string): Promise<void> {
  await request(`/api/auth/${accountID}/connections/${connectionID}`, { method: "DELETE" });
}

export async function beginWorkspaceOAuth(accountID: string, provider: ConnectionProvider): Promise<"redirected" | "completed"> {
  if (typeof window === "undefined") return "redirected";
  if (!API_BASE) throw new Error("API server is not configured.");

  const startURL = `${API_BASE}/api/auth/${accountID}/connections/oauth/${provider}/start`;
  if (provider !== "vercel") {
    window.location.assign(startURL);
    return "redirected";
  }

  const popup = window.open(startURL, "inframap-vercel-connect", "popup,width=760,height=780");
  if (!popup) {
    window.location.assign(startURL);
    return "redirected";
  }
  await new Promise<void>((resolve) => {
    const interval = window.setInterval(() => {
      if (!popup.closed) return;
      window.clearInterval(interval);
      resolve();
    }, 400);
  });
  return "completed";
}

export async function getGitHubRepositories(accountID: string, connectionID: string): Promise<GitHubRepository[]> {
  const response = await request<{ repositories: GitHubRepository[] }>(
    `/api/auth/${accountID}/connections/${connectionID}/github/repositories`,
    { method: "GET" },
  );
  return response.repositories;
}

export async function updateGitHubRepositories(accountID: string, connectionID: string, repositoryIDs: number[]): Promise<GitHubRepository[]> {
  const response = await request<{ repositories: GitHubRepository[] }>(
    `/api/auth/${accountID}/connections/${connectionID}/github/repositories`,
    { method: "PUT", body: JSON.stringify({ repository_ids: repositoryIDs }) },
  );
  return response.repositories;
}

export async function getSupabaseProjects(accountID: string, connectionID: string): Promise<SupabaseProject[]> {
  const response = await request<{ projects: SupabaseProject[] }>(
    `/api/auth/${accountID}/connections/${connectionID}/supabase/projects`,
    { method: "GET" },
  );
  return response.projects;
}

export async function updateSupabaseProjects(accountID: string, connectionID: string, projectRefs: string[]): Promise<SupabaseProject[]> {
  const response = await request<{ projects: SupabaseProject[] }>(
    `/api/auth/${accountID}/connections/${connectionID}/supabase/projects`,
    { method: "PUT", body: JSON.stringify({ project_refs: projectRefs }) },
  );
  return response.projects;
}

export async function getVercelProjects(accountID: string, connectionID: string): Promise<VercelProject[]> {
  const response = await request<{ projects: VercelProject[] }>(
    `/api/auth/${accountID}/connections/${connectionID}/vercel/projects`,
    { method: "GET" },
  );
  return response.projects;
}

export async function updateVercelProjects(accountID: string, connectionID: string, projectIDs: string[]): Promise<VercelProject[]> {
  const response = await request<{ projects: VercelProject[] }>(
    `/api/auth/${accountID}/connections/${connectionID}/vercel/projects`,
    { method: "PUT", body: JSON.stringify({ project_ids: projectIDs }) },
  );
  return response.projects;
}
