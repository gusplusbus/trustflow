import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Paper, Box, Stack, Typography, TextField, Button, Alert, Grid,
} from "@mui/material";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { projectSchema, type ProjectFormValues } from "../lib/validation";
import { createProject } from "../lib/projects";

export default function ProjectCreator() {
  const nav = useNavigate();
  const [submitErr, setSubmitErr] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting, isSubmitSuccessful },
    setValue,
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

  const onSubmit = async (data: ProjectFormValues) => {
    setSubmitErr(null);
    try {
      await createProject(data);
      nav("/dashboard"); // or to `/projects/:id` if you return it
    } catch (e: any) {
      setSubmitErr(e?.message || "Failed to create project");
    }
  };

  // Because TextField returns strings, coerce numeric fields:
  const numberReg = (name: "durationEstimate" | "teamSize") => ({
    ...register(name, { valueAsNumber: true }),
    onChange: (e: React.ChangeEvent<HTMLInputElement>) =>
      setValue(name, e.target.value === "" ? ("" as any) : Number(e.target.value), { shouldValidate: true }),
  });

  return (
    <Paper elevation={1} sx={{ p: 3 }}>
      <Box component="form" onSubmit={handleSubmit(onSubmit)}>
        <Stack spacing={2}>
          <div>
            <Typography variant="h6">Create a new project</Typography>
            <Typography variant="body2" color="text.secondary">
              Define the basics. You can add integrations later.
            </Typography>
          </div>

          {submitErr && <Alert severity="error">{submitErr}</Alert>}
          {isSubmitSuccessful && !submitErr && (
            <Alert severity="success">Project created</Alert>
          )}

          <TextField
            label="Title"
            placeholder="e.g., TrustFlow Portal"
            inputProps={{ maxLength: 84 }}
            error={!!errors.title}
            helperText={errors.title?.message}
            {...register("title")}
            fullWidth
          />

          <TextField
            label="Description"
            placeholder="Short summary of goals"
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
            <Button variant="outlined" onClick={() => nav(-1)}>
              Cancel
            </Button>
            <Button type="submit" variant="contained" disabled={isSubmitting}>
              {isSubmitting ? "Saving…" : "Create project"}
            </Button>
          </Stack>
        </Stack>
      </Box>
    </Paper>
  );
}

