import { houseSchema, type House } from '../model/schema'

const BOUND_HOUSE_STORAGE_KEY = 'municipal.bound-house'

export const getBoundHouse = (): House | null => {
  try {
    const storedHouse = window.localStorage.getItem(BOUND_HOUSE_STORAGE_KEY)

    if (!storedHouse) {
      return null
    }

    const result = houseSchema.safeParse(JSON.parse(storedHouse))

    return result.success ? result.data : null
  } catch {
    return null
  }
}

export const saveBoundHouse = (house: House): void => {
  try {
    window.localStorage.setItem(BOUND_HOUSE_STORAGE_KEY, JSON.stringify(houseSchema.parse(house)))
  } catch {
    return
  }
}
