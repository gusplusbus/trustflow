import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import { Alert, Button, Divider, Paper, Stack, Typography } from "@mui/material";
import { useDeleteProject, useProject } from "../../hooks/project";

export default function ProjectViewer() {
  const { id = "" } = useParams();
  const nav = useNavigate();
  const { data, loading, error } = useProject(id);
  const { remove, loading: deleting, error: delErr } = useDeleteProject(id);

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
          <Button component={RouterLink} to={`/projects/${id}/edit`} variant="outlined">Edit</Button>
          <Button onClick={onDelete} color="error" disabled={deleting}>Delete</Button>
          <Button onClick={() => nav(-1)}>Back</Button>
        </Stack>
      </Stack>
    </Paper>
  );
}
