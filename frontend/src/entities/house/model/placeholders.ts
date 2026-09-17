import { houseSchema, type House } from './schema'

export const PLACEHOLDER_HOUSE: House = houseSchema.parse({
  id: 'house-1',
  city: 'Казань',
  street: 'ул. Баумана',
  building: '10',
  apartment: '12',
})
