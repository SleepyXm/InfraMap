import type { OAuthIdentityProvider, User } from "@/app/components/types/users";

export const API_BASE = process.env.NEXT_PUBLIC_API_BASE;
export const WSAPI_BASE = process.env.NEXT_PUBLIC_WS_API_BASE;

let refreshPromise: Promise<boolean> | null = null;

export class ApiError extends Error {
  constructor(
    message: string,
    public readonly status: number,
    public readonly payload: Record<string, unknown>,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

function apiUrl(path: string): string {
  if (!API_BASE) throw new Error("API server is not configured.");
  return `${API_BASE}${path}`;
}

function asPayload(value: unknown): Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value)
    ? value as Record<string, unknown>
    : {};
}

async function responseBody(response: Response): Promise<unknown> {
  const text = await response.text();
  if (!text) return undefined;
  try {
    return JSON.parse(text) as unknown;
  } catch {
    if (!response.ok) return { error: text };
    throw new Error("The server returned an invalid response.");
  }
}

export async function request<T = unknown>(path: string, options: RequestInit, isRetry = false): Promise<T> {
  const isFormData = typeof FormData !== "undefined" && options.body instanceof FormData;
  const res = await fetch(apiUrl(path), {
    ...options,
    credentials: "include",
    headers: isFormData
      ? options.headers
      : { "Content-Type": "application/json", ...(options.headers || {}) },
  });

  const skipRefresh = ["/auth/login", "/auth/signup", "/auth/refresh", "/auth/logout"];
  if (res.status === 401 && !isRetry && !skipRefresh.some(p => path.includes(p))) {
    if (!refreshPromise) {
      refreshPromise = fetch(apiUrl("/api/auth/refresh"), {
        method: "POST",
        credentials: "include",
      })
        .then((r) => r.ok)
        .finally(() => { refreshPromise = null; });
    }
    const refreshed = await refreshPromise;
    if (refreshed) return request<T>(path, options, true);

    throw new Error("UNAUTHENTICATED");
  }

  const data = await responseBody(res);
  const payload = asPayload(data);
  if (!res.ok) {
    const message = typeof payload.detail === "string"
      ? payload.detail
      : typeof payload.error === "string"
        ? payload.error
        : `Request failed with status ${res.status}`;
    throw new ApiError(message, res.status, payload);
  }
  return data as T;
}

type AuthMessageResponse = {
  message: string;
};

const STALE_TIME = 5 * 60 * 1000;

function cachedValue<T>(key: string): T | null {
  const value = localStorage.getItem(key);
  if (!value) return null;
  try {
    return JSON.parse(value) as T;
  } catch {
    localStorage.removeItem(key);
    return null;
  }
}

export async function validateUser(force = false): Promise<{ user: User } | null> {
  try {
    const lastCheck = localStorage.getItem("user_validated_at");
    const parsedUser = cachedValue<User>("user");

    if (
      lastCheck &&
      parsedUser &&
      Date.now() - Number(lastCheck) < STALE_TIME && !force
    ) {
      return {
        user: parsedUser,
      };
    }

    const { user } = await request<{ user: User }>("/api/auth/me", { method: "GET" });
    
    localStorage.setItem("user", JSON.stringify(user));
    localStorage.setItem("user_validated_at", String(Date.now()));

    return { user };
  } catch {
    return null;
  }
}


export async function signup(username: string, email: string, password: string): Promise<AuthMessageResponse> {
  return request<AuthMessageResponse>("/api/auth/signup", {
    method: "POST",
    body: JSON.stringify({ username, email, password }),
  });
}

export async function login(email: string, password: string): Promise<{ user: User; } | null> {
  await request<AuthMessageResponse>("/api/auth/login", { method: "POST", body: JSON.stringify({ email, password }) });
  ["user", "user_validated_at"].forEach(k => localStorage.removeItem(k));
  return validateUser();
}

export function beginOAuth(provider: OAuthIdentityProvider, intent: "login" | "connect" = "login"): void {
  if (typeof window === "undefined") return;
  window.location.assign(apiUrl(`/api/auth/oauth/${provider}/start${intent === "connect" ? "?intent=connect" : ""}`));
}

export async function completeOAuthCallback(force = false): Promise<User> {
  const result = await validateUser(force);
  if (!result) throw new Error("Authentication could not be completed.");
  return result.user;
}


export async function logout(): Promise<void> {
  try {
    await request<AuthMessageResponse>("/api/auth/logout", { method: "POST" });
  } finally {
    ["user", "user_validated_at"].forEach(k => localStorage.removeItem(k));
  }
}
