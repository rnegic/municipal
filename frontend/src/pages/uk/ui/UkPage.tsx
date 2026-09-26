import { Counter, Switch, Typography } from '@maxhub/max-ui'
import { useState, type ReactNode } from 'react'

import {
  CategoryIcon,
  IncidentPhotos,
  incidentTexts,
  isPriorityQueueItem,
  sortUkQueueByPriority,
  type IncidentStatus,
  type UkQueueItem,
  useSetIncidentStatusMutation,
  useUkQueueQuery,
} from '@/entities/incident'
import { EventCreateFab } from '@/features/event-create'
import {
  IncidentMergeBar,
  IncidentMergeSheet,
  IncidentMergeToggle,
  useIncidentMerge,
} from '@/features/incident-merge'
import { ThemeToggle } from '@/features/theme-switch'
import { ResidentModeLink, UkSessionBadge, UkSignOutButton } from '@/features/uk-auth'
import { IconLocation } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import { formatDateTime } from '@/shared/lib/date'
import { Button } from '@/shared/ui/button'
import { Card } from '@/shared/ui/card'
import { Checkbox } from '@/shared/ui/checkbox'
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

interface QueueSelection {
  isActive: boolean
  isSelected: boolean
  isSelectable: boolean
  onToggle: () => void
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
  selection,
  onMove,
}: {
  incident: UkQueueItem
  column: QueueColumn
  isPending: boolean
  selection: QueueSelection
  onMove: (incidentId: string, status: IncidentStatus) => void
}) => {
  const nextStatus = getNextStatus(incident.status)
  const prevStatus = getPrevStatus(incident.status)
  const isPriority = isPriorityQueueItem(incident)
  const isDimmed = selection.isActive && !selection.isSelectable && !selection.isSelected
  const canToggle = selection.isActive && (selection.isSelectable || selection.isSelected)

  return (
    <Card
      className={cn(
        s.card,
        selection.isActive && s.cardSelecting,
        isPriority && s.cardPriority,
        selection.isSelected && s.cardSelected,
        isDimmed && s.cardDimmed,
      )}
      padding="compact"
      draggable={!selection.isActive}
      onClick={canToggle ? selection.onToggle : undefined}
      onDragStart={(event) => {
        event.dataTransfer.effectAllowed = 'move'
        event.dataTransfer.setData(
          'text/plain',
          JSON.stringify({ id: incident.id, status: incident.status }),
        )
      }}
    >
      <div className={s.content}>
        <div className={s.cardHead}>
          {selection.isActive ? (
            <Checkbox
              checked={selection.isSelected}
              disabled={!selection.isSelectable && !selection.isSelected}
              aria-label={incidentTexts.queue.selectCard}
              onChange={selection.onToggle}
              onClick={(event) => event.stopPropagation()}
            />
          ) : null}
          <div className={s.cardTitleRow}>
            <CategoryIcon category={incident.category} size={18} className={s.categoryIcon} />
            <Typography.Text className={s.cardTitle} variant="body-strong">
              {incident.title}
            </Typography.Text>
          </div>
          <span className={incident.severity === 'critical' ? s.critical : s.warning}>
            {incidentTexts.severity[incident.severity]}
          </span>
        </div>
        {isPriority || incident.mergedCount > 0 ? (
          <div className={s.badges}>
            {isPriority ? (
              <span className={s.priority}>
                {incidentTexts.queue.priorityBadge} · {incident.affectedCount}
              </span>
            ) : null}
            {incident.mergedCount > 0 ? (
              <span className={s.merged}>{incidentTexts.merge.duplicatesBadge(incident.mergedCount)}</span>
            ) : null}
          </div>
        ) : null}
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
                {incidentTexts.queue.reporter}
              </Typography.Text>
              <Typography.Text variant="note-strong">{incident.reporterName}</Typography.Text>
            </div>
            <div className={s.metaItem}>
              <Typography.Text variant="note" color="tertiary">
                {incidentTexts.queue.signatories}
              </Typography.Text>
              <Typography.Text className={isPriority ? s.signatories : undefined} variant="note-strong">
                {incident.affectedCount}
              </Typography.Text>
            </div>
            <div className={s.metaItem}>
              <Typography.Text variant="note" color="tertiary">
                {incidentTexts.queue.due}
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
                {incidentTexts.queue.createdAt}
              </Typography.Text>
              <Typography.Text variant="note-strong">{formatDateTime(incident.createdAt)}</Typography.Text>
            </div>
          </div>
        </div>
        {!selection.isActive && (column.prevAction || (column.action && nextStatus)) ? (
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
      </div>
    </Card>
  )
}

const QueueColumnView = ({
  column,
  items,
  isPending,
  isSelecting,
  getSelection,
  onMove,
}: {
  column: QueueColumn
  items: readonly UkQueueItem[]
  isPending: boolean
  isSelecting: boolean
  getSelection: (incident: UkQueueItem) => QueueSelection
  onMove: (incidentId: string, status: IncidentStatus) => void
}) => {
  const [isOver, setIsOver] = useState(false)

  return (
    <section
      className={cn(s.column, isOver && s.columnOver)}
      aria-labelledby={`uk-column-${column.status}`}
      onDragOver={(event) => {
        if (isSelecting) {
          return
        }
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
              selection={getSelection(incident)}
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
  const merge = useIncidentMerge(queueQuery.data?.items ?? [])

  const moveIncident = (incidentId: string, status: IncidentStatus): void => {
    statusMutation.mutate({ incidentId, status })
  }

  const getSelection = (incident: UkQueueItem): QueueSelection => ({
    isActive: merge.isSelecting,
    isSelected: merge.isSelected(incident.id),
    isSelectable: merge.isSelectable(incident),
    onToggle: () => merge.toggleSelection(incident.id),
  })

  const sessionHeader = (
    <UkSessionBadge
      actions={
        <div className={s.sessionActions}>
          <div className={s.sessionIcons}>
            <ThemeToggle />
            <UkSignOutButton />
          </div>
          <ResidentModeLink />
        </div>
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
  const closedItems = sortUkQueueByPriority(items.filter((item) => item.status === 'done'))
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
      floatingAction={
        houses.length > 0 && !merge.isSelecting ? <EventCreateFab houses={houses} /> : undefined
      }
    >
      <div className={s.boardToolbar}>
        <Typography.Text className={s.sortHint} variant="note" color="tertiary">
          {incidentTexts.queue.sortHint}
        </Typography.Text>
        <div className={s.toolbarActions}>
          <IncidentMergeToggle
            isSelecting={merge.isSelecting}
            onStart={merge.startSelecting}
            onCancel={merge.stopSelecting}
          />
          <span className={s.archiveToggle}>
            <Typography.Text variant="note" color="secondary">
              Архив закрытых ({closedItems.length})
            </Typography.Text>
            <Switch
              checked={showArchive}
              className={s.archiveSwitch}
              aria-label="Показать архив закрытых"
              onChange={(event) => setShowArchive(event.target.checked)}
            />
          </span>
        </div>
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
                selection={getSelection(incident)}
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
              items={sortUkQueueByPriority(items.filter((item) => item.status === column.status))}
              isPending={statusMutation.isPending}
              isSelecting={merge.isSelecting}
              getSelection={getSelection}
              onMove={moveIncident}
            />
          ))}
        </div>
      )}
      {merge.isSelecting ? (
        <IncidentMergeBar
          selectedCount={merge.selectedItems.length}
          blocker={merge.blocker}
          onSubmit={merge.openSheet}
          onCancel={merge.stopSelecting}
        />
      ) : null}
      <IncidentMergeSheet
        open={merge.isSheetOpen}
        items={merge.selectedItems}
        targetId={merge.targetId}
        affectedCount={merge.mergedAffectedCount}
        blocker={merge.blocker}
        isPending={merge.isPending}
        error={merge.error}
        onTargetChange={merge.setTargetId}
        onSubmit={merge.submit}
        onClose={merge.closeSheet}
      />
    </PageLayout>
  )
}
