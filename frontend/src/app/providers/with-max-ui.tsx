import { MaxUI } from '@maxhub/max-ui'

import { THEME_ROOT_CLASS } from '../styles/themeRoot'

import { createProvider, type ProviderWrapperProps } from './create-provider'

const MaxUIProvider = ({ children }: ProviderWrapperProps) => (
  <MaxUI className={THEME_ROOT_CLASS}>{children}</MaxUI>
)

export const WithMaxUI = createProvider(MaxUIProvider)
