import { Typography } from '@maxhub/max-ui'

import { incidentTexts } from '@/entities/incident'
import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button'
import { incidentMergeTexts } from '../../config/texts'
import s from './IncidentMergeBar.module.scss'

export interface IncidentMergeBarProps {
  selectedCount: number
  blocker: string | null
  onSubmit: () => void
  onCancel: () => void
  className?: string
}

export const IncidentMergeBar = ({
  selectedCount,
  blocker,
  onSubmit,
  onCancel,
  className,
}: IncidentMergeBarProps) => (
  <div className={cn(s.root, className)} role="status">
    <div className={s.summary}>
      <Typography.Text variant="body-strong">
        {incidentTexts.merge.selectedCount(selectedCount)}
      </Typography.Text>
      <Typography.Text variant="note" color="secondary">
        {blocker ?? incidentTexts.merge.selectionHint}
      </Typography.Text>
    </div>
    <div className={s.actions}>
      <Button size="small" tone="ghost" onClick={onCancel}>
        {incidentMergeTexts.sheetCancel}
      </Button>
      <Button size="small" disabled={blocker !== null} onClick={onSubmit}>
        {incidentTexts.merge.submitAction(selectedCount)}
      </Button>
    </div>
  </div>
)
