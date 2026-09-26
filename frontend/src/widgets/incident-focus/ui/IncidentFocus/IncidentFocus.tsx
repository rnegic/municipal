import { Link } from 'react-router-dom'

import { incidentTexts, type Incident } from '@/entities/incident'
import { AllOkIllustration } from '@/shared/assets/illustrations'
import { cn } from '@/shared/lib/cn'
import { Button } from '@/shared/ui/button'
import { EmptyState } from '@/shared/ui/empty-state'
import { Section, type SectionTone } from '@/shared/ui/section'
import { IncidentCard } from '../IncidentCard'
import s from './IncidentFocus.module.scss'

export interface IncidentFocusProps {
  incidents: readonly Incident[]
  tone?: SectionTone
  allIncidentsHref?: string
  className?: string
}

export const IncidentFocus = ({
  incidents,
  tone = 'default',
  allIncidentsHref,
  className,
}: IncidentFocusProps) => {
  const [incident] = incidents
  const hasMore = incidents.length > 1 && allIncidentsHref !== undefined

  return (
    <Section
      title={incidentTexts.focus.sectionTitle}
      tone={tone}
      className={cn(s.root, className)}
      action={
        hasMore ? (
          <Button asChild size="small" tone="ghost" className={s.allAction}>
            <Link to={allIncidentsHref}>{incidentTexts.focus.viewAllAction(incidents.length)}</Link>
          </Button>
        ) : undefined
      }
    >
      {incident ? (
        <IncidentCard incident={incident} />
      ) : (
        <EmptyState
          tone={tone === 'inverse' ? 'inverse' : 'success'}
          illustration={<AllOkIllustration />}
          title={incidentTexts.focus.houseOkEmptyTitle}
          description={incidentTexts.focus.houseOkEmptyDescription}
        />
      )}
    </Section>
  )
}
