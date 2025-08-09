import { Outlet, Link, useLocation, useNavigate } from "react-router-dom";
import { AppBar, Toolbar, Typography, Button, Container, Box } from "@mui/material";
import { isAuthed, logout, revokeRefreshToken } from "./lib/auth";

export default function App() {
  const loc = useLocation();
  const nav = useNavigate();

  const handleLogout = async () => {
    await revokeRefreshToken().catch(() => {});
    logout();
    nav("/");
  };

  return (
    <Box minHeight="100dvh" display="flex" flexDirection="column">
      <AppBar position="static" color="default" elevation={0}>
        <Toolbar>
          <Typography variant="h6" sx={{ flexGrow: 1 }}>
            <Link to="/" style={{ textDecoration: "none", color: "inherit" }}>TrustFlow</Link>
          </Typography>
          {isAuthed() ? (
            <Button onClick={handleLogout}>Logout</Button>
          ) : (
            <>
              {loc.pathname !== "/dashboard" && (
                <Button component={Link} to="/dashboard">Dashboard</Button>
              )}
              <Button component={Link} to="/register">Register</Button>
            </>
          )}
        </Toolbar>
      </AppBar>
      <Container sx={{ py: 6, flexGrow: 1 }}>
        <Outlet />
      </Container>
    </Box>
  );
}
