import type { ReactNode } from 'react'

import { Navigate } from 'react-router-dom'

import { useCurrentHouseQuery } from '@/entities/user'
import { ROUTES } from '@/shared/config/routes'
import { ApiErrorState } from '@/shared/ui/api-error-state'
import { LoadingState } from '@/shared/ui/loading-state'
import { PageLayout } from '@/shared/ui/page-layout'

export interface RequireAddressProps {
  children: ReactNode
}

export const RequireAddress = ({ children }: RequireAddressProps) => {
  const houseQuery = useCurrentHouseQuery()

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

  return houseQuery.data ? children : <Navigate to={ROUTES.onboarding} replace />
}
