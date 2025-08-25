import { api } from "./api";

export type ProjectWalletResponse = {
  address: `0x${string}`;
  chainId: number;
  updatedAt?: string;
};

export async function getProjectWallet(projectId: string, signal?: AbortSignal) {
  try {
    return await api<ProjectWalletResponse>(`/api/projects/${projectId}/wallet`, {
      method: "GET",
      signal,
    });
  } catch (e: any) {
    // mirror other libs: treat 404 as “no wallet yet”
    if (e?.status === 404) return null;
    throw e;
  }
}

export async function putProjectWallet(
  projectId: string,
  body: { address: `0x${string}`; chainId: number }
) {
  await api<void>(`/api/projects/${projectId}/wallet`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export async function deleteProjectWallet(projectId: string) {
  try {
    await api<void>(`/api/projects/${projectId}/wallet`, { method: "DELETE" });
  } catch (e: any) {
    if (e?.status === 404) return; // idempotent
    throw e;
  }
}
