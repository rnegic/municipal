import { Radio, Typography } from '@maxhub/max-ui'

import { incidentTexts, type UkQueueItem } from '@/entities/incident'
import { describeApiError } from '@/shared/lib/api-error'
import { cn } from '@/shared/lib/cn'
import { formatDateTime } from '@/shared/lib/date'
import { BottomSheet } from '@/shared/ui/bottom-sheet'
import { Button } from '@/shared/ui/button'
import { incidentMergeTexts } from '../../config/texts'
import s from './IncidentMergeSheet.module.scss'

export interface IncidentMergeSheetProps {
  open: boolean
  items: readonly UkQueueItem[]
  targetId: string | null
  affectedCount: number
  blocker: string | null
  isPending: boolean
  error: unknown
  onTargetChange: (incidentId: string) => void
  onSubmit: () => void
  onClose: () => void
}

export const IncidentMergeSheet = ({
  open,
  items,
  targetId,
  affectedCount,
  blocker,
  isPending,
  error,
  onTargetChange,
  onSubmit,
  onClose,
}: IncidentMergeSheetProps) => (
  <BottomSheet
    open={open}
    title={incidentTexts.merge.sheetTitle}
    description={incidentTexts.merge.sheetDescription}
    onClose={onClose}
  >
    <fieldset className={s.group}>
      <legend className={s.legend}>
        <Typography.Text variant="note" color="tertiary">
          {incidentTexts.merge.targetLegend}
        </Typography.Text>
      </legend>
      {items.map((item) => {
        const isTarget = item.id === targetId

        return (
          <label key={item.id} className={cn(s.option, isTarget && s.optionTarget)}>
            <Radio
              name="incident-merge-target"
              value={item.id}
              checked={isTarget}
              disabled={isPending}
              onChange={() => onTargetChange(item.id)}
            />
            <span className={s.optionBody}>
              <Typography.Text variant="body-strong">{item.title}</Typography.Text>
              <Typography.Text variant="note" color="secondary">
                {incidentMergeTexts.cardSignatories(item.affectedCount)} ·{' '}
                {formatDateTime(item.createdAt)}
              </Typography.Text>
              <Typography.Text variant="note" color="tertiary">
                {isTarget ? incidentMergeTexts.keepOpen : incidentMergeTexts.becomesDuplicate}
              </Typography.Text>
            </span>
          </label>
        )
      })}
    </fieldset>
    <div className={s.summary}>
      <Typography.Text variant="description-strong">
        {incidentTexts.merge.resultAffected(affectedCount)}
      </Typography.Text>
      <Typography.Text variant="note" color="secondary">
        {incidentTexts.merge.resultDuplicates(Math.max(items.length - 1, 0))}
      </Typography.Text>
      <Typography.Text variant="note" color="tertiary">
        {incidentTexts.merge.targetHint}
      </Typography.Text>
    </div>
    {error ? (
      <Typography.Text className={s.error} variant="note">
        {describeApiError(error).description}
      </Typography.Text>
    ) : null}
    <div className={s.actions}>
      <Button size="medium" tone="secondary" stretched disabled={isPending} onClick={onClose}>
        {incidentMergeTexts.sheetCancel}
      </Button>
      <Button
        size="medium"
        stretched
        loading={isPending}
        disabled={isPending || blocker !== null || targetId === null}
        onClick={onSubmit}
      >
        {incidentTexts.merge.submitAction(items.length)}
      </Button>
    </div>
  </BottomSheet>
)
