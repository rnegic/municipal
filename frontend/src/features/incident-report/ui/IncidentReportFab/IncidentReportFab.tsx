import { useNavigate } from 'react-router-dom'

import { incidentTexts } from '@/entities/incident'
import { IconPlus } from '@/shared/assets/icons'
import { ROUTES } from '@/shared/config/routes'
import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button'
import s from './IncidentReportFab.module.scss'

export interface IncidentReportFabProps {
  className?: string
}

export const IncidentReportFab = ({ className }: IncidentReportFabProps) => {
  const navigate = useNavigate()

  return (
    <Button
      className={cn(s.root, className)}
      size="large"
      stretched
      tone="secondary"
      iconBefore={<IconPlus />}
      onClick={() => navigate(ROUTES.incidentCreate)}
    >
      {incidentTexts.report.action}
    </Button>
  )
}
