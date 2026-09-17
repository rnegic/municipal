import { ProgressSteps } from '@/shared/ui/progress-steps'

import { incidentTexts } from '../../config/texts'
import { getStatusIndex, INCIDENT_STATUS_STEPS } from '../../lib/incident-progress'
import type { IncidentStatus } from '../../model/types'

export interface IncidentStatusTrackProps {
  status: IncidentStatus
  className?: string
}

export const IncidentStatusTrack = ({ status, className }: IncidentStatusTrackProps) => (
  <ProgressSteps
    steps={INCIDENT_STATUS_STEPS}
    currentIndex={getStatusIndex(status)}
    label={incidentTexts.a11y.statusTrack}
    className={className}
  />
)
