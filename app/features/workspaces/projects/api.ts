import { request } from "@/app/components/handlers/auth";
import type { CreateSupabaseResource, CreateWorkspaceProject, ProjectControl, ProjectResource, WorkspaceProject, WorkspaceProjectDetail } from "./types";

export async function getWorkspaceProjects(accountID: string): Promise<WorkspaceProject[]> {
  const response = await request<{ projects: WorkspaceProject[] }>(`/api/accounts/${accountID}/projects`, { method: "GET" });
  return response.projects;
}

export async function createWorkspaceProject(accountID: string, input: CreateWorkspaceProject): Promise<WorkspaceProject> {
  const response = await request<{ project: WorkspaceProject }>(`/api/accounts/${accountID}/projects`, {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.project;
}

export function getWorkspaceProject(accountID: string, projectID: string): Promise<WorkspaceProjectDetail> {
  return request<WorkspaceProjectDetail>(`/api/accounts/${accountID}/projects/${projectID}`, { method: "GET" });
}

export function getProjectControl(accountID: string, projectID: string): Promise<ProjectControl> {
  return request<ProjectControl>(`/api/accounts/${accountID}/projects/${projectID}/control`, { method: "GET" });
}

export async function attachGitHubResource(accountID: string, projectID: string, connectionID: string, repositoryID: number): Promise<ProjectResource> {
  const response = await request<{ resource: ProjectResource }>(`/api/accounts/${accountID}/projects/${projectID}/resources/github/attach`, {
    method: "POST",
    body: JSON.stringify({ connection_id: connectionID, repository_id: repositoryID }),
  });
  return response.resource;
}

export async function attachVercelResource(accountID: string, projectID: string, connectionID: string, vercelProjectID: string): Promise<ProjectResource> {
  const response = await request<{ resource: ProjectResource }>(`/api/accounts/${accountID}/projects/${projectID}/resources/vercel/attach`, {
    method: "POST",
    body: JSON.stringify({ connection_id: connectionID, vercel_project_id: vercelProjectID }),
  });
  return response.resource;
}

export async function attachSupabaseResource(accountID: string, projectID: string, connectionID: string, projectRef: string): Promise<ProjectResource> {
  const response = await request<{ resource: ProjectResource }>(`/api/accounts/${accountID}/projects/${projectID}/resources/supabase/attach`, {
    method: "POST",
    body: JSON.stringify({ connection_id: connectionID, project_ref: projectRef }),
  });
  return response.resource;
}

export async function createSupabaseResource(accountID: string, projectID: string, input: CreateSupabaseResource): Promise<ProjectResource> {
  const response = await request<{ resource: ProjectResource }>(`/api/accounts/${accountID}/projects/${projectID}/resources/supabase/create`, {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.resource;
}
