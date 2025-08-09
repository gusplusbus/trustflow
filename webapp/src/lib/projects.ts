import { api } from "./api";
import type { ProjectFormValues } from "./validation";

export type ProjectResponse = {
  id: string;
  created_at: string;
  updated_at: string;
  // echo back fields we sent (optional)
  title: string;
  description: string;
  duration_estimate: number;
  team_size: number;
  application_close_time: string;
};

export async function createProject(values: ProjectFormValues) {
  // match your eventual API shape; adjust keys as your BE expects
  return api<ProjectResponse>("/api/projects", {
    method: "POST",
    body: JSON.stringify({
      title: values.title,
      description: values.description,
      duration_estimate: values.durationEstimate,
      team_size: values.teamSize,
      application_close_time: values.applicationCloseTime,
    }),
  });
}
