import { API_BASE, request } from "./auth";
import type { Connection, ConnectionCategory, ConnectionProvider, GitHubRepository } from "@/app/components/types/connections";
import type { SupabaseService, VercelService } from "@/app/components/types/services";

type ConnectionResponse = {
  connections: Array<Omit<Connection, "category" | "status"> & Partial<Pick<Connection, "category" | "status">>> | null;
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
  return (response.connections ?? []).map((connection) => ({
    ...connection,
    category: connection.category ?? providerCategories[connection.provider] ?? "account",
    status: connection.status ?? "active",
  }));
}

export async function deleteWorkspaceConnection(accountID: string, connectionID: string): Promise<void> {
  await request(`/api/auth/${accountID}/connections/${connectionID}`, { method: "DELETE" });
}

export async function beginWorkspaceOAuth(accountID: string, provider: ConnectionProvider, returnTo = ""): Promise<"redirected" | "completed"> {
  if (typeof window === "undefined") return "redirected";
  if (!API_BASE) throw new Error("API server is not configured.");

  const query = new URLSearchParams();
  if (returnTo) query.set("return_to", returnTo);
  const startURL = `${API_BASE}/api/auth/${accountID}/connections/oauth/${provider}/start${query.size ? `?${query}` : ""}`;
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
  const response = await request<{ repositories: GitHubRepository[] | null }>(`/api/auth/${accountID}/connections/${connectionID}/github/repositories`, { method: "GET" });
  return response.repositories ?? [];
}

export async function updateGitHubRepositories(accountID: string, connectionID: string, repositoryIDs: number[]): Promise<GitHubRepository[]> {
  const response = await request<{ repositories: GitHubRepository[] | null }>(`/api/auth/${accountID}/connections/${connectionID}/github/repositories`, {
    method: "PUT",
    body: JSON.stringify({ repository_ids: repositoryIDs }),
  });
  return response.repositories ?? [];
}

export async function getSupabaseServices(accountID: string, connectionID: string): Promise<SupabaseService[]> {
  const response = await request<{ services: SupabaseService[] | null }>(`/api/auth/${accountID}/connections/${connectionID}/supabase/services`, { method: "GET" });
  return response.services ?? [];
}

export async function updateSupabaseServices(accountID: string, connectionID: string, serviceRefs: string[]): Promise<SupabaseService[]> {
  const response = await request<{ services: SupabaseService[] | null }>(`/api/auth/${accountID}/connections/${connectionID}/supabase/services`, {
    method: "PUT",
    body: JSON.stringify({ service_refs: serviceRefs }),
  });
  return response.services ?? [];
}

export async function getVercelServices(accountID: string, connectionID: string): Promise<VercelService[]> {
  const response = await request<{ services: VercelService[] | null }>(`/api/auth/${accountID}/connections/${connectionID}/vercel/services`, { method: "GET" });
  return response.services ?? [];
}

export async function updateVercelServices(accountID: string, connectionID: string, serviceIDs: string[]): Promise<VercelService[]> {
  const response = await request<{ services: VercelService[] | null }>(`/api/auth/${accountID}/connections/${connectionID}/vercel/services`, {
    method: "PUT",
    body: JSON.stringify({ service_ids: serviceIDs }),
  });
  return response.services ?? [];
}
