import type { ReactNode } from 'react'

export type UserRole = 'resident' | 'dispatcher'

export interface RequireRoleProps {

  role: UserRole
  children: ReactNode
}

export const RequireRole = ({ children }: RequireRoleProps) => <>{children}</>
