import { Counter, Switch, Typography } from '@maxhub/max-ui'
import { useState, type ReactNode } from 'react'

import {
  IncidentPhotos,
  incidentTexts,
  type IncidentStatus,
  type UkQueueItem,
  useSetIncidentStatusMutation,
  useUkQueueQuery,
} from '@/entities/incident'
import { EventCreateFab } from '@/features/event-create'
import { ThemeToggle } from '@/features/theme-switch'
import { UkSessionBadge, UkSignOutButton } from '@/features/uk-auth'
import { IconLocation } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
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
  prevAction?: string
}

const QUEUE_COLUMNS: readonly QueueColumn[] = [
  { status: 'accepted', title: 'Новые', action: 'Взять в работу', actionTone: 'primary' },
  {
    status: 'in_progress',
    title: 'В работе',
    action: 'Передать на проверку',
    actionTone: 'secondary',
    prevAction: 'В новые',
  },
  {
    status: 'verifying',
    title: 'На проверке у жителей',
    action: 'Закрыть обращение',
    actionTone: 'secondary',
    prevAction: 'В работу',
  },
]

const ARCHIVE_COLUMN: QueueColumn = {
  status: 'done',
  title: 'Закрыто',
  actionTone: 'secondary',
  prevAction: 'Вернуть в проверку',
}

