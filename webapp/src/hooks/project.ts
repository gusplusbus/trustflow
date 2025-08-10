import { useCallback, useEffect, useState } from "react";
import {
  getProject, updateProject, deleteProject, createProject,
  type ProjectResponse,
} from "../lib/projects";
import type { ProjectFormValues } from "../lib/validation";

export function useProject(id?: string) {
  const [data, setData] = useState<ProjectResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState<boolean>(!!id); // ← start true if we have an id

  const load = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setError(null);
    try {
      const res = await getProject(id);
      setData(res);
    } catch (e: any) {
      setData(null); // ← make state explicit on failure
      setError(e?.message || "Failed to load project");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => { load(); }, [load]);
  return { data, error, loading, reload: load };
}

export function useCreateProject() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<ProjectResponse | null>(null);

  const mutate = useCallback(async (values: ProjectFormValues) => {
    setLoading(true);
    setError(null);
    try {
      const res = await createProject(values);
      setResult(res);
      return res;
    } catch (e: any) {
      setError(e?.message || "Failed to create project");
      throw e;
    } finally {
      setLoading(false);
    }
  }, []);

  return { create: mutate, loading, error, result };
}

export function useUpdateProject(id: string) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mutate = useCallback(async (values: Partial<ProjectFormValues>) => {
    setLoading(true);
    setError(null);
    try {
      return await updateProject(id, values);
    } catch (e: any) {
      setError(e?.message || "Failed to update project");
      throw e;
    } finally {
      setLoading(false);
    }
  }, [id]);

  return { update: mutate, loading, error };
}

export function useDeleteProject(id: string) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const mutate = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      await deleteProject(id);
    } catch (e: any) {
      setError(e?.message || "Failed to delete project");
      throw e;
    } finally {
      setLoading(false);
    }
  }, [id]);

  return { remove: mutate, loading, error };
}
