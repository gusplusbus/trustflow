import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  getProject, updateProject, deleteProject, createProject,
  type ProjectResponse,
  listProjects,
} from "../lib/projects";
import type { ProjectFormValues } from "../lib/projects";
import z from "zod";
import { useSearchParams } from "react-router-dom";

const SortByEnum = z.enum(["created_at", "updated_at", "title", "team_size", "duration"]);
const SortDirEnum = z.enum(["asc", "desc"]);

const ParamsSchema = z.object({
  page: z.preprocess(v => (v === null || v === undefined || v === "" ? undefined : Number(v)),
                     z.number().int().nonnegative().catch(0)).default(0),
  page_size: z.preprocess(v => (v === null || v === undefined || v === "" ? undefined : Number(v)),
                          z.number().int().min(1).max(200).catch(20)).default(20),
  sort_by: SortByEnum.catch("created_at").default("created_at"),
  sort_dir: SortDirEnum.catch("desc").default("desc"),
  q: z.string().catch("").default(""),
});

export type ListState = z.infer<typeof ParamsSchema>;
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



function toSearchParams(obj: ListState) {
  const p = new URLSearchParams();
  p.set("page", String(obj.page));
  p.set("page_size", String(obj.page_size));
  p.set("sort_by", obj.sort_by);
  p.set("sort_dir", obj.sort_dir);
  if (obj.q) p.set("q", obj.q); else p.delete("q");
  return p;
}

export function useListProjects() {
  const [sp, setSp] = useSearchParams();

  const state = useMemo(() => ParamsSchema.parse({
    page: sp.get("page"),
    page_size: sp.get("page_size"),
    sort_by: sp.get("sort_by"),
    sort_dir: sp.get("sort_dir"),
    q: sp.get("q"),
  }), [sp]);

  const setState = (next: Partial<ListState>) => {
    const merged = { ...state, ...next };
    setSp(toSearchParams(merged), { replace: false });
  };

  const [rows, setRows] = useState<ProjectResponse[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  useEffect(() => {
    const controller = new AbortController();
    abortRef.current?.abort();
    abortRef.current = controller;

    const run = async () => {
      setLoading(true);
      setError(null);
      try {
        const res = await listProjects(state, controller.signal);
        setRows(res.projects ?? []);
        setTotal(Number(res.total ?? 0));
      } catch (e: any) {
        if (e?.name !== "AbortError") setError(e?.message || "Failed to load projects");
      } finally {
        setLoading(false);
      }
    };
    run();
    return () => controller.abort();
  }, [state]);

  return {
    rows, total, loading, error,
    page: state.page,
    pageSize: state.page_size,
    sortBy: state.sort_by,
    sortDir: state.sort_dir,
    q: state.q,
    setPage: (page: number) => setState({ page }),
      setPageSize: (page_size: number) => setState({ page_size, page: 0 }),
      setSort: (sort_by: ListState["sort_by"], sort_dir: ListState["sort_dir"]) => setState({ sort_by, sort_dir, page: 0 }),
      setQ: (q: string) => setState({ q, page: 0 }),
  };
}
