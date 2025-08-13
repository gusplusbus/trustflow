import * as React from "react";
import {
  Alert, Button, Chip, Divider, FormControl, InputLabel, Link,
  MenuItem, Paper, Select, Stack, TextField, Typography
} from "@mui/material";
import GitHubIcon from "@mui/icons-material/GitHub";
import { createOwnership } from "../../../lib/ownership";

export type Ownership = {
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

type Props = {
  projectId: string;
  ownerships: Ownership[];
  onOwnershipSaved: () => Promise<void> | void;
};

const APP_SLUG = "trusflow";

export default function ConnectRepo({ projectId, ownerships, onOwnershipSaved }: Props) {
  const [provider, setProvider] = React.useState<"github">("github");
  const [repoOwner, setRepoOwner] = React.useState(ownerships[0]?.organization ?? "");
  const [repoName,  setRepoName]  = React.useState(ownerships[0]?.repository ?? "");
  const [saving, setSaving] = React.useState(false);
  const [err, setErr] = React.useState<string | null>(null);

  React.useEffect(() => {
    if (ownerships[0]) {
      setRepoOwner((v) => v || ownerships[0].organization || "");
      setRepoName((v)  => v || ownerships[0].repository || "");
    }
  }, [ownerships]);

  const onSave = async () => {
    if (!repoOwner.trim() || !repoName.trim()) return;
    setErr(null);
    setSaving(true);
    try {
      const web_url = `https://github.com/${repoOwner.trim()}/${repoName.trim()}`;
      await createOwnership(projectId, {
        organization: repoOwner.trim(),
        repository: repoName.trim(),
        provider,
        web_url,
      });
      await onOwnershipSaved();
    } catch (e: any) {
      setErr(e?.message || "Failed to save repository ownership");
    } finally {
      setSaving(false);
    }
  };

  const hasRepo = !!repoOwner && !!repoName;
  const appPageUrl = `https://github.com/apps/${APP_SLUG}`;
  const installUrl = `https://github.com/apps/${APP_SLUG}/installations/new`;
  const repoUrl = hasRepo ? `https://github.com/${repoOwner}/${repoName}` : "";
  const repoInstallsUrl = hasRepo ? `${repoUrl}/settings/installations` : "";
  const openIssuesUrl = hasRepo ? `${repoUrl}/issues?q=is%3Aissue+state%3Aopen` : "";

  return (
    <Paper sx={{ p: 3 }}>
      <Stack direction="row" alignItems="center" justifyContent="space-between" sx={{ mb: 2 }}>
        <Typography variant="h6">Repository</Typography>
        <Chip size="small" color={ownerships.length ? "success" : "default"}
              label={ownerships.length ? "Connected" : "Not connected"} />
      </Stack>

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

        <TextField label="Owner" placeholder="e.g. your-org" fullWidth
          value={repoOwner} onChange={(e) => setRepoOwner(e.target.value)} />
        <TextField label="Repository" placeholder="e.g. your-repo" fullWidth
          value={repoName} onChange={(e) => setRepoName(e.target.value)} />
      </Stack>

      <Stack spacing={1} sx={{ mb: 2 }}>
        <Typography variant="body2" color="text.secondary">
          Connect your GitHub repo so TrustFlow can read issues for this project.
        </Typography>

        <Stack direction={{ xs: "column", sm: "row" }} spacing={2} alignItems={{ sm: "center" }}>
          <Chip size="small" label="Step 1" color="primary" />
          <Typography>
            <strong>Install the app</strong>:{" "}
            <Link underline="hover" href={appPageUrl} target="_blank" rel="noreferrer">open app page</Link>{" "}
            → click <b>Install</b> → pick your org → select the repo(s). Or{" "}
            <Link underline="hover" href={installUrl} target="_blank" rel="noreferrer">install now</Link>.
          </Typography>
        </Stack>

        <Stack direction={{ xs: "column", sm: "row" }} spacing={2} alignItems={{ sm: "center" }}>
          <Chip size="small" label="Step 2" color="primary" />
          <Typography>
            <strong>Confirm repo</strong>: {hasRepo ? (
              <Link underline="hover" href={repoUrl} target="_blank" rel="noreferrer">
                {repoOwner}/{repoName}
              </Link>
            ) : <em>enter owner/repo above</em>}. Manage installed apps at{" "}
            {hasRepo ? (
              <Link underline="hover" href={repoInstallsUrl} target="_blank" rel="noreferrer">
                Settings → Installed GitHub Apps
              </Link>
            ) : <em>repo settings (appears after you enter owner/repo)</em>}.
          </Typography>
        </Stack>

        <Stack direction={{ xs: "column", sm: "row" }} spacing={2} alignItems={{ sm: "center" }}>
          <Chip size="small" label="Step 3" color="primary" />
          <Typography>
            <strong>Preview issues</strong>: {hasRepo ? (
              <Link underline="hover" href={openIssuesUrl} target="_blank" rel="noreferrer">
                view open issues on GitHub
              </Link>
            ) : <em>enter owner/repo to see the link</em>}.
          </Typography>
        </Stack>
      </Stack>

      {err && <Alert severity="error" sx={{ mb: 1 }}>{err}</Alert>}

      <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
        <Button variant="contained" startIcon={<GitHubIcon />} href={installUrl} target="_blank" rel="noreferrer">
          Connect GitHub
        </Button>
        <Button variant="outlined" disabled={!hasRepo} component={Link as any} href={hasRepo ? repoUrl : undefined} target="_blank" rel="noreferrer">
          Open repo
        </Button>
        <Button variant="outlined" disabled={!hasRepo} component={Link as any} href={hasRepo ? repoInstallsUrl : undefined} target="_blank" rel="noreferrer">
          Repo installations
        </Button>
        <Button variant="outlined" disabled={!hasRepo || saving} onClick={onSave} sx={{ ml: { sm: "auto" } }}>
          {saving ? "Saving…" : "Save ownership"}
        </Button>
      </Stack>

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
  );
}
