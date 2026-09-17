import type { House } from '../model/schema'

/** «Казань, ул. Баумана 10» */
export const formatHouseAddress = (house: House): string =>
  `${house.city}, ${house.street} ${house.building}`
