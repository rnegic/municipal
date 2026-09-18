import { createContext, useContext } from 'react'

import type { ColorSchemeType } from '@maxhub/max-ui'

export interface ThemeContextValue {
  colorScheme: ColorSchemeType
  toggleColorScheme: () => void
}

export const STORAGE_KEY = 'municipal.color-scheme'

export const ThemeContext = createContext<ThemeContextValue | null>(null)

export const useTheme = (): ThemeContextValue => {
  const value = useContext(ThemeContext)

  if (!value) {
    throw new Error('useTheme must be used within ThemeProvider')
  }

  return value
}
