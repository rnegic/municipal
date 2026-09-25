import { Typography } from '@maxhub/max-ui'

import {
  IncidentStatusBadge,
  IncidentPhotos,
  incidentTexts,
  type Incident,
  type IncidentSeverity,
} from '@/entities/incident'
import { IncidentConfirmButton } from '@/features/incident-confirm'
import { IncidentJoinButton } from '@/features/incident-join'
import { IncidentShareButton } from '@/features/incident-share'
import { IconAlertTriangle } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import { Card } from '@/shared/ui/card'
import s from './IncidentCard.module.scss'

export interface IncidentCardProps {
  incident: Incident
  className?: string
}

const SEVERITY_TO_ICON_CLASS: Record<IncidentSeverity, string | undefined> = {
  critical: s.severityIconCritical,
  warning: undefined,
}

export const IncidentCard = ({ incident, className }: IncidentCardProps) => (
  <Card className={cn(s.root, className)}>
    <div className={s.head}>
      <div className={s.titleRow}>
        <span className={cn(s.severityIcon, SEVERITY_TO_ICON_CLASS[incident.severity])}>
          <IconAlertTriangle size={24} />
        </span>
        <Typography.Text className={s.title} variant="title">
          {incident.title}
        </Typography.Text>
      </div>
      <IncidentStatusBadge status={incident.status} />
    </div>
    <Typography.Text variant="description" color="secondary">
      {incident.description}
    </Typography.Text>
    {incident.mergedCount > 0 ? (
      <Typography.Text variant="note" color="tertiary">
        {incidentTexts.merge.duplicatesBadge(incident.mergedCount)}
      </Typography.Text>
    ) : null}
    <IncidentPhotos photos={incident.photos} />
    <IncidentShareButton incident={incident} className={s.shareButton} />
    {incident.status === 'verifying' ? (
      <div className={s.footer}>
        <IncidentConfirmButton
          incidentId={incident.id}
          alreadyConfirmed={incident.confirmedByMe}
          stretched
        />
      </div>
    ) : (
      <IncidentJoinButton
        incidentId={incident.id}
        affectedCount={incident.affectedCount}
        supporters={incident.supporters}
        alreadyJoined={incident.joinedByMe}
        className={s.join}
      />
    )}
  </Card>
)
