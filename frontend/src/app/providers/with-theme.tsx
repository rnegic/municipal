import { ThemeProvider } from '@/features/theme-switch'

import { createProvider, type ProviderWrapperProps } from './create-provider'

const ThemeWrapper = ({ children }: ProviderWrapperProps) => (
  <ThemeProvider>{children}</ThemeProvider>
)

export const WithTheme = createProvider(ThemeWrapper)
