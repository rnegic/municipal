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
import { AllOkIllustration } from '@/shared/assets/illustrations'
import { cn } from '@/shared/lib/cn'
import { Card } from '@/shared/ui/card'
import { EmptyState } from '@/shared/ui/empty-state'
import { Section, type SectionTone } from '@/shared/ui/section'
import s from './IncidentFocus.module.scss'

export interface IncidentFocusProps {
  incident: Incident | null
  tone?: SectionTone
  className?: string
}

const SEVERITY_TO_ICON_CLASS: Record<IncidentSeverity, string | undefined> = {
  critical: s.severityIconCritical,
  warning: undefined,
}

export const IncidentFocus = ({ incident, tone = 'default', className }: IncidentFocusProps) => (
  <Section title={incidentTexts.focus.sectionTitle} tone={tone} className={cn(s.root, className)}>
    {incident ? (
      <Card className={s.card}>
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
        <IncidentPhotos photos={incident.photos} />
        <IncidentShareButton incident={incident} />
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
            alreadyJoined={incident.joinedByMe}
            className={s.join}
          />
        )}
      </Card>
    ) : (
      <EmptyState
        tone={tone === 'inverse' ? 'inverse' : 'success'}
        illustration={<AllOkIllustration />}
        title={incidentTexts.focus.houseOkEmptyTitle}
        description={incidentTexts.focus.houseOkEmptyDescription}
      />
    )}
  </Section>
)
