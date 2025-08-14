import * as React from "react";
import { useParams } from "react-router-dom";
import { Alert, Stack } from "@mui/material";
import ProjectInfo from "./project_info";
import ConnectRepo from "./connect_repo";
import ListOriginIssues from "./list_origin_issues";
import { useDeleteProject, useProject } from "../../../hooks/project";

export default function Viewer() {
  const { id = "" } = useParams();
  const { data, loading, error, reload } = useProject(id);
  const { remove, loading: deleting, error: delErr } = useDeleteProject(id);

  if (loading) return <>Loading…</>;
  if (error)   return <Alert severity="error">{error}</Alert>;
  if (!data)   return <Alert severity="warning">Not found</Alert>;

  // Ownerships are already hydrated by useProject(..., include_ownerships=true)
  const ownerships = data.ownerships ?? [];

  return (
    <Stack spacing={2}>
      <ProjectInfo
        id={id}
        data={data}
        deleting={deleting}
        delErr={delErr || null}
        onDelete={remove}
      />

      <ConnectRepo
        projectId={id}
        ownerships={ownerships}
        onOwnershipSaved={reload}   // re-fetch the project (and ownerships) on save
      />

      <ListOriginIssues projectId={id} />
    </Stack>
  );
}
