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
