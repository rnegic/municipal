import { Typography } from '@maxhub/max-ui'

import { cn } from '@/shared/lib/cn'
import { incidentTexts } from '../../config/texts'
import type { IncidentStatus } from '../../model/types'
import s from './IncidentStatusBadge.module.scss'

export interface IncidentStatusBadgeProps {
  status: IncidentStatus
  className?: string
}

const STATUS_TO_CLASS: Record<IncidentStatus, string> = {
  accepted: s.accepted,
  in_progress: s.inProgress,
  verifying: s.resolved,
  done: s.resolved,
  false_alarm: s.falseAlarm,
}

export const IncidentStatusBadge = ({ status, className }: IncidentStatusBadgeProps) => (
  <span className={cn(s.root, STATUS_TO_CLASS[status], className)}>
    <Typography.Text variant="note-strong" color="inherit">
      {incidentTexts.statuses[status]}
    </Typography.Text>
  </span>
)
