import { z } from 'zod'

import { houseSchema } from '@/entities/house'

export const userRoleSchema = z.enum(['resident', 'uk_dispatcher'])

export const userSchema = z.object({
  id: z.string(),
  fullName: z.string(),
  role: userRoleSchema,
})

export const meResponseSchema = z.object({
  user: userSchema,
  house: houseSchema.nullable(),
})

export type User = z.infer<typeof userSchema>
export type UserRole = z.infer<typeof userRoleSchema>
export type MeResponse = z.infer<typeof meResponseSchema>
