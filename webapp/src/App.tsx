import * as React from "react";
import { Outlet, Link, useLocation, useNavigate } from "react-router-dom";
import {
  AppBar, Toolbar, Typography, Button, Container, Box,
  IconButton, Drawer, useMediaQuery
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
  const authed = isAuthed();

  // one state to rule them all
  const [navOpen, setNavOpen] = React.useState(false);
  const toggleNav = () => setNavOpen((v) => !v);

  // when breakpoint changes: open by default on desktop, closed on mobile
  React.useEffect(() => {
    setNavOpen(isDesktop); // true on md+, false on xs-sm
  }, [isDesktop]);

  const handleLogout = async () => {
    await revokeRefreshToken().catch(() => {});
    logout();
    setNavOpen(isDesktop); // after logout, keep desktop layout tidy
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
          ...(authed && isDesktop && navOpen && {
            width: `calc(100% - ${drawerWidth}px)`,
            ml: `${drawerWidth}px`,
          }),
        }}
      >
        <Toolbar>
          {authed && (
            <IconButton edge="start" onClick={toggleNav} aria-label="open navigation" sx={{ mr: 1 }}>
              <MenuIcon />
            </IconButton>
          )}

          <Typography variant="h6" sx={{ flexGrow: 1 }}>
            {/* title optional */}
          </Typography>

          {authed ? (
            <Button onClick={handleLogout}>Logout</Button>
          ) : (
            <>
              {loc.pathname !== "/login" && <Button component={Link} to="/login">Login</Button>}
              <Button component={Link} to="/register">Register</Button>
            </>
          )}
        </Toolbar>
      </AppBar>

      {authed && (
        <>
          {/* Mobile: temporary modal drawer, controlled by navOpen */}
          <Drawer
            variant="temporary"
            open={navOpen}
            onClose={toggleNav}
            ModalProps={{ keepMounted: true }}
            sx={{
              display: { xs: "block", md: "none" },
              "& .MuiDrawer-paper": { width: drawerWidth },
            }}
          >
            <Sidebar onNavigate={() => !isDesktop && toggleNav()} />
          </Drawer>

          {/* Desktop: persistent drawer (controlled), not permanent */}
          <Drawer
            variant="persistent"
            open={navOpen}
            sx={{
              display: { xs: "none", md: "block" },
              "& .MuiDrawer-paper": { width: drawerWidth, boxSizing: "border-box" },
            }}
          >
            <Sidebar />
          </Drawer>
        </>
      )}

      <Box
        component="main"
        sx={{
          flexGrow: 1,
          ml: { xs: 0, md: authed && navOpen ? `${drawerWidth}px` : 0 },
          width: "100%",
        }}
      >
        <Toolbar />
        <Container maxWidth="xl" sx={{ py: 3 }}>
          <Outlet />
        </Container>
      </Box>
    </Box>
  );
}
