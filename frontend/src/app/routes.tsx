import { lazy } from 'react'

import { Route, Routes } from 'react-router-dom'

import { ROUTES } from '@/shared/config/routes'

import { RequireAddress } from './providers/require-address'
import { RequireRole } from './providers/require-role'

const FeedPage = lazy(() => import('@/pages/feed').then((module) => ({ default: module.FeedPage })))
const IncidentCreatePage = lazy(() =>
  import('@/pages/incident-create').then((module) => ({ default: module.IncidentCreatePage })),
)
const OnboardingPage = lazy(() =>
  import('@/pages/onboarding').then((module) => ({ default: module.OnboardingPage })),
)
const NotFoundPage = lazy(() =>
  import('@/pages/not-found').then((module) => ({ default: module.NotFoundPage })),
)
const UkPage = lazy(() => import('@/pages/uk').then((module) => ({ default: module.UkPage })))

export const AppRoutes = () => (
  <Routes>
    <Route
      path={ROUTES.feed}
      element={
        <RequireAddress>
          <RequireRole role="resident">
            <FeedPage />
          </RequireRole>
        </RequireAddress>
      }
    />
    <Route path={ROUTES.incidentCreate} element={<IncidentCreatePage />} />
    <Route
      path={ROUTES.dispatcher}
      element={
        <RequireRole role="dispatcher">
          <UkPage />
        </RequireRole>
      }
    />
    <Route path={ROUTES.onboarding} element={<OnboardingPage />} />
    <Route path="*" element={<NotFoundPage />} />
  </Routes>
)
