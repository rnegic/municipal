import { z } from 'zod'

export const houseSchema = z.object({
  id: z.string(),
  city: z.string(),
  street: z.string(),
  building: z.string(),
  apartment: z.string().nullable(),
})

export type House = z.infer<typeof houseSchema>
