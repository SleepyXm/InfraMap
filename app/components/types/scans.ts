import type { WorkspaceResource } from "./workspaces";

export type EZDeployComponent = {
  id: string;
  source_id?: string;
  kind: "backend" | "container" | string;
  name: string;
  root: string;
  runtime: string;
  entry: string;
  start_command?: string;
  confidence: string;
  evidence: string[];
  routes: string[];
  environment: string[];
  dockerfiles: string[];
  suggested_providers: string[];
};

export type EZDeployDependency = {
  id: string;
  kind: "database" | "cache" | "object_storage" | string;
  name: string;
  evidence: string[];
  suggested_providers: string[];
};

export type EZDeployAnalysis = {
  id: string;
  repository: string;
  branch: string;
  revision: string;
  analyzed_at: string;
  files_scanned: number;
  languages: Record<string, number>;
  components: EZDeployComponent[];
  dependencies: EZDeployDependency[];
  dockerfiles: Array<{ path: string; base_image: string; exposed_ports: number[] }>;
  warnings: string[];
};

export type EZDeployDecision = {
  component_id: string;
  action: "connect" | "deploy" | "existing" | "later";
  provider: "aws" | "vercel" | "supabase" | "external" | "";
};

export type EZDeployAnalysisResponse = {
  analysis: EZDeployAnalysis;
  decisions: EZDeployDecision[];
  source?: WorkspaceResource;
};
