import { z } from 'zod'

const timestampSchema = z.iso.datetime({ offset: true })

export const apiKeySchema = z.object({
  id: z.string(),
  name: z.string(),
  prefix: z.string(),
  createdAt: timestampSchema,
  lastUsedAt: timestampSchema.nullable(),
  revokedAt: timestampSchema.nullable(),
})

export const createdApiKeySchema = apiKeySchema.extend({
  key: z.string(),
})

export const apiKeyListSchema = z.object({
  items: z.array(apiKeySchema),
})

export type ApiKey = z.infer<typeof apiKeySchema>
export type CreatedApiKey = z.infer<typeof createdApiKeySchema>
