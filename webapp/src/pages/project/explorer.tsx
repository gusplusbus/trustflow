import * as React from "react";
import { Button, Stack, Typography } from "@mui/material";
import { useNavigate } from "react-router-dom";
import { DataTable, type AllowedSortBy, type SortDir } from "../../components/DataTable";
import { useListProjects } from "../../hooks/project";
import type { GridColDef } from "@mui/x-data-grid";
import type { ProjectResponse } from "../../lib/projects";

export default function ProjectExplorer() {
  const nav = useNavigate();
  const {
    rows, total, loading, error,
    page, pageSize, sortBy, sortDir, q,
    setPage, setPageSize, setSort, setQ,
  } = useListProjects();

  const columns = React.useMemo<GridColDef<ProjectResponse>[]>(() => [
    { field: "title", headerName: "Title", flex: 1, minWidth: 200 },
    { field: "description", headerName: "Description", flex: 1.5, minWidth: 280 },
    { field: "team_size", headerName: "Team", type: "number", width: 110 },
    { field: "duration_estimate", headerName: "Duration (d)", type: "number", width: 140 },
    { field: "created_at", headerName: "Created", width: 160, valueFormatter: p => p.value ? new Date(String(p.value)).toLocaleString() : "" },
    { field: "updated_at", headerName: "Updated", width: 160, valueFormatter: p => p.value ? new Date(String(p.value)).toLocaleString() : "" },
    { field: "application_close_time", headerName: "Apply by", width: 180, valueFormatter: p => p.value ? new Date(String(p.value)).toLocaleString() : "" },
  ], []);

  return (
    <Stack spacing={2}>
      <Stack direction="row" alignItems="center" justifyContent="space-between">
        <Typography variant="h5">Projects</Typography>
        <Button variant="contained" onClick={() => nav("/projects/create")}>Create Project</Button>
      </Stack>

      <DataTable<ProjectResponse>
        rows={rows}
        rowCount={total}
        loading={loading}
        errorText={error}

        columns={columns}

        page={page}
        pageSize={pageSize}
        sortBy={sortBy as AllowedSortBy}
        sortDir={sortDir as SortDir}
        q={q}

        onPageChange={setPage}
        onPageSizeChange={setPageSize}
        onSortChange={(by, dir) => setSort(by, dir)}
        onSearchChange={setQ}

        onRowOpen={(id) => nav(`/projects/${id}`)}
        getRowId={(r) => r.id}
      />
    </Stack>
  );
}
