import * as React from "react";
import {
  Alert, Button, Chip, CircularProgress, Checkbox, FormControlLabel,
  MenuItem, Paper, Stack, TextField, Typography
} from "@mui/material";
import GitHubIcon from "@mui/icons-material/GitHub";
import Modal from "../../../components/Modal";
import { useOwnershipIssues, type ImportFilters } from "../../../hooks/project";
import { postOwnershipIssues } from "../../../lib/ownership";

type Props = { projectId: string };

export default function ListOriginIssues({ projectId }: Props) {
  const {
    filters, setFilters,
    issues, loading: issuesLoading, error: issuesErr, rateRemaining,
    listIssues,
  } = useOwnershipIssues({ projectId });

  // auto-load origin issues once on mount
  const didLoadRef = React.useRef(false);
  React.useEffect(() => {
    if (didLoadRef.current) return;
    didLoadRef.current = true;
    void listIssues();
  }, [listIssues]);

  const [modalOpen, setModalOpen] = React.useState(false);
  const openModal = () => setModalOpen(true);
  const closeModal = () => setModalOpen(false);

  // selection
  const [selectedIds, setSelectedIds] = React.useState<Set<number>>(new Set());
  const [selectAll, setSelectAll] = React.useState(false);

  React.useEffect(() => {
    setSelectedIds(new Set());
    setSelectAll(false);
  }, [issues]);

  const toggleOne = (id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  const toggleAll = () => {
    if (selectAll) {
      setSelectedIds(new Set());
      setSelectAll(false);
    } else {
      setSelectedIds(new Set(issues.map((i) => i.id)));
      setSelectAll(true);
    }
  };

  // list from modal
  const [listing, setListing] = React.useState(false);
  const handleList = async () => {
    setListing(true);
    try {
      await listIssues();
      setModalOpen(false);
    } finally {
      setListing(false);
    }
  };

  // post just IDs
  const [importing, setImporting] = React.useState(false);
  const [importErr, setImportErr] = React.useState<string | null>(null);
  const [importOk, setImportOk] = React.useState<string | null>(null);

  const handleImportSelected = async () => {
    setImportErr(null);
    setImportOk(null);

    const ids = Array.from(selectedIds);
    if (ids.length === 0) {
      setImportErr("Select at least one issue.");
      return;
    }

    // Convert ids -> [{ id, number }]
    const selected = ids
      .map((id) => {
        const it = issues.find((x) => x.id === id);
        return it ? { id: it.id, number: it.number } : null;
      })
      .filter((x): x is { id: number; number: number } => !!x);

    if (selected.length === 0) {
      setImportErr("Could not resolve selected issue numbers.");
      return;
    }

    try {
      setImporting(true);
      await postOwnershipIssues(projectId, selected); // sends { issues: [...] }
      setImportOk(`Imported ${selected.length} issue${selected.length > 1 ? "s" : ""}.`);
      setSelectedIds(new Set());
      setSelectAll(false);
    } catch (e: any) {
      setImportErr(e?.message || "Import failed");
    } finally {
      setImporting(false);
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack direction="row" alignItems="center" justifyContent="space-between" gap={2}>
        <Typography variant="h6">Issues</Typography>

        <Stack direction="row" alignItems="center" gap={1} flexWrap="wrap">
          {rateRemaining != null && <Chip size="small" label={`GitHub rate left: ${rateRemaining}`} />}
          <Button variant="outlined" startIcon={<GitHubIcon />} onClick={openModal}>
            Load issues
          </Button>
          <Button
            variant="contained"
            onClick={handleImportSelected}
            disabled={importing || issues.length === 0 || selectedIds.size === 0}
          >
            {importing ? "Importing…" : `Import selected (${selectedIds.size || 0})`}
          </Button>
        </Stack>
      </Stack>

      {importErr && <Alert severity="error" sx={{ mt: 2 }}>{importErr}</Alert>}
      {importOk && <Alert severity="success" sx={{ mt: 2 }}>{importOk}</Alert>}

      {issuesLoading ? (
        <Stack alignItems="center" sx={{ mt: 2 }}>
          <CircularProgress />
        </Stack>
      ) : issuesErr ? (
        <Alert severity="error" sx={{ mt: 2 }}>{issuesErr}</Alert>
      ) : (
        <Stack spacing={1} sx={{ mt: 2, maxHeight: "60vh", overflowY: "auto", pr: 1 }}>
          {issues.length === 0 && (
            <Typography color="text.secondary">
              No issues match the current filters.
            </Typography>
          )}

          {issues.length > 0 && (
            <FormControlLabel
              control={
                <Checkbox
                  checked={selectAll}
                  indeterminate={!selectAll && selectedIds.size > 0}
                  onChange={toggleAll}
                />
              }
              label={`Select all (${selectedIds.size}/${issues.length})`}
              sx={{ mb: 1 }}
            />
          )}

          {issues.map((it) => (
            <Stack key={it.id} direction="row" alignItems="center" gap={1} flexWrap="wrap">
              <Checkbox checked={selectedIds.has(it.id)} onChange={() => toggleOne(it.id)} />
              <Chip size="small" label={`#${it.number}`} />
              <a href={it.html_url} target="_blank" rel="noreferrer" style={{ textDecoration: "none" }}>
                {it.title}
              </a>
              <Chip size="small" label={it.state} variant="outlined" />
              {it.user?.login && <Chip size="small" label={`@${it.user.login}`} />}
              {it.labels && it.labels.length > 0 && (
                <Chip
                  size="small"
                  variant="outlined"
                  label={it.labels.slice(0, 2).map((l) => l?.name).filter(Boolean).join(", ")}
                />
              )}
            </Stack>
          ))}
        </Stack>
      )}

      <Modal
        open={modalOpen}
        onClose={closeModal}
        title="List Issues"
        actions={
          <>
            <Button onClick={closeModal} disabled={listing}>Cancel</Button>
            <Button
              onClick={handleList}
              variant="contained"
              disabled={
                listing ||
                filters.per_page < 1 ||
                filters.per_page > 100
              }
            >
              {listing ? "Listing…" : "List"}
            </Button>
          </>
        }
      >
        <Stack spacing={2}>
          <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
            <TextField
              label="State"
              select
              value={filters.state}
              onChange={(e) => setFilters((v) => ({ ...v, state: e.target.value as ImportFilters["state"] }))}
              fullWidth
              SelectProps={{ MenuProps: { disableScrollLock: true } }}
            >
              <MenuItem value="open">Open</MenuItem>
              <MenuItem value="closed">Closed</MenuItem>
              <MenuItem value="all">All</MenuItem>
            </TextField>

            <TextField
              label="Labels (comma-separated)"
              placeholder="bug,help wanted"
              value={filters.labels}
              onChange={(e) => setFilters((v) => ({ ...v, labels: e.target.value }))}
              fullWidth
            />
          </Stack>

          <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
            <TextField
              label="Assignee"
              placeholder="login, * for any assigned, blank for any"
              value={filters.assignee}
              onChange={(e) => setFilters((v) => ({ ...v, assignee: e.target.value }))}
              fullWidth
            />
            <TextField
              label="Since (updated after)"
              type="datetime-local"
              InputLabelProps={{ shrink: true }}
              value={filters.since}
              onChange={(e) => setFilters((v) => ({ ...v, since: e.target.value }))}
              fullWidth
            />
          </Stack>

          <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
            <TextField
              label="Per page"
              type="number"
              inputProps={{ min: 1, max: 100 }}
              value={filters.per_page}
              onChange={(e) => setFilters((v) => ({ ...v, per_page: Number(e.target.value || 0) }))}
              fullWidth
            />
            <TextField
              label="Search text (optional)"
              placeholder="Search title/body (uses Search API)"
              value={filters.search}
              onChange={(e) => setFilters((v) => ({ ...v, search: e.target.value }))}
              fullWidth
            />
          </Stack>

          <Typography variant="body2" color="text.secondary">
            Loads via your backend using the GitHub App installation token.
          </Typography>
        </Stack>
      </Modal>
    </Paper>
  );
}
