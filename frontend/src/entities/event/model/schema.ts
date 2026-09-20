import { z } from 'zod'

export const eventStatusSchema = z.enum([
  'planned',
  'in_progress',
  'awaiting_confirmation',
  'disputed',
  'closed',
])

const timestampSchema = z.iso.datetime({ offset: true })

export const eventSchema = z.object({
  id: z.string(),
  houseId: z.string(),
  entrance: z.string().nullable(),
  riser: z.string().nullable(),
  reason: z.string(),
  responsible: z.string(),
  scheduledFrom: timestampSchema,
  scheduledTo: timestampSchema,
  status: eventStatusSchema,
  createdAt: timestampSchema,
  resolvedAt: timestampSchema.nullable(),
})

export const eventListResponseSchema = z.object({
  items: z.array(eventSchema),
})

export const createEventRequestSchema = z.object({
  houseId: z.string().min(1),
  reason: z.string().min(1).max(120),
  responsible: z.string().min(1).max(80),
  entrance: z.string().optional(),
  riser: z.string().optional(),
  scheduledFrom: timestampSchema,
  scheduledTo: timestampSchema,
})

export type Event = z.infer<typeof eventSchema>
export type EventStatus = z.infer<typeof eventStatusSchema>
export type EventListResponse = z.infer<typeof eventListResponseSchema>
export type CreateEventInput = z.infer<typeof createEventRequestSchema>
