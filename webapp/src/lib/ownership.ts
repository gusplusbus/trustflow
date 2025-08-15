import { api } from "./api";

export type OwnershipFormValues = {
  organization: string;
  repository: string;
  provider?: string;
  web_url?: string;
};

export async function createOwnership(projectId: string, body: OwnershipFormValues) {
  return api<{ id: string }>(`/api/projects/${projectId}/ownership`, {
    method: "POST",
    body: JSON.stringify(body),
  });
}

export type OwnershipIssuesItem = {
  id: number;
  number: number;
  title: string;
  state: "open" | "closed";
  html_url: string;
  user_login?: string;
  labels?: string[];
  created_at: string;
  updated_at: string;
};

export type OwnershipIssuesResponse = {
  items: OwnershipIssuesItem[];
  total: number; // -1 when unknown
  rate?: { limit?: number; remaining?: number; reset?: number };
};

export type OwnershipIssuesQuery = {
  state?: "open" | "closed" | "all";
  labels?: string;
  assignee?: string;        // "", "*", or login
  since?: string;           // RFC3339
  per_page?: number;        // 1..100
  page?: number;            // default 1
  search?: string;          // optional
};

export async function listOwnershipIssues(
  projectId: string,
  q: OwnershipIssuesQuery
) {
  const params = new URLSearchParams();
  if (q.state) params.set("state", q.state);
  if (q.labels) params.set("labels", q.labels);
  if (q.assignee !== undefined) params.set("assignee", q.assignee);
  if (q.since) params.set("since", q.since);
  if (q.per_page) params.set("per_page", String(q.per_page));
  if (q.page) params.set("page", String(q.page));
  if (q.search) params.set("search", q.search);

  return api<OwnershipIssuesResponse>(
    `/api/projects/${projectId}/ownership/issues?${params.toString()}`
  );
}

export async function postOwnershipIssues(
  projectId: string,
  issues: { id: number; number: number }[]
) {
  return api(`/api/projects/${projectId}/issues`, {
    method: "POST",
    body: JSON.stringify({ issues }), // backend expects { issues: [...] }
  });
}

export type ImportedIssue = {
  id: string;                 // backend id (string/uuid)
  number: number;
  title: string;
  state: "open" | "closed";
  html_url: string;
  user_login?: string;
  labels?: string[];
  created_at: string;         // GitHub created
  updated_at: string;         // GitHub updated
};

export type ImportedIssuesResponse = {
  issues: ImportedIssue[];
};

export async function getImportedIssues(projectId: string) {
  return api<ImportedIssuesResponse>(`/api/projects/${projectId}/issues`, {
    method: "GET",
  });
}
