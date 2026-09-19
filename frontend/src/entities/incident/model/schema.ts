import { z } from 'zod'

export const incidentSeveritySchema = z.enum(['critical', 'warning'])

export const incidentStatusSchema = z.enum(['accepted', 'in_progress', 'verifying', 'done'])

const timestampSchema = z.iso.datetime({ offset: true })

export const incidentSchema = z.object({
  id: z.string(),
  houseId: z.string(),
  title: z.string(),
  description: z.string(),
  severity: incidentSeveritySchema,
  status: incidentStatusSchema,
  affectedCount: z.number().int().nonnegative(),
  createdAt: timestampSchema,
  joinedByMe: z.boolean(),
  confirmedByMe: z.boolean(),
})

export const incidentListResponseSchema = z.object({
  items: z.array(incidentSchema),
})

export const residentRequestSchema = z.object({
  id: z.string(),
  title: z.string(),
  status: incidentStatusSchema,
  createdAt: timestampSchema,
  dueAt: timestampSchema.nullable(),
})

export const paginatedResidentRequestsSchema = z.object({
  items: z.array(residentRequestSchema),
  total: z.number().int().nonnegative(),
  offset: z.number().int().nonnegative(),
  limit: z.number().int().nonnegative(),
})

export const joinResponseSchema = z.object({
  incidentId: z.string(),
  affectedCount: z.number().int().nonnegative(),
  joined: z.boolean(),
})

export const confirmResponseSchema = z.object({
  incidentId: z.string(),
  status: incidentStatusSchema,
  confirmedAt: timestampSchema,
})

export const createIncidentRequestSchema = z.object({
  title: z.string().min(1).max(120),
  description: z.string().min(1).max(2000),
  severity: incidentSeveritySchema,
  entrance: z.string().optional(),
  riser: z.string().optional(),
})

export type Incident = z.infer<typeof incidentSchema>
export type IncidentSeverity = z.infer<typeof incidentSeveritySchema>
export type IncidentStatus = z.infer<typeof incidentStatusSchema>
export type IncidentListResponse = z.infer<typeof incidentListResponseSchema>
export type ResidentRequest = z.infer<typeof residentRequestSchema>
export type PaginatedResidentRequests = z.infer<typeof paginatedResidentRequestsSchema>
export type JoinResponse = z.infer<typeof joinResponseSchema>
export type ConfirmResponse = z.infer<typeof confirmResponseSchema>
export type CreateIncidentInput = z.infer<typeof createIncidentRequestSchema>
