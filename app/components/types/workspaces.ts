import type { EZDeployComponent } from "./scans";

export type WorkspaceStatus = "draft" | "active" | "paused" | "archived";

export type Workspace = {
  id: string;
  account_id: string;
  name: string;
  slug: string;
  description: string;
  status: WorkspaceStatus;
  created_by?: string;
  created_at: string;
  updated_at: string;
};

export type CreateWorkspace = {
  name: string;
  description: string;
};

export type WorkspaceResource = {
  id: string;
  workspace_id: string;
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

export type WorkspaceDetail = {
  workspace: Workspace;
  sources: WorkspaceResource[];
  components: EZDeployComponent[];
  environments: WorkspaceEnvironment[];
  services: WorkspaceResource[];
  access: "owner" | "admin" | "member" | "viewer";
};

export type WorkspaceEnvironment = {
  id: string;
  name: string;
  status: string;
};
