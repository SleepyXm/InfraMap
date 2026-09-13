import { request } from "./auth";
import { normalizeEZDeployAnalysis } from "@/app/components/handlers/scans";
import { type CreateWorkspace, type Workspace, type WorkspaceDetail, type WorkspaceResource } from "@/app/components/types/workspaces";

export async function getWorkspaces(accountID: string): Promise<Workspace[]> {
  const response = await request<{ workspaces: Workspace[] | null }>(`/api/accounts/${accountID}/workspaces`, { method: "GET" });
  return response.workspaces ?? [];
}

export async function createWorkspace(accountID: string, input: CreateWorkspace): Promise<Workspace> {
  const response = await request<{ workspace: Workspace }>(`/api/accounts/${accountID}/workspaces`, { method: "POST", body: JSON.stringify(input) });
  return response.workspace;
}

export async function deleteWorkspace(accountID: string, workspaceID: string): Promise<void> {
  await request(`/api/accounts/${accountID}/workspaces/${workspaceID}`, { method: "DELETE" });
}

export async function getWorkspace(accountID: string, workspaceID: string): Promise<WorkspaceDetail> {
  const [workspaceResponse, sourceResponse, serviceResponse] = await Promise.all([
    request<{ workspace: Workspace; access: WorkspaceDetail["access"] }>(`/api/accounts/${accountID}/workspaces/${workspaceID}`, { method: "GET" }),
    request<{ sources: WorkspaceResource[] }>(`/api/accounts/${accountID}/sources/${workspaceID}`, { method: "GET" }),
    request<{ services: WorkspaceResource[] }>(`/api/accounts/${accountID}/services/${workspaceID}`, { method: "GET" }),
  ]);
  const sources = (sourceResponse.sources ?? []).map((source) => ({ ...source, metadata: source.metadata ?? {} }));
  const services = (serviceResponse.services ?? []).map((service) => ({ ...service, metadata: service.metadata ?? {} }));
  const components = sources.flatMap((source) => normalizeEZDeployAnalysis(source.metadata.ezdeploy_analysis)?.components.map((component) => ({ ...component, source_id: source.id })) ?? []);
  return { ...workspaceResponse, sources, services, components, environments: [] };
}

export async function attachGitHubSource(accountID: string, workspaceID: string, connectionID: string, repositoryID: number): Promise<WorkspaceResource> {
  const response = await request<{ source: WorkspaceResource }>(`/api/accounts/${accountID}/sources/${workspaceID}/github`, {
    method: "POST",
    body: JSON.stringify({ connection_id: connectionID, repository_id: repositoryID }),
  });
  return response.source;
}
