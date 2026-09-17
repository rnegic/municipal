import type { ReactNode } from 'react'

import { Navigate } from 'react-router-dom'

import { getBoundHouse } from '@/entities/house'
import { ROUTES } from '@/shared/config/routes'

export interface RequireAddressProps {
  children: ReactNode
}

export const RequireAddress = ({ children }: RequireAddressProps) =>
  getBoundHouse() ? children : <Navigate to={ROUTES.onboarding} replace />