const getSlaLabel = (dueAt: string | null): string => {
  if (!dueAt) {
    return 'Не задан'
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

const getPrevStatus = (status: IncidentStatus): IncidentStatus | null => {
  if (status === 'in_progress') {
    return 'accepted'
  }
  if (status === 'verifying') {
    return 'in_progress'
  }
  if (status === 'done') {
    return 'verifying'
  }
  return null
}

const isOverdue = (dueAt: string | null): boolean => dueAt !== null && new Date(dueAt).getTime() <= Date.now()

const getCountTone = (count: number): string => {
  if (count > 7) {
    return s.countHigh
  }
  if (count > 3) {
    return s.countMedium
  }
  return s.countLow
}

const getColumnTone = (status: IncidentStatus): string => {
  if (status === 'in_progress') {
    return s.columnTitleProgress
  }
  if (status === 'verifying') {
    return s.columnTitleVerify
  }
  return s.columnTitleNew
}

const QueueCard = ({
  incident,
  column,
  isPending,
  onMove,
}: {
  incident: UkQueueItem
  column: QueueColumn
  isPending: boolean
  onMove: (incidentId: string, status: IncidentStatus) => void
}) => {
  const nextStatus = getNextStatus(incident.status)
  const prevStatus = getPrevStatus(incident.status)

  return (
    <Card
      className={s.card}
      padding="compact"
      draggable
      onDragStart={(event) => {
        event.dataTransfer.effectAllowed = 'move'
        event.dataTransfer.setData(
          'text/plain',
          JSON.stringify({ id: incident.id, status: incident.status }),
        )
      }}
    >
      <div className={s.cardHead}>
        <Typography.Text variant="body-strong">{incident.title}</Typography.Text>
        <span className={incident.severity === 'critical' ? s.critical : s.warning}>
          {incidentTexts.severity[incident.severity]}
        </span>
      </div>
      <Typography.Text variant="description" color="secondary">
        {incident.description}
      </Typography.Text>
      <IncidentPhotos photos={incident.photos} />
      <div className={s.details}>
        <div className={s.addressRow}>
          <IconLocation size={16} className={s.addressIcon} />
          <Typography.Text className={s.address} variant="note" color="tertiary">
            {incident.houseAddress}
          </Typography.Text>
        </div>
        <div className={s.meta}>
          <div className={s.metaItem}>
            <Typography.Text variant="note" color="tertiary">
              Заявитель
            </Typography.Text>
            <Typography.Text variant="note-strong">{incident.reporterName}</Typography.Text>
          </div>
          <div className={s.metaItem}>
            <Typography.Text variant="note" color="tertiary">
              Жителей
            </Typography.Text>
            <Typography.Text variant="note-strong">{incident.affectedCount}</Typography.Text>
          </div>
          <div className={s.metaItem}>
            <Typography.Text variant="note" color="tertiary">
              Срок
            </Typography.Text>
            <Typography.Text
              className={isOverdue(incident.dueAt) ? s.slaOverdue : undefined}
              variant="note-strong"
            >
              {getSlaLabel(incident.dueAt)}
            </Typography.Text>
          </div>
          <div className={s.metaItem}>
            <Typography.Text variant="note" color="tertiary">
              Создано
            </Typography.Text>
            <Typography.Text variant="note-strong">{formatDateTime(incident.createdAt)}</Typography.Text>
          </div>
        </div>
      </div>
      {column.prevAction || (column.action && nextStatus) ? (
        <div className={s.actions}>
          {column.prevAction && prevStatus ? (
            <Button
              size="small"
              tone="secondary"
              stretched
              disabled={isPending}
              onClick={() => onMove(incident.id, prevStatus)}
            >
              {column.prevAction}
            </Button>
          ) : null}
          {column.action && nextStatus ? (
            <Button
              size="small"
              tone={column.actionTone}
              stretched
              disabled={isPending}
              onClick={() => onMove(incident.id, nextStatus)}
            >
              {column.action}
            </Button>
          ) : null}
        </div>
      ) : null}
    </Card>
  )
}

const QueueColumnView = ({
  column,
  items,
  isPending,
  onMove,
}: {
  column: QueueColumn
  items: readonly UkQueueItem[]
  isPending: boolean
  onMove: (incidentId: string, status: IncidentStatus) => void
}) => {
  const [isOver, setIsOver] = useState(false)

  return (
    <section
      className={cn(s.column, isOver && s.columnOver)}
      aria-labelledby={`uk-column-${column.status}`}
      onDragOver={(event) => {
        event.preventDefault()
        if (!isOver) {
          setIsOver(true)
        }
      }}
      onDragLeave={() => setIsOver(false)}
      onDrop={(event) => {
        event.preventDefault()
        setIsOver(false)
        const raw = event.dataTransfer.getData('text/plain')
        if (!raw) {
          return
        }
        const dragged = JSON.parse(raw) as { id: string; status: IncidentStatus }
        if (dragged.status !== column.status) {
          onMove(dragged.id, column.status)
        }
      }}
    >
      <div className={s.columnHeader}>
        <div className={cn(s.columnTitle, getColumnTone(column.status))} id={`uk-column-${column.status}`}>
          <Typography.Title className={s.columnTitleText} variant="small-strong">
            {column.title}
          </Typography.Title>
        </div>
        <Counter className={getCountTone(items.length)} value={items.length} variant="mute" />
      </div>
      {items.length === 0 ? (
        <EmptyState title="Пусто" description="Здесь пока нет обращений" />
      ) : (
        <div className={s.cards}>
          {items.map((incident) => (
            <QueueCard
              key={incident.id}
              incident={incident}
              column={column}
              isPending={isPending}
              onMove={onMove}
            />
          ))}
        </div>
      )}
    </section>
  )
}

const UkHero = ({ children }: { children: ReactNode }) => (
  <div className={s.hero}>
    <div className={s.heroInner}>{children}</div>
  </div>
)

export const UkPage = () => {
  const queueQuery = useUkQueueQuery()
  const statusMutation = useSetIncidentStatusMutation()
  const [showArchive, setShowArchive] = useState(false)

  const moveIncident = (incidentId: string, status: IncidentStatus): void => {
    statusMutation.mutate({ incidentId, status })
  }

  const sessionHeader = (
    <UkSessionBadge
      actions={
        <>
          <ThemeToggle />
          <UkSignOutButton />
        </>
      }
    />
  )

  if (queueQuery.isPending) {
    return (
      <PageLayout width="wide" hero={<UkHero>{sessionHeader}</UkHero>}>
        <LoadingState />
      </PageLayout>
    )
  }

  if (queueQuery.isError) {
    return (
      <PageLayout width="wide" hero={<UkHero>{sessionHeader}</UkHero>}>
        <ApiErrorState error={queueQuery.error} onRetry={() => queueQuery.refetch()} />
      </PageLayout>
    )
  }

  const items = queueQuery.data.items
  const activeItems = items.filter((item) => item.status !== 'done')
  const closedItems = items.filter((item) => item.status === 'done')
  const houses = Array.from(
    new Map(
      items.map((item) => [item.houseId, { id: item.houseId, address: item.houseAddress }] as const),
    ).values(),
  )

  return (
    <PageLayout
      width="wide"
      hero={
        <UkHero>
          <div className={s.header}>
            {sessionHeader}
            <div className={s.headerTitle}>
              <div className={s.headerBody}>
                <Typography.Title className={s.heroTitle} variant="large-strong">
                  АРМ диспетчера
                </Typography.Title>
                <Typography.Text className={s.heroSubtitle} variant="description" color="secondary">
                  Очередь обращений по домам вашей УК
                </Typography.Text>
              </div>
              <Typography.Text className={s.heroMeta} variant="note" color="tertiary">
                {activeItems.length} активных обращений
              </Typography.Text>
            </div>
          </div>
        </UkHero>
      }
      floatingAction={houses.length > 0 ? <EventCreateFab houses={houses} /> : undefined}
    >
      <div className={s.boardToolbar}>
        <span className={s.archiveToggle}>
          <Typography.Text variant="note" color="secondary">
            Архив закрытых ({closedItems.length})
          </Typography.Text>
          <Switch
            checked={showArchive}
            aria-label="Показать архив закрытых"
            onChange={(event) => setShowArchive(event.target.checked)}
          />
        </span>
      </div>
      {showArchive ? (
        closedItems.length === 0 ? (
          <EmptyState title="Архив пуст" description="Закрытые обращения появятся здесь" />
        ) : (
          <div className={s.archive}>
            {closedItems.map((incident) => (
              <QueueCard
                key={incident.id}
                incident={incident}
                column={ARCHIVE_COLUMN}
                isPending={statusMutation.isPending}
                onMove={moveIncident}
              />
            ))}
          </div>
        )
      ) : (
        <div className={s.board}>
          {QUEUE_COLUMNS.map((column) => (
            <QueueColumnView
              key={column.status}
              column={column}
              items={items.filter((item) => item.status === column.status)}
              isPending={statusMutation.isPending}
              onMove={moveIncident}
            />
          ))}
        </div>
      )}
    </PageLayout>
  )
}
