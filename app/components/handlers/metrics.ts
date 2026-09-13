import { request } from "./auth";
import type { WorkspaceMetrics, WorkspaceResourceTelemetry } from "../types/metrics";

type NullableArrays<T> = { [K in keyof T]: T[K] extends unknown[] ? T[K] | null : T[K] };
type MetricsResponse = Omit<NullableArrays<WorkspaceMetrics>, "resources"> & {
  resources: NullableArrays<WorkspaceResourceTelemetry>[] | null;
};

export async function getWorkspaceMetrics(accountID: string, workspaceID: string): Promise<WorkspaceMetrics> {
  const response = await request<MetricsResponse>(`/api/accounts/${accountID}/metrics/${workspaceID}`, { method: "GET" });
  return {
    ...response,
    providers: response.providers ?? [],
    resources: (response.resources ?? []).map((resource) => ({
      ...resource,
      service_ids: resource.service_ids ?? [],
      deployments: resource.deployments ?? [],
      supabase_services: resource.supabase_services ?? [],
      supabase_usage: resource.supabase_usage ?? [],
      warnings: resource.warnings ?? [],
    })),
    deployments: response.deployments ?? [],
    supabase_services: response.supabase_services ?? [],
    supabase_usage: response.supabase_usage ?? [],
    warnings: response.warnings ?? [],
  };
}
