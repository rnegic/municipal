import {
  INCIDENT_LEGAL_TERMS,
  INCIDENT_LEGAL_TERM_DEFAULT,
  type IncidentLegalTerm,
} from '../config/legal'
import type { IncidentCategory } from '../model/types'

export const getIncidentLegalTerm = (category: IncidentCategory | null): IncidentLegalTerm =>
  category === null ? INCIDENT_LEGAL_TERM_DEFAULT : INCIDENT_LEGAL_TERMS[category]
