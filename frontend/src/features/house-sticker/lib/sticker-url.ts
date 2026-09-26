import type { House } from '@/entities/house'
import { env } from '@/shared/config/env'

export const buildHouseStickerUrl = (house: House): string =>
  new URL(
    `${env.apiBaseUrl}/houses/${encodeURIComponent(house.id)}/sticker`,
    window.location.origin,
  ).href
