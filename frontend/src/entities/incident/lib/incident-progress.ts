import type { IncidentStatus, IncidentStatusStep } from '../model/types'

import { incidentTexts } from '../config/texts'

export const INCIDENT_STATUS_ORDER: readonly IncidentStatus[] = [
  'accepted',
  'in_progress',
  'verifying',
  'done',
]

export const INCIDENT_STATUS_STEPS: readonly IncidentStatusStep[] = INCIDENT_STATUS_ORDER.map(
  (status) => ({ id: status, label: incidentTexts.statuses[status] }),
)

export const getStatusIndex = (status: IncidentStatus): number =>
  Math.max(INCIDENT_STATUS_ORDER.indexOf(status), 0)
