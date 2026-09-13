import { request } from "./auth";
import type { AttachAWSHostInput, AWSCloudFormationBundle, CreateSupabaseResource } from "@/app/components/types/services";
import type { WorkspaceResource } from "@/app/components/types/workspaces";

export async function attachVercelService(accountID: string, workspaceID: string, connectionID: string, serviceID: string): Promise<WorkspaceResource> {
  const response = await request<{ service: WorkspaceResource }>(`/api/accounts/${accountID}/services/${workspaceID}/vercel`, {
    method: "POST",
    body: JSON.stringify({ connection_id: connectionID, vercel_service_id: serviceID }),
  });
  return response.service;
}

export async function attachSupabaseService(accountID: string, workspaceID: string, connectionID: string, serviceRef: string): Promise<WorkspaceResource> {
  const response = await request<{ service: WorkspaceResource }>(`/api/accounts/${accountID}/services/${workspaceID}/supabase`, {
    method: "POST",
    body: JSON.stringify({ connection_id: connectionID, service_ref: serviceRef }),
  });
  return response.service;
}

export async function createSupabaseService(accountID: string, workspaceID: string, input: CreateSupabaseResource): Promise<WorkspaceResource> {
  const response = await request<{ service: WorkspaceResource }>(`/api/accounts/${accountID}/services/${workspaceID}/supabase/provision`, {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.service;
}

export function getAWSCloudFormation(accountID: string, workspaceID: string): Promise<AWSCloudFormationBundle> {
  return request<AWSCloudFormationBundle>(`/api/accounts/${accountID}/services/${workspaceID}/aws/cloudformation`, { method: "GET" });
}

export async function attachAWSService(accountID: string, workspaceID: string, input: AttachAWSHostInput): Promise<WorkspaceResource> {
  const response = await request<{ service: WorkspaceResource }>(`/api/accounts/${accountID}/services/${workspaceID}/aws`, {
    method: "POST",
    body: JSON.stringify(input),
  });
  return response.service;
}
