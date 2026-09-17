import { Suspense } from 'react'

import { Spinner, Typography } from '@maxhub/max-ui'
import { HashRouter } from 'react-router-dom'

import { commonTexts } from '@/shared/config/texts'

import { createProvider, type ProviderWrapperProps } from './create-provider'
import s from './with-router.module.scss'

const RouteFallback = () => (
  <div className={s.fallback}>
    <Spinner size={24} appearance="themed" />
    <Typography.Text variant="description" color="secondary">
      {commonTexts.states.loading}
    </Typography.Text>
  </div>
)

const RouterProvider = ({ children }: ProviderWrapperProps) => (
  <HashRouter>
    <Suspense fallback={<RouteFallback />}>{children}</Suspense>
  </HashRouter>
)

export const WithRouter = createProvider(RouterProvider)
