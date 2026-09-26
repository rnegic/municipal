import { useCallback, useMemo, useState } from 'react'

import {
  getMergeBlocker,
  getMergedAffectedCount,
  isIncidentClosed,
  suggestMergeTargetId,
  useMergeIncidentsMutation,
  type UkQueueItem,
} from '@/entities/incident'

export interface IncidentMergeModel {
  isSelecting: boolean
  isSheetOpen: boolean
  isPending: boolean
  error: unknown
  selectedItems: readonly UkQueueItem[]
  targetId: string | null
  blocker: string | null
  mergedAffectedCount: number
  isSelected: (incidentId: string) => boolean
  isSelectable: (item: UkQueueItem) => boolean
  startSelecting: () => void
  stopSelecting: () => void
  toggleSelection: (incidentId: string) => void
  setTargetId: (incidentId: string) => void
  openSheet: () => void
  closeSheet: () => void
  submit: () => void
}

export const useIncidentMerge = (items: readonly UkQueueItem[]): IncidentMergeModel => {
  const [isSelecting, setIsSelecting] = useState(false)
  const [isSheetOpen, setIsSheetOpen] = useState(false)
  const [selectedIds, setSelectedIds] = useState<readonly string[]>([])
  const [manualTargetId, setManualTargetId] = useState<string | null>(null)
  const mergeMutation = useMergeIncidentsMutation()

  const selectedItems = useMemo(
    () => items.filter((item) => selectedIds.includes(item.id)),
    [items, selectedIds],
  )

  const blocker = useMemo(() => getMergeBlocker(selectedItems), [selectedItems])

  const targetId = useMemo(
    () =>
      manualTargetId && selectedItems.some((item) => item.id === manualTargetId)
        ? manualTargetId
        : suggestMergeTargetId(selectedItems),
    [manualTargetId, selectedItems],
  )

  const reset = useCallback(() => {
    setIsSelecting(false)
    setIsSheetOpen(false)
    setSelectedIds([])
    setManualTargetId(null)
    mergeMutation.reset()
  }, [mergeMutation])

  const isSelected = useCallback(
    (incidentId: string) => selectedIds.includes(incidentId),
    [selectedIds],
  )

  const isSelectable = useCallback(
    (item: UkQueueItem) =>
      !isIncidentClosed(item.status) &&
      (selectedItems.length === 0 || selectedItems[0].houseId === item.houseId),
    [selectedItems],
  )

  const toggleSelection = useCallback((incidentId: string) => {
    setSelectedIds((current) =>
      current.includes(incidentId)
        ? current.filter((id) => id !== incidentId)
        : [...current, incidentId],
    )
  }, [])

  const submit = useCallback(() => {
    if (blocker || !targetId) {
      return
    }

    mergeMutation.mutate(
      {
        targetIncidentId: targetId,
        sourceIncidentIds: selectedItems
          .filter((item) => item.id !== targetId)
          .map((item) => item.id),
      },
      { onSuccess: reset },
    )
  }, [blocker, mergeMutation, reset, selectedItems, targetId])

  return {
    isSelecting,
    isSheetOpen,
    isPending: mergeMutation.isPending,
    error: mergeMutation.error,
    selectedItems,
    targetId,
    blocker,
    mergedAffectedCount: getMergedAffectedCount(selectedItems),
    isSelected,
    isSelectable,
    startSelecting: () => setIsSelecting(true),
    stopSelecting: reset,
    toggleSelection,
    setTargetId: setManualTargetId,
    openSheet: () => setIsSheetOpen(true),
    closeSheet: () => setIsSheetOpen(false),
    submit,
  }
}
