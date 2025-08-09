export type User = { id: string; email: string };
export type LoginResponse = {
  token: string;
  refresh_token: string;
  id?: string; email?: string; // backend returns user fields; optional here
};

const ACCESS_KEY = "auth:access";
const REFRESH_KEY = "auth:refresh";

export const tokens = {
  get access() { return localStorage.getItem(ACCESS_KEY) || ""; },
  set access(v: string) { localStorage.setItem(ACCESS_KEY, v); },
  get refresh() { return localStorage.getItem(REFRESH_KEY) || ""; },
  set refresh(v: string) { localStorage.setItem(REFRESH_KEY, v); },
  clear() { localStorage.removeItem(ACCESS_KEY); localStorage.removeItem(REFRESH_KEY); },
};

export function isAuthed() {
  return !!tokens.access;
}

export async function login(email: string, password: string) {
  const res = await fetch("/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) throw new Error(await res.text());
  const data: LoginResponse = await res.json();
  tokens.access = data.token;
  if (data.refresh_token) tokens.refresh = data.refresh_token;
  return data;
}

export async function register(email: string, password: string) {
  const res = await fetch("/auth/users", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}

export async function refreshAccessToken() {
  if (!tokens.refresh) throw new Error("No refresh token");
  const res = await fetch("/auth/refresh", {
    method: "POST",
    // backend expects refresh token in Authorization header:
    headers: { Authorization: `Bearer ${tokens.refresh}` },
  });
  if (!res.ok) throw new Error(await res.text());
  const data = (await res.json()) as { token: string };
  tokens.access = data.token;
  return data.token;
}

export async function revokeRefreshToken() {
  if (!tokens.refresh) return;
  await fetch("/auth/revoke", {
    method: "POST",
    headers: { Authorization: `Bearer ${tokens.refresh}` },
  }).catch(() => {});
  tokens.clear();
}

export function logout() {
  tokens.clear();
}
