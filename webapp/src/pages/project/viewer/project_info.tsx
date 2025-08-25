import { Link as RouterLink, useNavigate } from "react-router-dom";
import {Chip, Alert, Button, Divider, Paper, Stack, Typography } from "@mui/material";
import type { ProjectResponse } from "../../../lib/projects";
import { useProjectWallet } from "../../../hooks/project";

type Props = {
  id: string;
  data: ProjectResponse;
  deleting: boolean;
  delErr: string | null;
  onDelete: () => Promise<void>;
};


export default function ProjectInfo({ id, data, deleting, delErr, onDelete }: Props) {
  const nav = useNavigate();
  const { wallet } = useProjectWallet(id)
  const handleDelete = async () => {
    if (!confirm("Delete this project? This cannot be undone.")) return;
    try {
      await onDelete();
      nav("/dashboard");
    } catch {}
  };

  return (
    <Paper sx={{ p: 3 }}>
      <Stack spacing={1}>
        <Typography variant="h5">{data.title}</Typography>
        <Typography color="text.secondary">{data.description}</Typography>
        <Divider sx={{ my: 2 }} />
        <Typography>Team size: {data.team_size}</Typography>
        <Typography>Duration: {data.duration_estimate} days</Typography>
        <Typography>Apply by: {new Date(data.application_close_time).toLocaleString()}</Typography>

        {delErr && <Alert severity="error">{delErr}</Alert>}

        <Stack direction="row" spacing={1} sx={{ mt: 2 }}>
          <Button component={RouterLink} to={`/projects/${id}/edit`} variant="outlined">
            Edit
          </Button>
          <Button onClick={handleDelete} color="error" disabled={deleting}>
            Delete
          </Button>
          <Button onClick={() => nav(-1)}>Back</Button>
          {wallet ? (
            <Chip label={`Wallet: ${wallet.address.slice(0,6)}…${wallet.address.slice(-4)}`} color="success" />
          ) : (
          <Button component={RouterLink} to={`/projects/${id}/wallet`} variant="contained">Attach wallet</Button>
          )}
          </Stack>
          </Stack>
          </Paper>
  );
}
