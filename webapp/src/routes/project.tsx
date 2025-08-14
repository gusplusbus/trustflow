import type { RouteObject } from "react-router-dom";
import ProjectCreator from "../pages/project/creator";
import ProjectViewer from "../pages/project/viewer/viewer";
import ProjectEditor from "../pages/project/editor";
import ProjectExplorer from "../pages/project/explorer";

export const projectRoutes: RouteObject[] = [
  { path: "projects/create", element: <ProjectCreator /> },
  { path: "projects/", element: <ProjectExplorer /> },
  { path: "projects/:id", element: <ProjectViewer /> },
  { path: "projects/:id/edit", element: <ProjectEditor /> },
];
