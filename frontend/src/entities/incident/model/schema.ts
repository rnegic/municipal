import { z } from 'zod'

export const incidentSeveritySchema = z.enum(['critical', 'warning'])

export const incidentStatusSchema = z.enum(['accepted', 'in_progress', 'verifying', 'done'])

export const incidentSchema = z.object({
  id: z.string(),
  houseId: z.string(),
  title: z.string(),
  description: z.string(),
  severity: incidentSeveritySchema,
  status: incidentStatusSchema,
  affectedCount: z.number().int().nonnegative(),
  createdAt: z.string(),
})

export const residentRequestSchema = z.object({
  id: z.string(),
  title: z.string(),
  status: incidentStatusSchema,
  createdAt: z.string(),
  dueAt: z.string().nullable(),
})

export type Incident = z.infer<typeof incidentSchema>
export type IncidentSeverity = z.infer<typeof incidentSeveritySchema>
export type IncidentStatus = z.infer<typeof incidentStatusSchema>
export type ResidentRequest = z.infer<typeof residentRequestSchema>
