import * as React from "react";
import { Button, Stack, Typography, Tooltip, IconButton } from "@mui/material";
import VisibilityOutlinedIcon from "@mui/icons-material/VisibilityOutlined";
import { Link, useNavigate } from "react-router-dom";
import { DataTable, type AllowedSortBy, type SortDir } from "../../components/DataTable";
import type { GridColDef } from "@mui/x-data-grid";
import type { ProjectResponse } from "../../lib/projects";
import { useListProjects } from "../../hooks/project";

const toLocal = (v: unknown) => (v ? new Date(String(v)).toLocaleString() : "");

export default function ProjectExplorer() {
  const nav = useNavigate();
  const {
    rows, total, loading, error,
    page, pageSize, sortBy, sortDir, q,
    setPage, setPageSize, setSort, setQ,
  } = useListProjects();

  const columns = React.useMemo<GridColDef<ProjectResponse>[]>(() => [
    // Title becomes a link
    {
      field: "title",
      headerName: "Title",
      flex: 1,
      minWidth: 200,
      renderCell: (p) => (
        <Link
          to={`/projects/${(p.row as any).id}`}
          onClick={(e) => e.stopPropagation()}
          style={{ textDecoration: "none", color: "inherit", fontWeight: 600 }}
        >
          {p.row.title}
        </Link>
      ),
    },

    { field: "description", headerName: "Description", flex: 1.5, minWidth: 280, sortable: false },
    { field: "team_size", headerName: "Team", type: "number", width: 110 },
    { field: "duration_estimate", headerName: "Duration (d)", type: "number", width: 140 },
    {
      field: "created_at",
      headerName: "Created",
      width: 180,
      renderCell: (p) => toLocal((p.row as any)?.created_at),
    },
    {
      field: "updated_at",
      headerName: "Updated",
      width: 180,
      renderCell: (p) => toLocal((p.row as any)?.updated_at),
    },
    {
      field: "application_close_time",
      headerName: "Apply by",
      width: 200,
      renderCell: (p) => toLocal((p.row as any)?.application_close_time),
    },
  ], []);

  return (
    <Stack spacing={2}>
      <Stack direction="row" alignItems="center" justifyContent="space-between">
        <Typography variant="h5">Projects</Typography>
        <Button variant="contained" onClick={() => nav("/projects/create")}>
          Create Project
        </Button>
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
