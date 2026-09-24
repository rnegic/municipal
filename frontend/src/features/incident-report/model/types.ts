import type { IncidentAuthority, IncidentCategory } from '@/entities/incident'

export type IncidentReportStep = 'description' | 'routing' | 'photo'

export type IncidentRoutingSource = 'auto' | 'manual'

export interface IncidentRoutingDecision {
  category: IncidentCategory
  authority: IncidentAuthority
  isUkResponsibility: boolean
  photoRequired: boolean
  reasoningText: string | null
  source: IncidentRoutingSource
}
