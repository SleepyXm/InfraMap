export type ConnectionCategory =
  | "account"
  | "source"
  | "cloud"
  | "deployment"
  | "data"
  | "server";

export type ConnectionStatus = "pending" | "active" | "invalid" | "disabled";

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
