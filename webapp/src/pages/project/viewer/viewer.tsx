import * as React from "react";
import { useParams } from "react-router-dom";
import { Alert, Stack } from "@mui/material";
import ProjectInfo from "./project_info";
import ConnectRepo, { type Ownership } from "./connect_repo";
import ListOriginIssues from "./list_origin_issues";
import { useDeleteProject, useProject } from "../../../hooks/project";
import { getProject, type ProjectResponse } from "../../../lib/projects";

export default function Viewer() {
  const { id = "" } = useParams();
  const { data, loading, error } = useProject(id);
  const { remove, loading: deleting, error: delErr } = useDeleteProject(id);

  const [ownerships, setOwnerships] = React.useState<Ownership[]>([]);
  const reloadOwnerships = React.useCallback(async () => {
    if (!id) return;
    const full: ProjectResponse & { ownerships?: Ownership[] } =
      await getProject(id, { includeOwnerships: true } as any);
    setOwnerships(full.ownerships ?? []);
  }, [id]);

  React.useEffect(() => { reloadOwnerships(); }, [reloadOwnerships]);

  if (loading) return <>Loading…</>;
  if (error) return <Alert severity="error">{error}</Alert>;
  if (!data)  return <Alert severity="warning">Not found</Alert>;

  const first = ownerships[0];
  const defOwner = first?.organization ?? "";
  const defRepo  = first?.repository ?? "";

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
        onOwnershipSaved={reloadOwnerships}
      />

      <ListOriginIssues
        projectId={id}
        defaultOwner={defOwner}
        defaultRepo={defRepo}
      />
    </Stack>
  );
}
