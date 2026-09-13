export type SupabaseService = {
  id: string;
  ref: string;
  organization_id: string;
  organization_slug: string;
  name: string;
  region: string;
  created_at: string;
  status: string;
  database?: {
    host: string;
    version: string;
    postgres_engine: string;
    release_channel: string;
  };
  selected: boolean;
};

export type VercelService = {
  id: string;
  name: string;
  framework: string;
  accountId: string;
  createdAt: number;
  updatedAt: number;
  selected: boolean;
};

export type CreateSupabaseResource = {
  connection_id: string;
  name: string;
  organization_slug: string;
  region_group: "americas" | "emea" | "apac";
  database_password: string;
};

export type AWSCloudFormationBundle = {
  template: string;
  filename: string;
  external_id: string;
  stack_name: string;
};

export type AttachAWSHostInput = {
  name: string;
  role_arn: string;
  external_id: string;
  region: string;
  instance_id: string;
  document_name: string;
  stack_id: string;
  public_ip: string;
};
