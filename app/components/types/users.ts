export type User = {
  id: string;
  username: string;
  email: string;
  verified: boolean;
  identities?: ConnectedIdentity[];
  accounts?: Account[];
};

// OAuthIdentityProvider identifies providers that can authenticate a user.
export type OAuthIdentityProvider = "github" | "google";

// ConnectedIdentity represents an external identity linked directly to a user.
export type ConnectedIdentity = {
  id: string;
  provider: OAuthIdentityProvider;
  provider_id: string;
  email: string;
  username: string;
  connected_at: string;
  primary: boolean;
};

// AccountRole identifies the user's permission level within an account.
export type AccountRole = "owner" | "admin" | "member" | "viewer";

// Account represents an InfraMap account the authenticated user belongs to.
export type Account = {
  id: string;
  name: string;
  slug: string;
  type: "personal" | "organization";
  role: AccountRole;
};

// ConnectionCategory identifies the functional category a connection belongs to.
export type ConnectionCategory =
  | "account"
  | "source"
  | "cloud"
  | "deployment"
  | "data"
  | "server";

// ConnectionStatus identifies the current health or usability state of a connection.
export type ConnectionStatus = "pending" | "active" | "invalid" | "disabled";

// ConnectionProvider identifies a supported external infrastructure or platform provider.
export type ConnectionProvider =
  | "github"
  | "gitlab"
  | "bitbucket"
  | "aws"
  | "gcp"
  | "azure"
  | "cloudflare"
  | "vercel"
  | "railway"
  | "render"
  | "netlify"
  | "supabase"
  | "neon"
  | "planetscale"
  | "digitalocean"
  | "hetzner"
  | "kubernetes"
  | "docker"
  | "ssh"
  | "generic";

// Connection represents the base connection returned by the backend.
export type Connection = {
  id: string;
  account_id: string;
  name: string;
  category: ConnectionCategory;
  provider: ConnectionProvider;
  connection_type: string;
  external_account_id?: string;
  status: ConnectionStatus;
  metadata?: Record<string, unknown>;
  created_by?: string;
  created_at: string;
  updated_at: string;
  last_verified_at?: string;
};

export type GitHubRepository = {
  id: number;
  name: string;
  full_name: string;
  private: boolean;
  description: string;
  default_branch: string;
  language: string;
  html_url: string;
  pushed_at: string;
  selected: boolean;
};

export type SupabaseProject = {
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

export type VercelProject = {
  id: string;
  name: string;
  framework: string;
  accountId: string;
  createdAt: number;
  updatedAt: number;
  selected: boolean;
};
