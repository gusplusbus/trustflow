import { Stack, Typography, Button } from "@mui/material";
import { Link } from "react-router-dom";

export default function Dashboard() {
  return (
    <Stack spacing={2}>
      <Typography variant="h5">Dashboard</Typography>
      <Typography color="text.secondary">You’re logged in. Wire real data next.</Typography>
      <Button component={Link} to="/projects/create" variant="outlined" size="small">
        New Project Creator
      </Button>
      <Button component={Link} to="/projects" variant="outlined" size="small">
        Projects Explorer
      </Button>
      <Button component={Link} to="/" variant="outlined" size="small">
        Back to Login
      </Button>
    </Stack>
  );
}
