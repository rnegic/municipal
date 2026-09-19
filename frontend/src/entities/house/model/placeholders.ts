import { houseSchema, type House } from './schema'

export const PLACEHOLDER_HOUSE: House = houseSchema.parse({
  id: 'house-1',
  address: 'Казань, ул. Баумана, д. 10',
})
