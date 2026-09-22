import { Link, useParams } from 'react-router-dom'

import { useIncidentQuery } from '@/entities/incident'
import { IconChevronLeft } from '@/shared/assets/icons'
import { ROUTES } from '@/shared/config/routes'
import { commonTexts } from '@/shared/config/texts'
import { ApiErrorState } from '@/shared/ui/api-error-state'
import { Button } from '@/shared/ui/button'
import { LoadingState } from '@/shared/ui/loading-state'
import { PageLayout } from '@/shared/ui/page-layout'
import { IncidentFocus } from '@/widgets/incident-focus'
import s from './IncidentPage.module.scss'

export const IncidentPage = () => {
  const { id } = useParams<{ id: string }>()
  const incidentQuery = useIncidentQuery(id)

  return (
    <PageLayout
      header={
        <Button
          className={s.back}
          asChild
          size="medium"
          tone="ghost"
          iconBefore={<IconChevronLeft size={18} />}
        >
          <Link to={ROUTES.feed}>{commonTexts.actions.goHome}</Link>
        </Button>
      }
    >
      {incidentQuery.isPending ? (
        <LoadingState />
      ) : incidentQuery.isError ? (
        <ApiErrorState error={incidentQuery.error} onRetry={() => incidentQuery.refetch()} />
      ) : (
        <IncidentFocus incident={incidentQuery.data} />
      )}
    </PageLayout>
  )
}
