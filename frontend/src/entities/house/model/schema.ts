import { z } from 'zod'

export const houseSchema = z.object({
  id: z.string(),
  address: z.string().min(1),
})

export type House = z.infer<typeof houseSchema>
