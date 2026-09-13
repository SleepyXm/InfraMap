export type CreateEZDeployDeployment = {
  domain: string;
  email: string;
  runtime: "native" | "docker";
};

export type EZDeployCommand = {
  command_id: string;
  status: string;
  status_detail?: string;
  response_code: number;
  standard_output?: string;
  standard_error?: string;
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
