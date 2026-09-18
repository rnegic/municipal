import { getBoundHouse, PLACEHOLDER_HOUSE } from '@/entities/house'
import { incidentTexts, PLACEHOLDER_ACTIVE_INCIDENT, PLACEHOLDER_REQUESTS } from '@/entities/incident'
import { IncidentReportFab } from '@/features/incident-report'
import { PageLayout } from '@/shared/ui/page-layout'
import { AppHeader } from '@/widgets/app-header'
import { IncidentFocus } from '@/widgets/incident-focus'
import { MyRequests } from '@/widgets/my-requests'
import s from './FeedPage.module.scss'

export const FeedPage = () => {
  const incident = PLACEHOLDER_ACTIVE_INCIDENT
  const house = getBoundHouse() ?? PLACEHOLDER_HOUSE

  return (
    <PageLayout
      hero={
        <div className={s.hero}>
          <div className={s.heroInner}>
            <AppHeader
              house={house}
              status={
                incident
                  ? { tone: 'danger', label: incidentTexts.focus.headerStatusAlert }
                  : { tone: 'success', label: incidentTexts.focus.houseOkTitle }
              }
            />
            <IncidentFocus incident={incident} tone="inverse" />
          </div>
        </div>
      }
      floatingAction={<IncidentReportFab />}
    >
      <MyRequests requests={PLACEHOLDER_REQUESTS} />
    </PageLayout>
  )
}
