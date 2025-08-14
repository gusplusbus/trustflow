import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  getProject, updateProject, deleteProject, createProject,
  type ProjectResponse,
  listProjects,
} from "../lib/projects";
import type { ProjectFormValues } from "../lib/projects";
import z from "zod";
import { useSearchParams } from "react-router-dom";
import { listOwnershipIssues, type OwnershipIssuesQuery, type OwnershipIssuesResponse } from "../lib/ownership";

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
        const res = await getProject(id, { includeOwnerships: true });
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

  // ---- read initial state from URL ONCE
  const readParams = () => {
    const page = Math.max(0, Number(sp.get("page") ?? 0) || 0);
    const page_size = Math.min(200, Math.max(1, Number(sp.get("page_size") ?? 20) || 20));
    const sort_by = (sp.get("sort_by") as ListState["sort_by"]) || "created_at";
    const sort_dir = (sp.get("sort_dir") as ListState["sort_dir"]) || "desc";
    const q = sp.get("q") || "";
    return { page, page_size, sort_by, sort_dir, q } as ListState;
  };

  const [state, setLocalState] = useState<ListState>(() => readParams());

  // guard to avoid feedback loops when *we* push to URL
  const pushingRef = useRef(false);

  // ---- keep URL in sync when local state changes (but don't re-derive local state from URL)
  const writeParams = (obj: ListState) => {
    const p = new URLSearchParams();
    p.set("page", String(obj.page));
    p.set("page_size", String(obj.page_size));
    p.set("sort_by", obj.sort_by);
    p.set("sort_dir", obj.sort_dir);
    if (obj.q) p.set("q", obj.q);
    return p;
  };

  const setState = (next: Partial<ListState>) => {
    setLocalState(prev => {
      const merged = { ...prev, ...next };
      pushingRef.current = true;
      setSp(writeParams(merged), { replace: false });
      // release the flag after this tick
      queueMicrotask(() => { pushingRef.current = false; });
      return merged;
    });
  };

  // ---- allow back/forward navigation to update local state
  useEffect(() => {
    if (pushingRef.current) return; // ignore our own push
    const fromUrl = readParams();
    setLocalState(prev => {
      // only update if something truly changed
      if (
        prev.page !== fromUrl.page ||
        prev.page_size !== fromUrl.page_size ||
        prev.sort_by !== fromUrl.sort_by ||
        prev.sort_dir !== fromUrl.sort_dir ||
        prev.q !== fromUrl.q
      ) {
        return fromUrl;
      }
      return prev;
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sp.toString()]);

  // ---- data fetching
  const [rows, setRows] = useState<ProjectResponse[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  // remember the "base" page size (the one user chose / coming from URL)
  const basePageSizeRef = useRef<number>(state.page_size);

  // whenever sort/search changes, reset the base page size from URL/state
  useEffect(() => {
    basePageSizeRef.current = state.page_size;
  }, [state.sort_by, state.sort_dir, state.q]);

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
  }, [state.page, state.page_size, state.sort_by, state.sort_dir, state.q]);

  // ---- same calc, unchanged
  const calcPageAndSize = (nextPage: number) => {
    const base = Math.max(1, basePageSizeRef.current || 1);
    const totalItems = Math.max(0, total);

    if (totalItems === 0) {
      return { page: Math.max(0, nextPage), page_size: base };
    }
    const pageCount = Math.max(1, Math.ceil(totalItems / base));
    const page = Math.min(Math.max(0, nextPage), pageCount - 1);
    return { page, page_size: base };
  };

  return {
    rows,
    total,
    loading,
    error,
    page: state.page,
    pageSize: state.page_size,
    sortBy: state.sort_by,
    sortDir: state.sort_dir,
    q: state.q,

    setPage: (nextPage: number) => {
      const { page, page_size } = calcPageAndSize(nextPage);
      setState({ page, page_size });
    },

    setPageSize: (page_size: number) => {
      basePageSizeRef.current = page_size;
      setState({ page: 0, page_size });
    },

    setSort: (sort_by: ListState["sort_by"], sort_dir: ListState["sort_dir"]) => {
      basePageSizeRef.current = state.page_size;
      const { page, page_size } = calcPageAndSize(0);
      setState({ sort_by, sort_dir, page, page_size });
    },

    setQ: (q: string) => {
      basePageSizeRef.current = state.page_size;
      const { page, page_size } = calcPageAndSize(0);
      setState({ q, page, page_size });
    },
  };
}

export function useOwnershipIssuesLoader(projectId?: string) {
  return useCallback(
    async (q: OwnershipIssuesQuery): Promise<OwnershipIssuesResponse> => {
      if (!projectId) throw new Error("project id required");
      return listOwnershipIssues(projectId, q);
    },
    [projectId]
  );
}

export type GitHubIssue = {
  id: number;
  number: number;
  title: string;
  state: "open" | "closed";
  html_url: string;
  user?: { login?: string };
  created_at: string;
  updated_at: string;
  labels?: Array<{ name?: string }>;
};

export type ImportFilters = {
  state: "open" | "closed" | "all";
  labels: string;
  assignee: string;
  since: string;
  per_page: number;
  search: string;
};

export const defaultFilters: ImportFilters = {
  state: "open",
  labels: "",
  assignee: "",
  since: "",
  per_page: 50,
  search: "",
};

type UseOwnershipIssuesArgs = {
  projectId: string;
};

export function useOwnershipIssues({
  projectId,
}: UseOwnershipIssuesArgs) {
  const [filters, setFilters] = React.useState<ImportFilters>({
    ...defaultFilters,
  });
  const [issues, setIssues] = React.useState<GitHubIssue[]>([]);
  const [loading, setLoading] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [rateRemaining, setRateRemaining] = React.useState<string | null>(null);

  const listIssues = React.useCallback(async () => {
    setError(null);
    setLoading(true);
    try {
      const sinceIso = filters.since ? new Date(filters.since).toISOString() : undefined;
      const json = await listOwnershipIssues(projectId, {
        state: filters.state,
        labels: filters.labels || undefined,
        assignee: filters.assignee,
        since: sinceIso,
        per_page: Math.max(1, Math.min(100, filters.per_page || 50)),
        page: 1,
        search: filters.search.trim() || undefined,
      });
      setIssues(
        (json.items || []).map(i => ({
          id: i.id,
          number: i.number,
          title: i.title,
          state: i.state,
          html_url: i.html_url,
          user: i.user_login ? { login: i.user_login } : undefined,
          labels: (i.labels || []).map(name => ({ name })),
          created_at: i.created_at,
          updated_at: i.updated_at,
        }))
      );
      setRateRemaining(json.rate?.remaining != null ? String(json.rate.remaining) : null);
    } catch (e: any) {
      setError(e?.message || "Failed to load issues");
    } finally {
      setLoading(false);
    }
  }, [projectId, filters]);

  return { filters, setFilters, issues, loading, error, rateRemaining, listIssues };
}
