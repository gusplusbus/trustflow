import {
  Box,
  Divider,
  List,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Typography,
} from "@mui/material";
import { Link, useLocation } from "react-router-dom";
import DashboardIcon from "@mui/icons-material/Dashboard";
import AddCircleOutlineIcon from "@mui/icons-material/AddCircleOutline";
import TravelExploreIcon from "@mui/icons-material/TravelExplore";

export const drawerWidth = 240;

type Props = { onNavigate?: () => void };

export default function Sidebar({ onNavigate }: Props) {
  const loc = useLocation();

  const items = [
    { to: "/dashboard", label: "Dashboard", icon: <DashboardIcon /> },
    { to: "/projects/create", label: "Create Project", icon: <AddCircleOutlineIcon /> },
    { to: "/projects", label: "Explore Projects", icon: <TravelExploreIcon /> },
  ];

  return (
    <Box role="navigation" sx={{ width: drawerWidth }}>
      {/* Title Section */}
      <Toolbar>
        <Typography
          variant="h6"
          component={Link}
          to="/"
          sx={{
            textDecoration: "none",
            color: "inherit",
            fontWeight: "bold",
            flexGrow: 1,
          }}
        >
          TrustFlow
        </Typography>
      </Toolbar>

      <Divider />

      {/* Menu Items */}
      <List sx={{ py: 0 }}>
        {items.map((item) => {
          const selected = loc.pathname.startsWith(item.to);
          return (
            <ListItemButton
              key={item.to}
              component={Link}
              to={item.to}
              selected={selected}
              onClick={onNavigate}
            >
              <ListItemIcon>{item.icon}</ListItemIcon>
              <ListItemText primary={item.label} />
            </ListItemButton>
          );
        })}
      </List>
    </Box>
  );
}
