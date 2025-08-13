import * as React from "react";
import {
  Alert, Button, Chip, CircularProgress, MenuItem, Paper, Stack, TextField, Typography
} from "@mui/material";
import GitHubIcon from "@mui/icons-material/GitHub";
import { listOwnershipIssues } from "../../../lib/ownership";
import Modal from "../../../components/Modal";

type GitHubIssue = {
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

type Props = {
  projectId: string;
  defaultOwner?: string;
  defaultRepo?: string;
};

type ImportFilters = {
  owner: string;
  repo: string;
  state: "open" | "closed" | "all";
  labels: string;
  assignee: string;
  since: string;
  per_page: number;
  search: string;
};

const defaultFilters: ImportFilters = {
  owner: "",
  repo: "",
  state: "open",
  labels: "",
  assignee: "",
  since: "",
  per_page: 50,
  search: "",
};

export default function ListOriginIssues({ projectId, defaultOwner = "", defaultRepo = "" }: Props) {
  const [issues, setIssues] = React.useState<GitHubIssue[]>([]);
  const [issuesLoading, setIssuesLoading] = React.useState(false);
  const [issuesErr, setIssuesErr] = React.useState<string | null>(null);
  const [rateRemaining, setRateRemaining] = React.useState<string | null>(null);

  const [modalOpen, setModalOpen] = React.useState(false);
  const [filters, setFilters] = React.useState<ImportFilters>({
    ...defaultFilters,
    owner: defaultOwner,
    repo: defaultRepo,
  });
  const [importing, setImporting] = React.useState(false);
  const [importErr, setImportErr] = React.useState<string | null>(null);

  React.useEffect(() => {
    setFilters((v) => ({
      ...v,
      owner: v.owner || defaultOwner,
      repo: v.repo || defaultRepo,
    }));
  }, [defaultOwner, defaultRepo]);

  const openModal = () => {
    setImportErr(null);
    setModalOpen(true);
  };
  const closeModal = () => {
    if (!importing) setModalOpen(false);
  };

  const handleImport = async () => {
    setImportErr(null);
    setImporting(true);
    setIssuesErr(null);
    setIssuesLoading(true);
    try {
      if (!filters.owner.trim() || !filters.repo.trim()) {
        throw new Error("Owner and repo are required");
      }
      const sinceIso = filters.since ? new Date(filters.since).toISOString() : undefined;

      const json: {
        items: Array<{
          id: number; number: number; title: string; state: "open" | "closed";
          html_url: string; user_login?: string; labels?: string[];
          created_at: string; updated_at: string;
        }>;
        rate?: { remaining?: number };
      } = await listOwnershipIssues(projectId, {
        owner: filters.owner.trim(),
        repo: filters.repo.trim(),
        state: filters.state,
        labels: filters.labels || undefined,
        assignee: filters.assignee,
        since: sinceIso,
        per_page: Math.max(1, Math.min(100, filters.per_page || 50)),
        page: 1,
        search: filters.search.trim() || undefined,
      });

      setRateRemaining(json?.rate?.remaining != null ? String(json.rate.remaining) : null);

      const mapped: GitHubIssue[] = (json.items || []).map((it) => ({
        id: it.id,
        number: it.number,
        title: it.title,
        state: it.state,
        html_url: it.html_url,
        user: it.user_login ? { login: it.user_login } : undefined,
        created_at: it.created_at,
        updated_at: it.updated_at,
        labels: (it.labels || []).map((name) => ({ name })),
      })) as any;

      setIssues(mapped);
      setModalOpen(false);
    } catch (e: any) {
      const msg = e?.message || "Import failed";
      setImportErr(msg);
      setIssuesErr(msg);
    } finally {
      setImporting(false);
      setIssuesLoading(false);
    }
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack direction="row" alignItems="center" justifyContent="space-between">
        <Typography variant="h6">Issues</Typography>
        <Stack direction="row" alignItems="center" gap={1}>
          {rateRemaining != null && <Chip size="small" label={`GitHub rate left: ${rateRemaining}`} />}
          <Button variant="outlined" startIcon={<GitHubIcon />} onClick={openModal}>
            Load issues
          </Button>
        </Stack>
      </Stack>

      {issuesLoading ? (
        <Stack alignItems="center" sx={{ mt: 2 }}>
          <CircularProgress />
        </Stack>
      ) : issuesErr ? (
        <Alert severity="error" sx={{ mt: 2 }}>{issuesErr}</Alert>
      ) : (
        <Stack
          spacing={1}
          sx={{ mt: 2, maxHeight: "60vh", overflowY: "auto", pr: 1 }} // <-- scrollable list
        >
          {issues.length === 0 && (
            <Typography color="text.secondary">
              No issues loaded yet. Click “Load issues”.
            </Typography>
          )}
          {issues.map((it) => (
            <Stack key={it.id} direction="row" alignItems="center" gap={1} flexWrap="wrap">
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
            <Button onClick={closeModal} disabled={importing}>Cancel</Button>
            <Button
              onClick={handleImport}
              variant="contained"
              disabled={
                importing ||
                !filters.owner.trim() ||
                !filters.repo.trim() ||
                filters.per_page < 1 ||
                filters.per_page > 100
              }
            >
              {importing ? "Importing…" : "Import"}
            </Button>
          </>
        }
      >
        <Stack spacing={2}>
          {importErr && <Alert severity="error">{importErr}</Alert>}

          <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
            <TextField
              label="Owner"
              placeholder="e.g. gusplusbus"
              value={filters.owner}
              onChange={(e) => setFilters((v) => ({ ...v, owner: e.target.value }))}
              fullWidth
              required
            />
            <TextField
              label="Repository"
              placeholder="e.g. trustflow"
              value={filters.repo}
              onChange={(e) => setFilters((v) => ({ ...v, repo: e.target.value }))}
              fullWidth
              required
            />
          </Stack>

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

          <Stack direction={{ xs: "column, sm: row" } as any} spacing={2}>
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
