import { z } from 'zod'

const timestampSchema = z.iso.datetime({ offset: true })

export const ukOrganizationSchema = z.object({
  id: z.string().min(1),
  inn: z.string().regex(/^\d{10}$|^\d{12}$/),
  ogrn: z.string().nullable(),
  name: z.string().min(1),
  licenseNumber: z.string().nullable(),
})

export const ukDispatcherSchema = z.object({
  id: z.string().min(1),
  fullName: z.string().min(1),
  position: z.string(),
  role: z.literal('uk_dispatcher'),
})

export const ukAuthMethodSchema = z.enum(['esia', 'esia_mock'])

export const ukSessionSchema = z.object({
  token: z.string().min(1),
  expiresAt: timestampSchema,
  authMethod: ukAuthMethodSchema,
  user: ukDispatcherSchema,
  organization: ukOrganizationSchema,
})

export type UkOrganization = z.infer<typeof ukOrganizationSchema>
export type UkDispatcher = z.infer<typeof ukDispatcherSchema>
export type UkAuthMethod = z.infer<typeof ukAuthMethodSchema>
export type UkSession = z.infer<typeof ukSessionSchema>
