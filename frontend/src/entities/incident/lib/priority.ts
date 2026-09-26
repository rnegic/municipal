import { UK_QUEUE_PRIORITY_MIN_AFFECTED } from '../config/domain'
import type { UkQueueItem } from '../model/schema'

const SEVERITY_WEIGHT = { critical: 0, warning: 1 } as const

const getDueTime = (dueAt: string | null): number =>
  dueAt === null ? Number.MAX_SAFE_INTEGER : new Date(dueAt).getTime()

export const sortUkQueueByPriority = (items: readonly UkQueueItem[]): UkQueueItem[] =>
  [...items].sort(
    (a, b) =>
      b.affectedCount - a.affectedCount ||
      SEVERITY_WEIGHT[a.severity] - SEVERITY_WEIGHT[b.severity] ||
      getDueTime(a.dueAt) - getDueTime(b.dueAt) ||
      new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
  )
  
export const isPriorityQueueItem = (item: UkQueueItem): boolean =>
  item.affectedCount >= UK_QUEUE_PRIORITY_MIN_AFFECTED
