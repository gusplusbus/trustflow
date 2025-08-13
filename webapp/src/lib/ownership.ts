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
