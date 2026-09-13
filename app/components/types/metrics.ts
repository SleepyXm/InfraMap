import type { VercelDeployment } from "./deployments";

export type WorkspaceProviderStatus = {
  provider: "github" | "vercel" | "supabase" | "aws";
  name: string;
  status: string;
  detail: string;
  connected: boolean;
};

export type SupabaseServiceHealth = {
  name: string;
  healthy: boolean;
  status: string;
  info: Record<string, unknown>;
  error: string;
};

export type SupabaseUsagePoint = {
  timestamp: string;
  total_auth_requests: number;
  total_realtime_requests: number;
  total_rest_requests: number;
  total_storage_requests: number;
};

export type WorkspaceMetrics = {
  observed_at: string;
  status: "operational" | "partial" | "attention";
	providers: WorkspaceProviderStatus[];
	resources: WorkspaceResourceTelemetry[];
  deployments: VercelDeployment[];
  supabase_services: SupabaseServiceHealth[];
  supabase_usage: SupabaseUsagePoint[];
  warnings: string[];
};

export type WorkspaceResourceTelemetry = {
  resource_id: string;
  provider: string;
  name: string;
  resource_type: string;
  status: string;
  service_ids: string[];
  deployments: VercelDeployment[];
  supabase_services: SupabaseServiceHealth[];
  supabase_usage: SupabaseUsagePoint[];
  warnings: string[];
};
