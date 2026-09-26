import { INCIDENT_MERGE_MIN_COUNT } from '../config/domain'
import { incidentTexts } from '../config/texts'
import type { UkQueueItem } from '../model/schema'

export const getMergeBlocker = (items: readonly UkQueueItem[]): string | null => {
  if (items.length < INCIDENT_MERGE_MIN_COUNT) {
    return incidentTexts.merge.blockers.tooFew
  }

  if (items.some((item) => item.status === 'done')) {
    return incidentTexts.merge.blockers.closed
  }

  if (items.some((item) => item.houseId !== items[0].houseId)) {
    return incidentTexts.merge.blockers.differentHouses
  }

  return null
}

export const suggestMergeTargetId = (items: readonly UkQueueItem[]): string | null => {
  const [target] = [...items].sort(
    (a, b) =>
      b.affectedCount - a.affectedCount ||
      new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime(),
  )

  return target?.id ?? null
}

export const getMergedAffectedCount = (items: readonly UkQueueItem[]): number =>
  items.reduce((sum, item) => sum + item.affectedCount, 0)
