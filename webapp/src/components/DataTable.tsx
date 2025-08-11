import * as React from "react";
import { Alert, Paper } from "@mui/material";
// values/


import { DataGrid, GridToolbar } from "@mui/x-data-grid";
// types (type-only!)
import type {
  GridColDef,
  GridPaginationModel,
  GridSortModel,
  GridFilterModel,
} from "@mui/x-data-grid";export type SortDir = "asc" | "desc";
export type AllowedSortBy = "created_at" | "updated_at" | "title" | "team_size" | "duration";

type Props<Row> = {
  rows: Row[];
  rowCount: number;
  loading?: boolean;
  errorText?: string | null;

  columns: GridColDef<Row>[];

  page: number;
  pageSize: number;
  sortBy: AllowedSortBy;
  sortDir: SortDir;
  q: string;

  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onSortChange: (by: AllowedSortBy, dir: SortDir) => void;
  onSearchChange: (q: string) => void;

  onRowOpen?: (id: string) => void;
  getRowId: (row: Row) => string;
};

export function DataTable<Row>(props: Props<Row>) {
  const {
    rows, rowCount, loading, errorText,
    columns, page, pageSize, sortBy, sortDir, q,
    onPageChange, onPageSizeChange, onSortChange, onSearchChange,
    onRowOpen, getRowId,
  } = props;

  // Mirror quick filter input locally, debounce before sending up
  const [quick, setQuick] = React.useState(q);
  React.useEffect(() => setQuick(q), [q]);
  React.useEffect(() => {
    const t = setTimeout(() => {
      if (quick !== q) onSearchChange(quick);
    }, 500);
    return () => clearTimeout(t);
  }, [quick, q, onSearchChange]);

  const paginationModel: GridPaginationModel = { page, pageSize };
  const sortModel: GridSortModel = [{ field: sortBy, sort: sortDir }];

  const handlePagination = (m: GridPaginationModel) => {
    if (m.page !== page) onPageChange(m.page);
    if (m.pageSize !== pageSize) onPageSizeChange(m.pageSize);
  };

  const handleSort = (m: GridSortModel) => {
    const s = m[0];
    if (!s?.field || !s.sort) return;
    const allowed: Record<string, true> = {
      created_at: true,
      updated_at: true,
      title: true,
      team_size: true,
      duration: true,
    };
    const field = allowed[s.field] ? (s.field as AllowedSortBy) : sortBy;
    onSortChange(field, s.sort as SortDir);
  };

  const handleFilter = (m: GridFilterModel) => {
    const v = (m.quickFilterValues?.[0] ?? "").toString();
    setQuick(v);
  };

  return (
    <Paper sx={{ height: 560, width: "100%" }}>
      {errorText && <Alert severity="error" sx={{ m: 1 }}>{errorText}</Alert>}
      <DataGrid
        getRowId={getRowId}
        rows={rows}
        rowCount={rowCount}
        loading={!!loading}
        columns={columns}
        paginationMode="server"
        sortingMode="server"
        filterMode="server"
        paginationModel={paginationModel}
        onPaginationModelChange={handlePagination}
        sortModel={sortModel}
        onSortModelChange={handleSort}
        onFilterModelChange={handleFilter}
        disableColumnFilter
        disableDensitySelector
        slots={{ toolbar: GridToolbar }}
        slotProps={{ toolbar: { showQuickFilter: true, quickFilterProps: { debounceMs: 0 } } }}
        onRowDoubleClick={(p) => onRowOpen?.(String(getRowId(p.row)))}
      />
    </Paper>
  );
}
