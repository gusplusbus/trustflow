import { z } from "zod";

// Shared project schema for FE (and later BE)
export const projectSchema = z.object({
  title: z.string().min(1, "Title is required").max(84, "Max 84 characters"),
  description: z
    .string()
    .min(1, "Description is required")
    .max(221, "Max 221 characters"),
  durationEstimate: z
    .number({ invalid_type_error: "Enter a number" })
    .int("Must be an integer")
    .positive("Must be > 0"),
  teamSize: z
    .number({ invalid_type_error: "Enter a number" })
    .int("Must be an integer")
    .min(1, "Minimum team size is 1"),
  applicationCloseTime: z
    .string()
    .min(1, "Close time is required"),
});

export type ProjectFormValues = z.infer<typeof projectSchema>;
