import { useState } from 'react'

import { incidentTexts } from '@/entities/incident'
import { IconCheckCircle } from '@/shared/assets/icons'
import { Button } from '@/shared/ui/button'

export interface IncidentConfirmButtonProps {
  stretched?: boolean
  className?: string
}

export const IncidentConfirmButton = ({ stretched, className }: IncidentConfirmButtonProps) => {
  const [confirmed, setConfirmed] = useState(false)

  return (
    <Button
      className={className}
      tone="secondary"
      size="small"
      stretched={stretched}
      disabled={confirmed}
      iconBefore={confirmed ? <IconCheckCircle size={16} /> : undefined}
      onClick={() => setConfirmed(true)}
    >
      {confirmed ? incidentTexts.confirm.done : incidentTexts.confirm.action}
    </Button>
  )
}
