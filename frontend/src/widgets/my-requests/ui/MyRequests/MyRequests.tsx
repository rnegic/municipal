import { Typography } from '@maxhub/max-ui'

import {
  IncidentStatusBadge,
  incidentTexts,
  type ResidentRequest,
} from '@/entities/incident'
import { IncidentConfirmButton } from '@/features/incident-confirm'
import { IconDocument } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import { formatDate, formatDateTime } from '@/shared/lib/date'
import { Card } from '@/shared/ui/card'
import { EmptyState } from '@/shared/ui/empty-state'
import { Section } from '@/shared/ui/section'
import s from './MyRequests.module.scss'

export interface MyRequestsProps {
  requests: readonly ResidentRequest[]
  className?: string
}

export const MyRequests = ({ requests, className }: MyRequestsProps) => (
  <Section title={incidentTexts.requests.sectionTitle} className={cn(s.root, className)}>
    {requests.length === 0 ? (
      <EmptyState
        icon={<IconDocument size={24} />}
        title={incidentTexts.requests.emptyTitle}
        description={incidentTexts.requests.emptyDescription}
      />
    ) : (
      <ul className={s.list}>
        {requests.map((request) => (
          <li key={request.id}>
            <Card padding="compact">
              <div className={s.head}>
                <Typography.Text className={s.title} variant="body-strong">
                  {request.title}
                </Typography.Text>
                <IncidentStatusBadge status={request.status} />
              </div>
              <div className={s.meta}>
                <Typography.Text variant="note" color="tertiary">
                  {incidentTexts.requests.createdAt(formatDateTime(request.createdAt))}
                </Typography.Text>
                {request.dueAt ? (
                  <Typography.Text variant="note" color="tertiary">
                    {incidentTexts.requests.dueAt(formatDate(request.dueAt))}
                  </Typography.Text>
                ) : null}
              </div>
              {request.status === 'verifying' ? (
                <div className={s.footer}>
                  <IncidentConfirmButton stretched />
                </div>
              ) : null}
            </Card>
          </li>
        ))}
      </ul>
    )}
  </Section>
)
