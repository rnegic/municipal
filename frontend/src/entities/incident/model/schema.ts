import { z } from 'zod'

import {
  INCIDENT_AUTHORITY_VALUES,
  INCIDENT_CATEGORY_VALUES,
  INCIDENT_DESCRIPTION_MAX_LENGTH,
  INCIDENT_DESCRIPTION_MIN_LENGTH,
  INCIDENT_ENTRANCE_MAX_LENGTH,
  INCIDENT_FLOOR_ZONE_MAX_LENGTH,
  INCIDENT_PHOTOS_MAX_COUNT,
  INCIDENT_REASONING_MAX_LENGTH,
  INCIDENT_TITLE_MAX_LENGTH,
  INCIDENT_TITLE_MIN_LENGTH,
} from '../config/domain'

export const incidentSeveritySchema = z.enum(['critical', 'warning'])

export const incidentStatusSchema = z.enum(['accepted', 'in_progress', 'verifying', 'done'])

export const incidentCategorySchema = z.enum(INCIDENT_CATEGORY_VALUES)

export const incidentAuthoritySchema = z.enum(INCIDENT_AUTHORITY_VALUES)

const timestampSchema = z.iso.datetime({ offset: true })

export const incidentPhotoSchema = z.object({
  id: z.string(),
  url: z.string(),
})

export const incidentSupporterSchema = z.object({
  id: z.string(),
  name: z.string().nullable(),
  avatarUrl: z.string().nullable(),
})

export const incidentSchema = z.object({
  id: z.string(),
  houseId: z.string(),
  title: z.string(),
  description: z.string(),
  category: incidentCategorySchema.nullable(),
  severity: incidentSeveritySchema,
  status: incidentStatusSchema,
  affectedCount: z.number().int().nonnegative(),
  createdAt: timestampSchema,
  dueAt: timestampSchema.nullable(),
  joinedByMe: z.boolean(),
  confirmedByMe: z.boolean(),
  photos: z.array(incidentPhotoSchema),
  supporters: z.array(incidentSupporterSchema).default([]),
  mergedCount: z.number().int().nonnegative().default(0),
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
  confirmedByMe: z.boolean(),
})

export const paginatedResidentRequestsSchema = z.object({
  items: z.array(residentRequestSchema),
  total: z.number().int().nonnegative(),
  offset: z.number().int().nonnegative(),
  limit: z.number().int().nonnegative(),
})

export const ukQueueItemSchema = z.object({
  id: z.string(),
  houseId: z.string(),
  houseAddress: z.string(),
  title: z.string(),
  description: z.string(),
  severity: incidentSeveritySchema,
  status: incidentStatusSchema,
  createdAt: timestampSchema,
  dueAt: timestampSchema.nullable(),
  affectedCount: z.number().int().nonnegative(),
  confirmedCount: z.number().int().nonnegative(),
  reporterName: z.string(),
  photos: z.array(incidentPhotoSchema),
  category: incidentCategorySchema.nullable().default(null),
  mergedCount: z.number().int().nonnegative().default(0),
})

export const paginatedUkQueueSchema = z.object({
  items: z.array(ukQueueItemSchema),
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

export const mergeIncidentsRequestSchema = z.object({
  targetIncidentId: z.string().min(1),
  sourceIncidentIds: z.array(z.string().min(1)).min(1),
})

export const mergeIncidentsResponseSchema = z.object({
  targetIncidentId: z.string(),
  mergedIncidentIds: z.array(z.string()),
  affectedCount: z.number().int().nonnegative(),
})

const optionalTrimmedText = (maxLength: number) =>
  z
    .string()
    .trim()
    .max(maxLength)
    .optional()
    .transform((value) => (value ? value : undefined))

export const analyzeIncidentRequestSchema = z.object({
  description: z
    .string()
    .trim()
    .min(INCIDENT_DESCRIPTION_MIN_LENGTH)
    .max(INCIDENT_DESCRIPTION_MAX_LENGTH),
})

export const analyzeIncidentResponseSchema = z
  .object({
    category: incidentCategorySchema,
    authority: incidentAuthoritySchema,
    isUkResponsibility: z.boolean(),
    photoRequired: z.boolean(),
    reasoningText: z.string().optional(),
  })
  .transform((value) => ({
    ...value,
    reasoningText: value.reasoningText?.trim().slice(0, INCIDENT_REASONING_MAX_LENGTH) || null,
  }))

export const createIncidentRequestSchema = z
  .object({
    title: z
      .string()
      .trim()
      .min(INCIDENT_TITLE_MIN_LENGTH)
      .max(INCIDENT_TITLE_MAX_LENGTH),
    description: z
      .string()
      .trim()
      .min(INCIDENT_DESCRIPTION_MIN_LENGTH)
      .max(INCIDENT_DESCRIPTION_MAX_LENGTH),
    category: incidentCategorySchema,
    entrance: optionalTrimmedText(INCIDENT_ENTRANCE_MAX_LENGTH),
    floorZone: optionalTrimmedText(INCIDENT_FLOOR_ZONE_MAX_LENGTH),
    photoUrls: z.array(z.string().min(1)).max(INCIDENT_PHOTOS_MAX_COUNT).optional(),
  })
  .transform((value) => ({ ...value, photoUrls: value.photoUrls ?? [] }))

export type Incident = z.infer<typeof incidentSchema>
export type IncidentPhoto = z.infer<typeof incidentPhotoSchema>
export type IncidentSupporter = z.infer<typeof incidentSupporterSchema>
export type IncidentSeverity = z.infer<typeof incidentSeveritySchema>
export type IncidentStatus = z.infer<typeof incidentStatusSchema>
export type IncidentCategory = z.infer<typeof incidentCategorySchema>
export type IncidentAuthority = z.infer<typeof incidentAuthoritySchema>
export type IncidentListResponse = z.infer<typeof incidentListResponseSchema>
export type ResidentRequest = z.infer<typeof residentRequestSchema>
export type PaginatedResidentRequests = z.infer<typeof paginatedResidentRequestsSchema>
export type UkQueueItem = z.infer<typeof ukQueueItemSchema>
export type PaginatedUkQueue = z.infer<typeof paginatedUkQueueSchema>
export type JoinResponse = z.infer<typeof joinResponseSchema>
export type ConfirmResponse = z.infer<typeof confirmResponseSchema>
export type MergeIncidentsInput = z.input<typeof mergeIncidentsRequestSchema>
export type MergeIncidentsResult = z.infer<typeof mergeIncidentsResponseSchema>
export type AnalyzeIncidentInput = z.input<typeof analyzeIncidentRequestSchema>
export type AnalyzeIncidentResult = z.infer<typeof analyzeIncidentResponseSchema>
export type CreateIncidentInput = z.input<typeof createIncidentRequestSchema>
