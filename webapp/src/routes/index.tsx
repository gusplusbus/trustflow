import { createBrowserRouter, Navigate } from "react-router-dom";
import App from "../App";
import Login from "../pages/Login";
import Register from "../pages/Register";
import Dashboard from "../pages/Dashboard";
import { isAuthed } from "../lib/auth";
import ProjectCreator from "../pages/ProjectCreator";
import type { JSX } from "react";

function RequireAuth({ children }: { children: JSX.Element }) {
  return isAuthed() ? children : <Navigate to="/" replace />;
}

export const router = createBrowserRouter([
  {
    path: "/",
    element: <App />,
    children: [
      { index: true, element: <Login /> },
      { path: "register", element: <Register /> },
      { path: "dashboard", element: <RequireAuth><Dashboard /></RequireAuth> },
      { path: "projects/create", element: <ProjectCreator /> },
    ],
  },
]);
