import { Typography } from '@maxhub/max-ui'
import { useNavigate } from 'react-router-dom'

import { IncidentReportForm } from '@/features/incident-report'
import { PageLayout } from '@/shared/ui/page-layout'
import { ROUTES } from '@/shared/config/routes'
import { incidentCreateTexts as texts } from '../config/texts'
import s from './IncidentCreatePage.module.scss'

export const IncidentCreatePage = () => {
  const navigate = useNavigate()

  return (
    <PageLayout>
      <main className={s.content}>
        <div className={s.intro}>
          <Typography.Title variant="large-strong">{texts.title}</Typography.Title>
          <Typography.Text variant="description" color="secondary">
            {texts.description}
          </Typography.Text>
        </div>
        <IncidentReportForm onSuccess={() => navigate(ROUTES.feed, { replace: true })} />
      </main>
    </PageLayout>
  )
}
