import { api } from "./api";
import type { ProjectFormValues } from "./validation";
import { z } from "zod";

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

export async function listProjects() {
  return api<ProjectResponse[]>("/api/projects");
}
