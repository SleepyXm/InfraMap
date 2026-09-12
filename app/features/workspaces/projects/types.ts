export type WorkspaceProjectStatus = "draft" | "active" | "paused" | "archived";

export type WorkspaceProject = {
  id: string;
  account_id: string;
  name: string;
  slug: string;
  description: string;
  status: WorkspaceProjectStatus;
  created_by?: string;
  created_at: string;
  updated_at: string;
};

export type CreateWorkspaceProject = {
  name: string;
  description: string;
};

export type ProjectResource = {
  id: string;
  project_id: string;
  connection_id: string;
  provider: string;
  resource_type: string;
  external_id: string;
  name: string;
  status: string;
  metadata: Record<string, unknown>;
  created_by?: string;
  created_at: string;
  updated_at: string;
};

export type WorkspaceProjectDetail = {
  project: WorkspaceProject;
  resources: ProjectResource[];
  role: "owner" | "admin" | "member" | "viewer";
};

export type CreateSupabaseResource = {
  connection_id: string;
  name: string;
  organization_slug: string;
  region_group: "americas" | "emea" | "apac";
  database_password: string;
};

export type ProjectProviderStatus = {
  provider: "github" | "vercel" | "supabase" | "aws";
  name: string;
  status: string;
  detail: string;
  connected: boolean;
};

export type VercelDeployment = {
  uid: string;
  name: string;
  url: string;
  state: string;
  readyState: string;
  target: string;
  created: number;
  buildingAt: number;
  ready: number;
  creator: Record<string, unknown>;
  meta: Record<string, unknown>;
  inspectorUrl: string;
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

export type ProjectControl = {
  observed_at: string;
  status: "operational" | "partial" | "attention";
  providers: ProjectProviderStatus[];
  deployments: VercelDeployment[];
  supabase_services: SupabaseServiceHealth[];
  supabase_usage: SupabaseUsagePoint[];
  warnings: string[];
};
