import { FormEvent, useState } from "react";
import { useNavigate, Link as RouterLink } from "react-router-dom";
import { Paper, Box, Stack, Typography, TextField, Button, Alert, Link } from "@mui/material";
import { register, login } from "../lib/auth";

export default function Register() {
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
      await register(email, password);
      await login(email, password); // auto-login after signup
      nav("/dashboard");
    } catch (e: any) {
      setErr(e?.message || "Registration failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <Paper elevation={1} sx={{ p: 3 }}>
      <Box component="form" onSubmit={onSubmit}>
        <Stack spacing={2}>
          <div>
            <Typography variant="h6">Create account</Typography>
            <Typography variant="body2" color="text.secondary">Register to continue</Typography>
          </div>
          {err && <Alert severity="error">{err}</Alert>}
          <TextField label="Email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} required fullWidth />
          <TextField label="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} required fullWidth />
          <Button type="submit" variant="contained" disabled={loading}>{loading ? "Creating..." : "Create account"}</Button>
          <Typography variant="body2">
            Already have an account?{" "}
            <Link component={RouterLink} to="/">Sign in</Link>
          </Typography>
        </Stack>
      </Box>
    </Paper>
  );
}
