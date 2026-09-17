import { MaxUI } from '@maxhub/max-ui'

import { useTheme } from '@/features/theme-switch'

import { THEME_ROOT_CLASS } from '../styles/themeRoot'

import { createProvider, type ProviderWrapperProps } from './create-provider'

const MaxUIProvider = ({ children }: ProviderWrapperProps) => {
  const { colorScheme } = useTheme()

  return (
    <MaxUI className={THEME_ROOT_CLASS} colorScheme={colorScheme}>
      {children}
    </MaxUI>
  )
}

export const WithMaxUI = createProvider(MaxUIProvider)
