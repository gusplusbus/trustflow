import { api } from "./api";
import type { ProjectFormValues } from "./validation";
import { z } from "zod";

export type ListProjectsParams = {
  page: number;           // 0-based
  page_size: number;      // 1..200
  sort_by: "created_at" | "updated_at" | "title" | "team_size" | "duration";
  sort_dir: "asc" | "desc";
  q?: string;
};

export type ListProjectsResponse = {
  projects: ProjectResponse[];
  total: number;
  page: number;
  page_size: number;
  sort_by: string;
  sort_dir: string;
  q?: string;
};

export type ProjectResponse = {
  id: string;
  created_at: string;
  updated_at: string;
  title: string;
  description: string;
  duration_estimate: number;
  team_size: number;
  application_close_time: string;
};

export const projectSchema = z.object({
  title: z.string().min(1, "Title is required").max(84, "Max 84 characters"),
  description: z
    .string()
    .min(1, "Description is required")
    .max(221, "Max 221 characters"),
  durationEstimate: z
    .number({ invalid_type_error: "Enter a number" })
    .int("Must be an integer")
    .positive("Must be > 0"),
  teamSize: z
    .number({ invalid_type_error: "Enter a number" })
    .int("Must be an integer")
    .min(1, "Minimum team size is 1"),
  applicationCloseTime: z
    .string()
    .min(1, "Close time is required"),
});

export type ProjectFormValues = z.infer<typeof projectSchema>;

function toDTO(values: Partial<ProjectFormValues>) {
  const body: any = {};
  if (values.title !== undefined) body.title = values.title;
  if (values.description !== undefined) body.description = values.description;
  if (values.durationEstimate !== undefined) body.duration_estimate = values.durationEstimate;
  if (values.teamSize !== undefined) body.team_size = values.teamSize;
  if (values.applicationCloseTime !== undefined) body.application_close_time = values.applicationCloseTime;
  return body;
}

export async function createProject(values: ProjectFormValues) {
  return api<ProjectResponse>("/api/projects", {
    method: "POST",
    body: JSON.stringify(toDTO(values)),
  });
}

export async function getProject(id: string) {
  return api<ProjectResponse>(`/api/projects/${id}`);
}

export async function updateProject(id: string, values: Partial<ProjectFormValues>) {
  return api<ProjectResponse>(`/api/projects/${id}`, {
    method: "PUT",
    body: JSON.stringify(toDTO(values)),
  });
}

export async function deleteProject(id: string) {
  return api<void>(`/api/projects/${id}`, { method: "DELETE" });
}

export async function listProjects(params: ListProjectsParams, signal?: AbortSignal): Promise<ListProjectsResponse> {
  const qs = new URLSearchParams();
  qs.set("page", String(params.page));
  qs.set("page_size", String(params.page_size));
  qs.set("sort_by", params.sort_by);
  qs.set("sort_dir", params.sort_dir);
  if (params.q) qs.set("q", params.q);

  return api<ListProjectsResponse>(`/api/projects?${qs.toString()}`, {
    method: "GET",
    headers: { Accept: "application/json" },
    signal,
  });
}

