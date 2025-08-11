import { createBrowserRouter, Navigate, type RouteObject } from "react-router-dom";
import App from "../App";
import Login from "../pages/Login";
import Register from "../pages/Register";
import Dashboard from "../pages/Dashboard";
import { isAuthed } from "../lib/auth";
import { projectRoutes } from "./project";
import type { JSX } from "react";

function RequireAuth({ children }: { children: JSX.Element }) {
  return isAuthed() ? children : <Navigate to="/" replace />;
}

const children: RouteObject[] = [
  { index: true, element: <Login /> },
  { path: "register", element: <Register /> },
  { path: "dashboard", element: <RequireAuth><Dashboard /></RequireAuth> },

  // Protect all project routes (adjust if you want creator public)
  ...projectRoutes.map(r => ({
    ...r,
    element: <RequireAuth>{r.element as JSX.Element}</RequireAuth>,
  })),

  // optional 404
  { path: "*", element: <div style={{ padding: 16 }}>Not found</div> },
];

export const router = createBrowserRouter([{ path: "/", element: <App />, children }]);
