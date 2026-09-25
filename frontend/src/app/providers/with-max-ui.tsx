import { MaxUI } from '@maxhub/max-ui'

import { useTheme } from '@/features/theme-switch'
import { cn } from '@/shared/lib/cn'

import { THEME_DARK_CLASS, THEME_LIGHT_CLASS, THEME_ROOT_CLASS } from '../styles/themeRoot'

import { createProvider, type ProviderWrapperProps } from './create-provider'

const MaxUIProvider = ({ children }: ProviderWrapperProps) => {
  const { colorScheme } = useTheme()

  return (
    <MaxUI
      className={cn(
        THEME_ROOT_CLASS,
        colorScheme === 'dark' ? THEME_DARK_CLASS : THEME_LIGHT_CLASS,
      )}
      colorScheme={colorScheme}
    >
      {children}
    </MaxUI>
  )
}

export const WithMaxUI = createProvider(MaxUIProvider)
