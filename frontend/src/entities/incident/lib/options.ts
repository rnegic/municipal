import { INCIDENT_AUTHORITY_VALUES, INCIDENT_CATEGORY_VALUES } from '../config/domain'
import { incidentTexts } from '../config/texts'
import type { IncidentAuthority, IncidentCategory } from '../model/types'

export interface IncidentDomainOption<TValue extends string> {
  value: TValue
  label: string
  description?: string
}

export const incidentCategoryOptions: ReadonlyArray<IncidentDomainOption<IncidentCategory>> =
  INCIDENT_CATEGORY_VALUES.map((value) => ({
    value,
    label: incidentTexts.categories[value],
  }))

export const incidentAuthorityOptions: ReadonlyArray<IncidentDomainOption<IncidentAuthority>> =
  INCIDENT_AUTHORITY_VALUES.map((value) => ({
    value,
    label: incidentTexts.authorities[value].name,
    description: incidentTexts.authorities[value].scope,
  }))
