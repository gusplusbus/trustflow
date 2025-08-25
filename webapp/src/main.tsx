import React from "react";
import ReactDOM from "react-dom/client";
import { RouterProvider } from "react-router-dom";
import { ThemeProvider, CssBaseline, createTheme } from "@mui/material";
import { router } from "./routes";
import { WalletShell } from "./pages/project/wallet/provider";

const theme = createTheme();

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <WalletShell>
        <RouterProvider router={router} />
      </WalletShell>
    </ThemeProvider>
  </React.StrictMode>
);
