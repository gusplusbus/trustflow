import * as React from "react";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import {
  Alert,
  Button,
  Divider,
  Paper,
  Stack,
  Typography,
  Chip,
  Link,
  TextField,
  CircularProgress,
  MenuItem,
  Select,
  FormControl,
  InputLabel,
} from "@mui/material";
import GitHubIcon from "@mui/icons-material/GitHub";
import Modal from "../../components/Modal";
import { useDeleteProject, useProject } from "../../hooks/project";
import { getProject, type ProjectResponse } from "../../lib/projects";
import { createOwnership } from "../../lib/ownership";

const APP_SLUG = "trusflow"; // your GitHub App slug

// ---- Types for GitHub responses (minimal) ----
type GitHubIssue = {
  id: number;
  number: number;
  title: string;
  state: "open" | "closed";
  html_url: string;
  user?: { login?: string };
  created_at: string;
  updated_at: string;
  pull_request?: object; // PRs show up in /issues if present
  labels?: Array<{ name?: string }>;
};
type GitHubSearchResponse = { total_count: number; items: GitHubIssue[] };

// ---- Ownership type (snake_case to match API) ----
type Ownership = {
  id: string;
  created_at: string;
  updated_at: string;
  project_id: string;
  user_id: string;
  organization: string;
  repository: string;
  provider?: string;
  web_url?: string;
};

