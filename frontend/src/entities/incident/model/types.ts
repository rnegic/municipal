export type {
  Incident,
  IncidentSeverity,
  IncidentStatus,
  ResidentRequest,
} from './schema'

export interface IncidentStatusStep {
  id: string
  label: string
}
