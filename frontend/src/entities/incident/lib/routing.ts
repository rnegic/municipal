import { PHOTO_REQUIRED_CATEGORY_VALUES, UK_AUTHORITY } from '../config/domain'
import type { IncidentAuthority, IncidentCategory } from '../model/types'

const PHOTO_REQUIRED_CATEGORIES: ReadonlySet<string> = new Set(PHOTO_REQUIRED_CATEGORY_VALUES)

export const isPhotoRequiredForCategory = (category: IncidentCategory): boolean =>
  PHOTO_REQUIRED_CATEGORIES.has(category)

export const isUkAuthority = (authority: IncidentAuthority): boolean => authority === UK_AUTHORITY
