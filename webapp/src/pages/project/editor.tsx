import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { Paper, Box, Stack, Typography, TextField, Button, Alert, Grid } from "@mui/material";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { projectSchema, type ProjectFormValues } from "../../lib/projects";
import { useProject, useUpdateProject } from "../../hooks/project";

export default function ProjectEditor() {
  const { id = "" } = useParams();
  const nav = useNavigate();
  const { data, loading, error } = useProject(id);
  const { update, loading: saving, error: saveErr } = useUpdateProject(id);
  const [initErr, setInitErr] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
    setValue,
    reset,
  } = useForm<ProjectFormValues>({
    resolver: zodResolver(projectSchema),
    defaultValues: {
      title: "",
      description: "",
      durationEstimate: 1,
      teamSize: 1,
      applicationCloseTime: "",
    },
  });

  useEffect(() => {
    if (!loading && data) {
      reset({
        title: data.title,
        description: data.description,
        durationEstimate: data.duration_estimate,
        teamSize: data.team_size,
        applicationCloseTime: data.application_close_time,
      });
    }
    if (!loading && !data && !error) setInitErr("Project not found");
  }, [data, loading, error, reset]);

  const onSubmit = async (form: ProjectFormValues) => {
    try {
      await update(form);
      nav(`/projects/${id}`);
    } catch {}
  };

  const numberReg = (name: "durationEstimate" | "teamSize") => ({
    ...register(name, { valueAsNumber: true }),
    onChange: (e: React.ChangeEvent<HTMLInputElement>) =>
      setValue(name, e.target.value === "" ? ("" as any) : Number(e.target.value), { shouldValidate: true }),
  });

  if (loading) return <>Loading…</>;
  if (error) return <Alert severity="error">{error}</Alert>;
  if (initErr) return <Alert severity="warning">{initErr}</Alert>;

  return (
    <Paper elevation={1} sx={{ p: 3 }}>
      <Box component="form" onSubmit={handleSubmit(onSubmit)}>
        <Stack spacing={2}>
          <div>
            <Typography variant="h6">Edit project</Typography>
            <Typography variant="body2" color="text.secondary">Update the details below.</Typography>
          </div>

          {saveErr && <Alert severity="error">{saveErr}</Alert>}

          <TextField
            label="Title"
            inputProps={{ maxLength: 84 }}
            error={!!errors.title}
            helperText={errors.title?.message}
            {...register("title")}
            fullWidth
          />

          <TextField
            label="Description"
            inputProps={{ maxLength: 221 }}
            multiline
            minRows={3}
            error={!!errors.description}
            helperText={errors.description?.message}
            {...register("description")}
            fullWidth
          />

          <Grid container spacing={2}>
            <Grid item xs={12} sm={4}>
              <TextField
                label="Duration estimate (days)"
                type="number"
                inputProps={{ min: 1 }}
                error={!!errors.durationEstimate}
                helperText={errors.durationEstimate?.message}
                {...numberReg("durationEstimate")}
                fullWidth
              />
            </Grid>

            <Grid item xs={12} sm={4}>
              <TextField
                label="Team size"
                type="number"
                inputProps={{ min: 1 }}
                error={!!errors.teamSize}
                helperText={errors.teamSize?.message}
                {...numberReg("teamSize")}
                fullWidth
              />
            </Grid>

            <Grid item xs={12} sm={4}>
              <TextField
                label="Application close time"
                type="datetime-local"
                error={!!errors.applicationCloseTime}
                helperText={errors.applicationCloseTime?.message}
                {...register("applicationCloseTime")}
                fullWidth
                InputLabelProps={{ shrink: true }}
              />
            </Grid>
          </Grid>

          <Stack direction="row" spacing={1} justifyContent="flex-end">
            <Button variant="outlined" onClick={() => nav(-1)}>Cancel</Button>
            <Button type="submit" variant="contained" disabled={saving}>
              {saving ? "Saving…" : "Save changes"}
            </Button>
          </Stack>
        </Stack>
      </Box>
    </Paper>
  );
}
