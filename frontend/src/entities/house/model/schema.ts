import { z } from 'zod'

export const managementCompanySchema = z.object({
  name: z.string().min(1),
  phone: z.string().nullable(),
  website: z.string().nullable(),
  officeAddress: z.string().nullable(),
})

export const houseSchema = z.object({
  id: z.string(),
  address: z.string().min(1),
  uk: managementCompanySchema.optional(),
})

export type House = z.infer<typeof houseSchema>
export type ManagementCompany = z.infer<typeof managementCompanySchema>
