import { Typography } from '@maxhub/max-ui'

import {
  eventTexts,
  formatEventSchedule,
  selectActiveEvents,
  useEventsQuery,
  type Event,
} from '@/entities/event'
import { Card } from '@/shared/ui/card'
import { Section, type SectionTone } from '@/shared/ui/section'
import s from './EventBulletin.module.scss'

export interface EventBulletinProps {
  houseId: string
  tone?: SectionTone
  className?: string
}

const formatScope = (event: Event): string =>
  [
    event.entrance ? eventTexts.entrance(event.entrance) : null,
    event.riser,
    event.responsible,
  ]
    .filter((part): part is string => Boolean(part))
    .join(' · ')

export const EventBulletin = ({ houseId, tone = 'default', className }: EventBulletinProps) => {
  const eventsQuery = useEventsQuery(houseId)
  const events = selectActiveEvents(eventsQuery.data?.items ?? [])

  if (events.length === 0) {
    return null
  }

  return (
    <Section title={eventTexts.sectionTitle} tone={tone} className={className}>
      <ul className={s.list}>
        {events.map((event) => (
          <li key={event.id}>
            <Card tone="progress" padding="compact">
              <Typography.Text variant="body-strong">
                {formatEventSchedule(event.scheduledFrom, event.scheduledTo)}
              </Typography.Text>
              <Typography.Text variant="description" color="secondary">
                {event.reason}
              </Typography.Text>
              <Typography.Text variant="note" color="tertiary">
                {formatScope(event)}
              </Typography.Text>
            </Card>
          </li>
        ))}
      </ul>
    </Section>
  )
}
