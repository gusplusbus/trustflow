import * as React from "react";
import { Outlet, Link, useLocation, useNavigate } from "react-router-dom";
import {
  AppBar,
  Toolbar,
  Typography,
  Button,
  Container,
  Box,
  IconButton,
  Drawer,
  useMediaQuery,
} from "@mui/material";
import { useTheme } from "@mui/material/styles";
import MenuIcon from "@mui/icons-material/Menu";
import { isAuthed, logout, revokeRefreshToken } from "./lib/auth";
import Sidebar, { drawerWidth } from "./components/Sidebar";

export default function App() {
  const loc = useLocation();
  const nav = useNavigate();
  const theme = useTheme();
  const isDesktop = useMediaQuery(theme.breakpoints.up("md"));

  const [mobileOpen, setMobileOpen] = React.useState(false);
  const toggleDrawer = () => setMobileOpen((v) => !v);

  const authed = isAuthed();

  const handleLogout = async () => {
    await revokeRefreshToken().catch(() => {});
    logout();
    setMobileOpen(false);
    nav("/");
  };

  return (
    <Box minHeight="100dvh" display="flex">
      <AppBar
        position="fixed"
        color="default"
        elevation={0}
        sx={{
          zIndex: (t) => t.zIndex.drawer + 1,
          ...(authed && isDesktop && {
            width: `calc(100% - ${drawerWidth}px)`,
            ml: `${drawerWidth}px`,
          }),
        }}
      >
        <Toolbar>
          {authed && (
            <IconButton
              edge="start"
              onClick={toggleDrawer}
              aria-label="open navigation"
              sx={{ mr: 1 }}
            >
              <MenuIcon />
            </IconButton>
          )}

          <Typography variant="h6" sx={{ flexGrow: 1 }}>
            <Link to="/" style={{ textDecoration: "none", color: "inherit" }}>
              TrustFlow
            </Link>
          </Typography>

          {authed ? (
            <Button onClick={handleLogout}>Logout</Button>
          ) : (
            <>
              {/* unauth header: keep it simple — no dashboard link */}
              {loc.pathname !== "/login" && (
                <Button component={Link} to="/login">Login</Button>
              )}
              <Button component={Link} to="/register">Register</Button>
            </>
          )}
        </Toolbar>
      </AppBar>

      {/* Drawers only if authed */}
      {authed && (
        <>
          <Drawer
            variant="temporary"
            open={mobileOpen}
            onClose={toggleDrawer}
            ModalProps={{ keepMounted: true }}
            sx={{
              display: { xs: "block", md: "none" },
              "& .MuiDrawer-paper": { width: drawerWidth },
            }}
          >
            <Sidebar onNavigate={toggleDrawer} />
          </Drawer>

          <Drawer
            variant="permanent"
            open
            sx={{
              display: { xs: "none", md: "block" },
              "& .MuiDrawer-paper": { width: drawerWidth, boxSizing: "border-box" },
            }}
          >
            <Sidebar />
          </Drawer>
        </>
      )}

      {/* Main */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          ml: { xs: 0, md: authed ? `${drawerWidth}px` : 0 },
          width: "100%",
        }}
      >
        {/* spacer below fixed AppBar */}
        <Toolbar />
        <Container sx={{ py: 3 }}>
          <Outlet />
        </Container>
      </Box>
    </Box>
  );
}
