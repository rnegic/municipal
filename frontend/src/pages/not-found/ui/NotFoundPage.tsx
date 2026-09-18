import { Link } from 'react-router-dom'

import { IconAlertTriangle } from '@/shared/assets/icons'
import { ROUTES } from '@/shared/config/routes'
import { commonTexts } from '@/shared/config/texts'
import { Button } from '@/shared/ui/button'
import { EmptyState } from '@/shared/ui/empty-state'
import { PageLayout } from '@/shared/ui/page-layout'

export const NotFoundPage = () => (
  <PageLayout>
    <EmptyState
      icon={<IconAlertTriangle size={24} />}
      title={commonTexts.notFound.title}
      description={commonTexts.notFound.description}
      action={
        <Button tone="secondary" asChild>
          <Link to={ROUTES.feed}>{commonTexts.actions.goHome}</Link>
        </Button>
      }
    />
  </PageLayout>
)