// ---- Filter form state ----
type ImportFilters = {
  owner: string;
  repo: string;
  state: "open" | "closed" | "all";
  labels: string;
  assignee: string; // "", "*", or login
  since: string; // datetime-local
  per_page: number; // 1..100
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

export default function ProjectViewer() {
  // -------- Router / data hooks --------
  const { id = "" } = useParams();
  const nav = useNavigate();
  const { data, loading, error } = useProject(id);
  const { remove, loading: deleting, error: delErr } = useDeleteProject(id);

  // -------- Ownership (repo connection) local state --------
  const [ownerships, setOwnerships] = React.useState<Ownership[]>([]);
  const [connected, setConnected] = React.useState(false);

  const [provider, setProvider] = React.useState<"github">("github");
  const [repoOwner, setRepoOwner] = React.useState("");
  const [repoName, setRepoName] = React.useState("");
  const [savingOwnership, setSavingOwnership] = React.useState(false);
  const [saveOwnershipErr, setSaveOwnershipErr] = React.useState<string | null>(null);

  // hydrate ownerships explicitly (works even if useProject doesn’t include them yet)
  React.useEffect(() => {
    let cancelled = false;
    const run = async () => {
      if (!id) return;
      try {
        const full: ProjectResponse & { ownerships?: Ownership[] } =
          await getProject(id, { includeOwnerships: true } as any);
        if (cancelled) return;
        const owns = full.ownerships ?? [];
        setOwnerships(owns);
        setConnected(owns.length > 0);
        if (owns.length > 0) {
          // prefill from the first ownership
          setRepoOwner((prev) => prev || owns[0].organization || "");
          setRepoName((prev) => prev || owns[0].repository || "");
          setProvider((prev) => prev || (owns[0].provider as any) || "github");
        }
      } catch {
        // ignore; base project still renders
      }
    };
    run();
    return () => {
      cancelled = true;
    };
  }, [id]);

  const onSaveOwnership = async () => {
    if (!repoOwner.trim() || !repoName.trim()) return;
    setSaveOwnershipErr(null);
    setSavingOwnership(true);
    try {
      const web_url = `https://github.com/${repoOwner.trim()}/${repoName.trim()}`;
      await createOwnership(id, {
        organization: repoOwner.trim(),
        repository: repoName.trim(),
        provider,
        web_url,
      });
      // refresh list
      const full: ProjectResponse & { ownerships?: Ownership[] } =
        await getProject(id, { includeOwnerships: true } as any);
      const owns = full.ownerships ?? [];
      setOwnerships(owns);
      setConnected(owns.length > 0);
    } catch (e: any) {
      setSaveOwnershipErr(e?.message || "Failed to save repository ownership");
    } finally {
      setSavingOwnership(false);
    }
  };

  // -------- Issues state (client-side experiment) --------
  const [issues, setIssues] = React.useState<GitHubIssue[]>([]);
  const [issuesLoading, setIssuesLoading] = React.useState(false);
  const [issuesErr, setIssuesErr] = React.useState<string | null>(null);
  const [rateRemaining, setRateRemaining] = React.useState<string | null>(null);

  // -------- Modal + filter state --------
  const [modalOpen, setModalOpen] = React.useState(false);
  const [filters, setFilters] = React.useState<ImportFilters>(defaultFilters);
  const [importing, setImporting] = React.useState(false);
  const [importErr, setImportErr] = React.useState<string | null>(null);

  // -------- Helpers --------
  const buildIssuesUrl = (f: ImportFilters) => {
    const base = `https://api.github.com/repos/${encodeURIComponent(f.owner)}/${encodeURIComponent(
      f.repo
    )}/issues`;
    const params = new URLSearchParams();
    params.set("state", f.state);
    if (f.labels.trim()) params.set("labels", f.labels);
    if (f.assignee !== "") params.set("assignee", f.assignee);
    if (f.since) {
      const sinceIso = new Date(f.since).toISOString();
      if (!isNaN(Date.parse(sinceIso))) params.set("since", sinceIso);
    }
    params.set("per_page", String(Math.max(1, Math.min(100, f.per_page || 50))));
    return `${base}?${params.toString()}`;
  };
  const buildSearchUrl = (f: ImportFilters) => {
    const base = `https://api.github.com/search/issues`;
    const terms: string[] = [`repo:${f.owner}/${f.repo}`, `is:issue`];
    if (f.state !== "all") terms.push(`state:${f.state}`);
    if (f.labels.trim()) {
      f.labels
        .split(",")
        .map((s) => s.trim())
        .filter(Boolean)
        .forEach((lbl) => terms.push(`label:"${lbl.replaceAll('"', '\\"')}"`));
    }
    if (f.assignee === "*") terms.push("assignee:*");
    else if (f.assignee) terms.push(`assignee:${f.assignee}`);
    if (f.search.trim()) terms.push(f.search.trim());
    if (f.since) {
      const d = new Date(f.since);
      if (!isNaN(d.valueOf())) terms.push(`updated:>=${d.toISOString().slice(0, 10)}`);
    }
    const params = new URLSearchParams();
    params.set("q", terms.join(" "));
    params.set("per_page", String(Math.max(1, Math.min(100, f.per_page || 50))));
    return `${base}?${params.toString()}`;
  };

  const openModal = () => {
    setImportErr(null);
    // prefill owner/repo from Repository section if present
    setFilters((v) => ({
      ...v,
      owner: v.owner || repoOwner,
      repo: v.repo || repoName,
    }));
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
      const url = filters.search.trim() ? buildSearchUrl(filters) : buildIssuesUrl(filters);
      const res = await fetch(url, { headers: { Accept: "application/vnd.github+json" } });
      setRateRemaining(res.headers.get("x-ratelimit-remaining"));
      if (!res.ok) throw new Error(`GitHub (${res.status}): ${await res.text()}`);
      if (filters.search.trim()) {
        const json: GitHubSearchResponse = await res.json();
        setIssues(json.items.filter((it) => !it.pull_request));
      } else {
        const json: GitHubIssue[] = await res.json();
        setIssues(json.filter((it) => !it.pull_request));
      }
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

  // -------- Existing viewer short-circuits --------
  if (loading) return <>Loading…</>;
  if (error) return <Alert severity="error">{error}</Alert>;
  if (!data) return <Alert severity="warning">Not found</Alert>;

  const onDelete = async () => {
    if (!confirm("Delete this project? This cannot be undone.")) return;
    try {
      await remove();
      nav("/dashboard");
    } catch {}
  };

  return (
    <Stack spacing={2}>
      {/* Top details card */}
      <Paper sx={{ p: 3 }}>
        <Stack spacing={1}>
          <Typography variant="h5">{data.title}</Typography>
          <Typography color="text.secondary">{data.description}</Typography>
          <Divider sx={{ my: 2 }} />
          <Typography>Team size: {data.team_size}</Typography>
          <Typography>Duration: {data.duration_estimate} days</Typography>
          <Typography>
            Apply by: {new Date(data.application_close_time).toLocaleString()}
          </Typography>

          {delErr && <Alert severity="error">{delErr}</Alert>}

          <Stack direction="row" spacing={1} sx={{ mt: 2 }}>
            <Button component={RouterLink} to={`/projects/${id}/edit`} variant="outlined">
              Edit
            </Button>
            <Button onClick={onDelete} color="error" disabled={deleting}>
              Delete
            </Button>
            <Button onClick={() => nav(-1)}>Back</Button>
          </Stack>
        </Stack>
      </Paper>

      {/* Repository section (integration + ownership) */}
      <Paper sx={{ p: 3 }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
          <Typography variant="h6">Repository</Typography>
          <Chip
            size="small"
            color={connected ? "success" : "default"}
            label={connected ? "Connected" : "Not connected"}
          />
        </Stack>

        {/* Inputs */}
        <Stack direction={{ xs: "column", sm: "row" }} spacing={2} sx={{ mb: 2 }}>
          <FormControl fullWidth>
            <InputLabel id="provider-label">Provider</InputLabel>
            <Select
              labelId="provider-label"
              value={provider}
              label="Provider"
              onChange={(e) => setProvider(e.target.value as "github")}
            >
              <MenuItem value="github">
                <Stack direction="row" alignItems="center" gap={1}>
                  <GitHubIcon fontSize="small" /> GitHub
                </Stack>
              </MenuItem>
            </Select>
          </FormControl>

          <TextField
            label="Owner"
            placeholder="e.g. your-org"
            helperText="GitHub org or user that owns the repo"
            value={repoOwner}
            onChange={(e) => setRepoOwner(e.target.value)}
            fullWidth
          />
          <TextField
            label="Repository"
            placeholder="e.g. your-repo"
            helperText="Repository name only"
            value={repoName}
            onChange={(e) => setRepoName(e.target.value)}
            fullWidth
          />
        </Stack>

        {/* Helpful guide + actions */}
        {(() => {
          const hasRepo = !!repoOwner && !!repoName;
          const appPageUrl = `https://github.com/apps/${APP_SLUG}`;
          const installUrl = `https://github.com/apps/${APP_SLUG}/installations/new`;
          const repoUrl = hasRepo ? `https://github.com/${repoOwner}/${repoName}` : "";
          const repoInstallsUrl = hasRepo ? `${repoUrl}/settings/installations` : "";
          const openIssuesUrl = hasRepo ? `${repoUrl}/issues?q=is%3Aissue+state%3Aopen` : "";

          return (
            <>
              <Stack spacing={1} sx={{ mb: 2 }}>
                <Typography variant="body2" color="text.secondary">
                  Connect your GitHub repo so TrustFlow can read issues for this project.
                </Typography>

                <Stack direction={{ xs: "column", sm: "row" }} spacing={2} alignItems={{ sm: "center" }}>
                  <Chip size="small" label="Step 1" color="primary" />
                  <Typography>
                    <strong>Install the app</strong>:{" "}
                    <Link underline="hover" href={appPageUrl} target="_blank" rel="noreferrer">
                      open app page
                    </Link>{" "}
                    → click <b>Install</b> → pick your org → select the repo(s). Or jump straight to{" "}
                    <Link underline="hover" href={installUrl} target="_blank" rel="noreferrer">
                      install now
                    </Link>.
                  </Typography>
                </Stack>

                <Stack direction={{ xs: "column", sm: "row" }} spacing={2} alignItems={{ sm: "center" }}>
                  <Chip size="small" label="Step 2" color="primary" />
                  <Typography>
                    <strong>Confirm repo</strong>: {hasRepo ? (
                      <Link underline="hover" href={repoUrl} target="_blank" rel="noreferrer">
                        {repoOwner}/{repoName}
                      </Link>
                    ) : (
                      <em>enter owner/repo above</em>
                    )}. You can manage installed apps per repo at{" "}
                    {hasRepo ? (
                      <Link underline="hover" href={repoInstallsUrl} target="_blank" rel="noreferrer">
                        Settings → Installed GitHub Apps
                      </Link>
                    ) : (
                      <em>repo settings (link appears after you enter owner/repo)</em>
                    )}.
                  </Typography>
                </Stack>

                <Stack direction={{ xs: "column", sm: "row" }} spacing={2} alignItems={{ sm: "center" }}>
                  <Chip size="small" label="Step 3" color="primary" />
                  <Typography>
                    <strong>Preview issues</strong> (optional):{" "}
                    {hasRepo ? (
                      <Link underline="hover" href={openIssuesUrl} target="_blank" rel="noreferrer">
                        view open issues on GitHub
                      </Link>
                    ) : (
                      <em>enter owner/repo to see the link</em>
                    )}.
                  </Typography>
                </Stack>
              </Stack>

              {saveOwnershipErr && <Alert severity="error" sx={{ mb: 1 }}>{saveOwnershipErr}</Alert>}

              <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
                <Button
                  variant="contained"
                  startIcon={<GitHubIcon />}
                  href={installUrl}
                  target="_blank"
                  rel="noreferrer"
                >
                  Connect GitHub
                </Button>

                <Button
                  variant="outlined"
                  disabled={!hasRepo}
                  component={Link as any}
                  href={hasRepo ? repoUrl : undefined}
                  target="_blank"
                  rel="noreferrer"
                >
                  Open repo
                </Button>

                <Button
                  variant="outlined"
                  disabled={!hasRepo}
                  component={Link as any}
                  href={hasRepo ? repoInstallsUrl : undefined}
                  target="_blank"
                  rel="noreferrer"
                >
                  Repo installations
                </Button>

                <Button
                  variant="outlined"
                  disabled={!hasRepo || savingOwnership}
                  onClick={onSaveOwnership}
                  sx={{ ml: { sm: "auto" } }}
                >
                  {savingOwnership ? "Saving…" : "Save ownership"}
                </Button>

                <Button
                  variant="outlined"
                  disabled={!hasRepo}
                  onClick={() => {
                    setFilters((v) => ({ ...v, owner: repoOwner, repo: repoName }));
                    setModalOpen(true);
                  }}
                >
                  Use in Issues
                </Button>
              </Stack>

              <Typography variant="body2" color="text.secondary" sx={{ mt: 1 }}>
                Private repos will require a backend proxy using the app installation; public repos work from the browser for testing.
              </Typography>
            </>
          );
        })()}

        <Divider sx={{ my: 2 }} />

        <Typography variant="subtitle1">Current repositories</Typography>
        {ownerships.length === 0 ? (
          <Typography variant="body2" color="text.secondary">None yet.</Typography>
        ) : (
          <Stack spacing={0.5} sx={{ mt: 1 }}>
            {ownerships.map((o) => (
              <Stack key={o.id} direction="row" alignItems="center" gap={1} flexWrap="wrap">
                <Chip size="small" label={`${o.organization}/${o.repository}`} />
                {o.provider && <Chip size="small" variant="outlined" label={o.provider} />}
                {o.web_url && (
                  <Link href={o.web_url} target="_blank" rel="noreferrer" underline="hover">
                    {o.web_url}
                  </Link>
                )}
              </Stack>
            ))}
          </Stack>
        )}
      </Paper>

      {/* Issues section */}
      <Paper sx={{ p: 3 }}>
        <Stack direction="row" alignItems="center" justifyContent="space-between">
          <Typography variant="h6">Issues</Typography>
          <Stack direction="row" alignItems="center" gap={1}>
            {rateRemaining != null && <Chip size="small" label={`GitHub rate left: ${rateRemaining}`} />}
            <Button variant="outlined" startIcon={<GitHubIcon />} onClick={openModal}>
              Attach issues
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
          <Stack spacing={1} sx={{ mt: 2 }}>
            {issues.length === 0 && (
              <Typography color="text.secondary">
                No issues loaded yet. Click “Attach issues” to fetch from GitHub.
              </Typography>
            )}
            {issues.map((it) => (
              <Stack key={it.id} direction="row" alignItems="center" gap={1} flexWrap="wrap">
                <Chip size="small" label={`#${it.number}`} />
                <Link href={it.html_url} target="_blank" rel="noreferrer" underline="hover">
                  {it.title}
                </Link>
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
      </Paper>

      {/* Modal for GitHub filters and fetch */}
      <Modal
        open={modalOpen}
        onClose={closeModal}
        title="Attach GitHub Issues (public)"
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
            For real customers, this will call your backend with an installation token from the GitHub App.
            For now, this uses the public GitHub API in the browser for experimentation.
          </Typography>
        </Stack>
      </Modal>
    </Stack>
  );
}
