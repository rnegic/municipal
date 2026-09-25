import { incidentTexts } from '@/entities/incident'
import { Button } from '@/shared/ui/button'

export interface IncidentMergeToggleProps {
  isSelecting: boolean
  onStart: () => void
  onCancel: () => void
  className?: string
}

export const IncidentMergeToggle = ({
  isSelecting,
  onStart,
  onCancel,
  className,
}: IncidentMergeToggleProps) => (
  <Button
    className={className}
    size="small"
    tone={isSelecting ? 'ghost' : 'secondary'}
    aria-pressed={isSelecting}
    onClick={isSelecting ? onCancel : onStart}
  >
    {isSelecting ? incidentTexts.merge.cancelAction : incidentTexts.merge.startAction}
  </Button>
)
