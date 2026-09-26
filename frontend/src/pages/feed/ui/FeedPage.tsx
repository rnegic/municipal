import { Navigate } from 'react-router-dom'

import { incidentTexts, useActiveIncidentsQuery, useMyRequestsQuery } from '@/entities/incident'
import { useCurrentHouseQuery } from '@/entities/user'
import { IncidentReportFab } from '@/features/incident-report'
import { ROUTES } from '@/shared/config/routes'
import { ApiErrorState } from '@/shared/ui/api-error-state'
import { LoadingState } from '@/shared/ui/loading-state'
import { PageLayout } from '@/shared/ui/page-layout'
import { Skeleton } from '@/shared/ui/skeleton'
import { AppHeader } from '@/widgets/app-header'
import { EventBulletin } from '@/widgets/event-bulletin'
import { IncidentFocus } from '@/widgets/incident-focus'
import { MyRequests } from '@/widgets/my-requests'
import s from './FeedPage.module.scss'

export const FeedPage = () => {
  const houseQuery = useCurrentHouseQuery()
  const incidentsQuery = useActiveIncidentsQuery(houseQuery.data?.id)
  const requestsQuery = useMyRequestsQuery(houseQuery.data?.id)

  if (houseQuery.isPending) {
    return (
      <PageLayout>
        <LoadingState />
      </PageLayout>
    )
  }

  if (houseQuery.isError) {
    return (
      <PageLayout>
        <ApiErrorState error={houseQuery.error} onRetry={() => houseQuery.refetch()} />
      </PageLayout>
    )
  }

  if (!houseQuery.data) {
    return <Navigate to={ROUTES.onboarding} replace />
  }

  if (incidentsQuery.isPending) {
    return (
      <PageLayout>
        <LoadingState />
      </PageLayout>
    )
  }

  if (incidentsQuery.isError) {
    return (
      <PageLayout>
        <ApiErrorState error={incidentsQuery.error} onRetry={() => incidentsQuery.refetch()} />
      </PageLayout>
    )
  }

  const incidents = incidentsQuery.data.items
  const incident = incidents[0] ?? null

  return (
    <PageLayout
      hero={
        <div className={s.hero}>
          <div className={s.heroInner}>
            <AppHeader
              house={houseQuery.data}
              status={
                incident
                  ? { tone: 'danger', label: incidentTexts.focus.headerStatusAlert }
                  : { tone: 'success', label: incidentTexts.focus.houseOkTitle }
              }
            />
            <EventBulletin houseId={houseQuery.data.id} tone="inverse" />
            <IncidentFocus
              incidents={incidents}
              tone="inverse"
              allIncidentsHref={ROUTES.houseIncidents}
            />
          </div>
        </div>
      }
      floatingAction={<IncidentReportFab />}
    >
      {requestsQuery.isPending ? (
        <div className={s.requestsSkeleton}>
          <Skeleton height={88} radius="card" />
          <Skeleton height={88} radius="card" />
          <Skeleton height={88} radius="card" />
        </div>
      ) : requestsQuery.isError ? (
        <ApiErrorState error={requestsQuery.error} onRetry={() => requestsQuery.refetch()} />
      ) : (
        <MyRequests requests={requestsQuery.data.items} />
      )}
    </PageLayout>
  )
}