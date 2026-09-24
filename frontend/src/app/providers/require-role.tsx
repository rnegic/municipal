import type { ReactNode } from 'react'

import { Navigate, useLocation } from 'react-router-dom'

import { useUkSession } from '@/entities/session'
import { ROUTES } from '@/shared/config/routes'

export type UserRole = 'resident' | 'uk_dispatcher'

export interface RequireRoleProps {
  role: UserRole
  children: ReactNode
}

export interface UkAuthIntentState {
  ukAuth?: boolean
  from?: string
}

export const RequireRole = ({ role, children }: RequireRoleProps) => {
  const session = useUkSession()
  const location = useLocation()

  if (role === 'uk_dispatcher' && session?.user.role !== 'uk_dispatcher') {
    const state: UkAuthIntentState = { ukAuth: true, from: location.pathname }

    return <Navigate to={ROUTES.onboarding} state={state} replace />
  }

  return <>{children}</>
}
