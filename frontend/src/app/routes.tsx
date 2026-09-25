import { lazy, useEffect, useRef } from 'react'

import { Route, Routes, useNavigate } from 'react-router-dom'

import { ROUTES } from '@/shared/config/routes'
import { getStartParam } from '@/shared/lib/max'

import { RequireAddress } from './providers/require-address'
import { RequireRole } from './providers/require-role'

const FeedPage = lazy(() => import('@/pages/feed').then((module) => ({ default: module.FeedPage })))
const HouseIncidentsPage = lazy(() =>
  import('@/pages/house-incidents').then((module) => ({ default: module.HouseIncidentsPage })),
)
const IncidentPage = lazy(() =>
  import('@/pages/incident').then((module) => ({ default: module.IncidentPage })),
)
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

const INCIDENT_START_PARAM_PREFIX = 'inc_'

const useIncidentDeepLink = () => {
  const navigate = useNavigate()

  const handledRef = useRef(false)

  useEffect(() => {
    if (handledRef.current) return
    handledRef.current = true

    const startParam = getStartParam()

    if (startParam?.startsWith(INCIDENT_START_PARAM_PREFIX)) {
      navigate(ROUTES.incident(startParam), { replace: true })
    }
  }, [navigate])
}

export const AppRoutes = () => {
  useIncidentDeepLink()

  return (
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
        path={ROUTES.houseIncidents}
        element={
          <RequireAddress>
            <RequireRole role="resident">
              <HouseIncidentsPage />
            </RequireRole>
          </RequireAddress>
        }
      />
      <Route path={ROUTES.incident()} element={<IncidentPage />} />
      <Route
        path={ROUTES.dispatcher}
        element={
          <RequireRole role="uk_dispatcher">
            <UkPage />
          </RequireRole>
        }
      />
      <Route path={ROUTES.onboarding} element={<OnboardingPage />} />
      <Route path="*" element={<NotFoundPage />} />
    </Routes>
  )
}
