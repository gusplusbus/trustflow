import * as React from "react";
import { Alert, Paper } from "@mui/material";
import {
  DataGrid,
  GridToolbar,
  type GridColDef,
  type GridPaginationModel,
  type GridSortModel,
  type GridFilterModel,
  type GridValidRowModel,
  type GridRowId,
} from "@mui/x-data-grid";

export type SortDir = "asc" | "desc";
export type AllowedSortBy = "created_at" | "updated_at" | "title" | "team_size" | "duration";

type Props<Row extends GridValidRowModel> = {
  rows: Row[];
  rowCount: number;
  loading?: boolean;
  errorText?: string | null;

  columns: GridColDef<Row>[];

  page: number;        // 0-based
  pageSize: number;
  sortBy: AllowedSortBy;
  sortDir: SortDir;
  q: string;

  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onSortChange: (by: AllowedSortBy, dir: SortDir) => void;
  onSearchChange: (q: string) => void;

  onRowOpen?: (id: string) => void;
  getRowId: (row: Row) => GridRowId;
};

export function DataTable<Row extends GridValidRowModel>(props: Props<Row>) {
  const {
    rows, rowCount, loading, errorText,
    columns, page, pageSize, sortBy, sortDir, q,
    onPageChange, onPageSizeChange, onSortChange, onSearchChange,
    onRowOpen, getRowId,
  } = props;

  // Keep quick filter input in sync with URL/state and debounce updates
  const [quick, setQuick] = React.useState(q);
  React.useEffect(() => setQuick(q), [q]);
  React.useEffect(() => {
    const t = setTimeout(() => {
      if (quick !== q) onSearchChange(quick);
    }, 500);
    return () => clearTimeout(t);
  }, [quick, q, onSearchChange]);

  const paginationModel = React.useMemo<GridPaginationModel>(
    () => ({ page, pageSize }),
    [page, pageSize]
  );

  const sortModel = React.useMemo<GridSortModel>(
    () => [{
      field: sortBy === "duration" ? "duration_estimate" : sortBy,
      sort: sortDir,
    }],
    [sortBy, sortDir]
  );

  const handlePagination = (m: GridPaginationModel) => {
    // Only call if different to avoid loops
    if (m.pageSize !== pageSize) {
      onPageSizeChange(m.pageSize);
    }
    if (m.page !== page) {
      onPageChange(m.page);
    }
  };

  const handleSort = (m: GridSortModel) => {
    const s = m[0];
    if (!s?.field || !s.sort) return;
    const map: Record<string, AllowedSortBy | null> = {
      created_at: "created_at",
      updated_at: "updated_at",
      title: "title",
      team_size: "team_size",
      duration_estimate: "duration", // UI field → API sort key
    };
    const apiField = map[s.field] ?? "created_at";
    onSortChange(apiField, s.sort as SortDir);
  };

  const handleFilter = (m: GridFilterModel) => {
    const v = (m.quickFilterValues?.[0] ?? "").toString();
    setQuick(v);
  };

  return (
    <Paper sx={{ height: 560, width: "100%" }}>
      {errorText && <Alert severity="error" sx={{ m: 1 }}>{errorText}</Alert>}
      <DataGrid<Row>
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
        pageSizeOptions={[10, 20, 50, 100]}

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
