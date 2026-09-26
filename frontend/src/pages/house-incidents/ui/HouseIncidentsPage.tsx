import { Link, Navigate } from 'react-router-dom'

import { incidentTexts, useActiveIncidentsQuery } from '@/entities/incident'
import { useCurrentHouseQuery } from '@/entities/user'
import { IconChevronLeft } from '@/shared/assets/icons'
import { AllOkIllustration } from '@/shared/assets/illustrations'
import { ROUTES } from '@/shared/config/routes'
import { commonTexts } from '@/shared/config/texts'
import { ApiErrorState } from '@/shared/ui/api-error-state'
import { Button } from '@/shared/ui/button'
import { EmptyState } from '@/shared/ui/empty-state'
import { LoadingState } from '@/shared/ui/loading-state'
import { PageLayout } from '@/shared/ui/page-layout'
import { Section } from '@/shared/ui/section'
import { IncidentCard } from '@/widgets/incident-focus'
import s from './HouseIncidentsPage.module.scss'

export const HouseIncidentsPage = () => {
  const houseQuery = useCurrentHouseQuery()
  const incidentsQuery = useActiveIncidentsQuery(houseQuery.data?.id)

  const header = (
    <Button className={s.back} asChild size="small" tone="ghost" iconBefore={<IconChevronLeft size={16} />}>
      <Link to={ROUTES.feed}>{commonTexts.actions.goHome}</Link>
    </Button>
  )

  if (houseQuery.isPending) {
    return (
      <PageLayout header={header}>
        <LoadingState />
      </PageLayout>
    )
  }

  if (houseQuery.isError) {
    return (
      <PageLayout header={header}>
        <ApiErrorState error={houseQuery.error} onRetry={() => houseQuery.refetch()} />
      </PageLayout>
    )
  }

  if (!houseQuery.data) {
    return <Navigate to={ROUTES.onboarding} replace />
  }

  return (
    <PageLayout header={header}>
      {incidentsQuery.isPending ? (
        <LoadingState />
      ) : incidentsQuery.isError ? (
        <ApiErrorState error={incidentsQuery.error} onRetry={() => incidentsQuery.refetch()} />
      ) : incidentsQuery.data.items.length === 0 ? (
        <EmptyState
          tone="success"
          illustration={<AllOkIllustration />}
          title={incidentTexts.houseIncidents.emptyTitle}
          description={incidentTexts.houseIncidents.emptyDescription}
        />
      ) : (
        <Section title={incidentTexts.houseIncidents.title}>
          <ul className={s.list}>
            {incidentsQuery.data.items.map((incident) => (
              <li key={incident.id}>
                <IncidentCard incident={incident} />
              </li>
            ))}
          </ul>
        </Section>
      )}
    </PageLayout>
  )
}
