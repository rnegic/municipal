import { Typography } from '@maxhub/max-ui'

import {
  incidentTexts,
  type IncidentStatus,
  type UkQueueItem,
  useSetIncidentStatusMutation,
  useUkQueueQuery,
} from '@/entities/incident'
import { EventCreateFab } from '@/features/event-create'
import { UkSessionBadge } from '@/features/uk-auth'
import { formatDateTime } from '@/shared/lib/date'
import { Button } from '@/shared/ui/button'
import { Card } from '@/shared/ui/card'
import { ApiErrorState } from '@/shared/ui/api-error-state'
import { EmptyState } from '@/shared/ui/empty-state'
import { LoadingState } from '@/shared/ui/loading-state'
import { PageLayout } from '@/shared/ui/page-layout'
import s from './UkPage.module.scss'

interface QueueColumn {
  status: IncidentStatus
  title: string
  action?: string
  actionTone: 'primary' | 'secondary'
}

const QUEUE_COLUMNS: readonly QueueColumn[] = [
  { status: 'accepted', title: 'Новые', action: 'Взять в работу', actionTone: 'primary' },
  { status: 'in_progress', title: 'В работе', action: 'Передать на проверку', actionTone: 'secondary' },
  { status: 'verifying', title: 'На проверке у жителей', action: 'Закрыть обращение', actionTone: 'secondary' },
]

const getSlaLabel = (dueAt: string | null): string => {
  if (!dueAt) {
    return 'SLA не задан'
  }

  const minutes = Math.ceil((new Date(dueAt).getTime() - Date.now()) / 60_000)
  if (minutes <= 0) {
    const overdueMinutes = Math.abs(minutes)
    return overdueMinutes >= 60
      ? `Просрочено на ${Math.floor(overdueMinutes / 60)} ч`
      : `Просрочено на ${overdueMinutes} мин`
  }

  return minutes >= 60 ? `Осталось ${Math.ceil(minutes / 60)} ч` : `Осталось ${minutes} мин`
}

const getNextStatus = (status: IncidentStatus): IncidentStatus | null => {
  if (status === 'accepted') {
    return 'in_progress'
  }
  if (status === 'in_progress') {
    return 'verifying'
  }
  if (status === 'verifying') {
    return 'done'
  }
  return null
}

const QueueCard = ({ incident, column }: { incident: UkQueueItem; column: QueueColumn }) => {
  const statusMutation = useSetIncidentStatusMutation()
  const nextStatus = getNextStatus(incident.status)

  return (
    <Card className={s.card} padding="compact">
      <div className={s.cardHead}>
        <Typography.Text variant="body-strong">{incident.title}</Typography.Text>
        <span className={incident.severity === 'critical' ? s.critical : s.warning}>
          {incidentTexts.severity[incident.severity]}
        </span>
      </div>
      <Typography.Text variant="description" color="secondary">
        {incident.description}
      </Typography.Text>
      <div className={s.details}>
        <Typography.Text variant="note" color="tertiary">
          {incident.houseAddress}
        </Typography.Text>
        <Typography.Text variant="note" color="tertiary">
          {incident.reporterName} · {incident.affectedCount} жителей
        </Typography.Text>
        <Typography.Text className={s.sla} variant="note-strong">
          {getSlaLabel(incident.dueAt)}
        </Typography.Text>
        <Typography.Text variant="note" color="tertiary">
          Создано {formatDateTime(incident.createdAt)}
        </Typography.Text>
      </div>
      {column.action && nextStatus ? (
        <Button
          size="medium"
          tone={column.actionTone}
          stretched
          disabled={statusMutation.isPending}
          onClick={() => statusMutation.mutate({ incidentId: incident.id, status: nextStatus })}
        >
          {column.action}
        </Button>
      ) : null}
    </Card>
  )
}

const QueueColumnView = ({ column, items }: { column: QueueColumn; items: readonly UkQueueItem[] }) => (
  <section className={s.column} aria-labelledby={`uk-column-${column.status}`}>
    <div className={s.columnHeader}>
      <Typography.Title id={`uk-column-${column.status}`} variant="medium-strong">
        {column.title}
      </Typography.Title>
      <span className={s.count}>{items.length}</span>
    </div>
    {items.length === 0 ? (
      <EmptyState title="Пусто" description="Здесь пока нет обращений" />
    ) : (
      <div className={s.cards}>
        {items.map((incident) => (
          <QueueCard key={incident.id} incident={incident} column={column} />
        ))}
      </div>
    )}
  </section>
)

export const UkPage = () => {
  const queueQuery = useUkQueueQuery()

  if (queueQuery.isPending) {
    return (
      <PageLayout header={<UkSessionBadge />}>
        <LoadingState />
      </PageLayout>
    )
  }

  if (queueQuery.isError) {
    return (
      <PageLayout header={<UkSessionBadge />}>
        <ApiErrorState error={queueQuery.error} onRetry={() => queueQuery.refetch()} />
      </PageLayout>
    )
  }

  const items = queueQuery.data.items
  const houses = Array.from(
    new Map(
      items.map((item) => [item.houseId, { id: item.houseId, address: item.houseAddress }] as const),
    ).values(),
  )

  return (
    <PageLayout
      header={
        <div className={s.header}>
          <UkSessionBadge />
          <div className={s.headerTitle}>
            <div>
              <Typography.Title variant="large-strong">АРМ диспетчера</Typography.Title>
              <Typography.Text variant="description" color="secondary">
                Очередь обращений по домам вашей УК
              </Typography.Text>
            </div>
            <Typography.Text variant="note" color="tertiary">
              {queueQuery.data.total} активных обращений
            </Typography.Text>
          </div>
        </div>
      }
      floatingAction={houses.length > 0 ? <EventCreateFab houses={houses} /> : undefined}
    >
      <div className={s.board}>
        {QUEUE_COLUMNS.map((column) => (
          <QueueColumnView
            key={column.status}
            column={column}
            items={items.filter((item) => item.status === column.status)}
          />
        ))}
      </div>
    </PageLayout>
  )
}
