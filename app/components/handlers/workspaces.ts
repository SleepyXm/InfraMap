import { API_BASE, request } from "@/app/components/handlers/auth";
import type { Connection, ConnectionCategory, ConnectionProvider, GitHubRepository, SupabaseProject } from "@/app/components/types/users";

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

export function beginWorkspaceOAuth(accountID: string, provider: ConnectionProvider): void {
  if (typeof window === "undefined") return;
  if (!API_BASE) throw new Error("API server is not configured.");

  window.location.assign(`${API_BASE}/api/auth/${accountID}/connections/oauth/${provider}/start`);
}

export async function getGitHubRepositories(accountID: string, connectionID: string): Promise<GitHubRepository[]> {
  const response = await request<{ repositories: GitHubRepository[] }>(
    `/api/auth/${accountID}/connections/${connectionID}/github/repositories`,
    { method: "GET" },
  );
  return response.repositories;
}

export async function updateGitHubRepositories(
  accountID: string,
  connectionID: string,
  repositoryIDs: number[],
): Promise<GitHubRepository[]> {
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

export async function updateSupabaseProjects(
  accountID: string,
  connectionID: string,
  projectRefs: string[],
): Promise<SupabaseProject[]> {
  const response = await request<{ projects: SupabaseProject[] }>(
    `/api/auth/${accountID}/connections/${connectionID}/supabase/projects`,
    { method: "PUT", body: JSON.stringify({ project_refs: projectRefs }) },
  );
  return response.projects;
}
