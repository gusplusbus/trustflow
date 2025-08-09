import { FormEvent, useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { Paper, Box, Stack, Typography, TextField, Button, Alert, Link } from "@mui/material";
import { login } from "../lib/auth";

export default function Login() {
  const nav = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: FormEvent) {
    e.preventDefault();
    setErr(null);
    setLoading(true);
    try {
      await login(email, password);
      nav("/dashboard");
    } catch (e: any) {
      setErr(e?.message || "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Paper elevation={1} sx={{ p: 3 }}>
      <Box component="form" onSubmit={onSubmit}>
        <Stack spacing={2}>
          <div>
            <Typography variant="h6">Welcome back</Typography>
            <Typography variant="body2" color="text.secondary">Sign in to continue</Typography>
          </div>
          {err && <Alert severity="error">{err}</Alert>}
          <TextField label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required fullWidth />
          <TextField label="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} required fullWidth />
          <Button type="submit" variant="contained" disabled={loading}>{loading ? "Signing in..." : "Sign in"}</Button>
          <Typography variant="body2">
            No account?{" "}
            <Link component={RouterLink} to="/register">Create one</Link>
          </Typography>
        </Stack>
      </Box>
    </Paper>
  );
}
