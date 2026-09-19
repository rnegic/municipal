import { Typography } from '@maxhub/max-ui'

import { incidentTexts, useConfirmIncidentMutation } from '@/entities/incident'
import { IconCheckCircle } from '@/shared/assets/icons'
import { describeApiError } from '@/shared/lib/api-error'
import { Button } from '@/shared/ui/button'
import s from './IncidentConfirmButton.module.scss'

export interface IncidentConfirmButtonProps {
  incidentId: string
  alreadyConfirmed?: boolean
  stretched?: boolean
  className?: string
}

export const IncidentConfirmButton = ({
  incidentId,
  alreadyConfirmed = false,
  stretched,
  className,
}: IncidentConfirmButtonProps) => {
  const confirmMutation = useConfirmIncidentMutation()
  const confirmed = alreadyConfirmed || confirmMutation.isSuccess

  return (
    <div className={s.root}>
      <Button
        className={className}
        tone="secondary"
        size="small"
        stretched={stretched}
        disabled={confirmed || confirmMutation.isPending}
        iconBefore={confirmed ? <IconCheckCircle size={16} /> : undefined}
        onClick={() => confirmMutation.mutate(incidentId)}
      >
        {confirmed ? incidentTexts.confirm.done : incidentTexts.confirm.action}
      </Button>
      {confirmMutation.isError ? (
        <Typography.Text variant="note" color="secondary">
          {describeApiError(confirmMutation.error).description}
        </Typography.Text>
      ) : null}
    </div>
  )
}
