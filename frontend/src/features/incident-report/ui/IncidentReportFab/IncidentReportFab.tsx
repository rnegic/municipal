import { useState } from 'react'

import { incidentTexts } from '@/entities/incident'
import { IconPlus } from '@/shared/assets/icons'
import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button'
import { IncidentReportSheet } from '../IncidentReportSheet'
import s from './IncidentReportFab.module.scss'

export interface IncidentReportFabProps {
  className?: string
}

export const IncidentReportFab = ({ className }: IncidentReportFabProps) => {
  const [open, setOpen] = useState(false)

  return (
    <>
      <Button
        className={cn(s.root, className)}
        size="large"
        stretched
        tone="secondary"
        iconBefore={<IconPlus />}
        onClick={() => setOpen(true)}
      >
        {incidentTexts.report.action}
      </Button>
      <IncidentReportSheet open={open} onClose={() => setOpen(false)} />
    </>
  )
}
