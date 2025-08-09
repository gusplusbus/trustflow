import { tokens, refreshAccessToken, logout } from "./auth";

// Generic fetch that:
// 1) Sends access token if present
// 2) If 401, tries to refresh (using refresh token) once, then retries the request
export async function api<T>(path: string, options: RequestInit = {}): Promise<T> {
  const doFetch = async (): Promise<Response> => {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...(options.headers as Record<string, string> | undefined),
    };
    if (tokens.access) headers.Authorization = `Bearer ${tokens.access}`;
    return fetch(path, { ...options, headers });
  };

  let res = await doFetch();

  if (res.status === 401 && tokens.refresh) {
    try {
      await refreshAccessToken();
      res = await doFetch(); // retry with new access token
    } catch {
      logout();
      throw new Error("Unauthorized");
    }
  }

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(text || res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}
