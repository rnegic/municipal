import { useCallback, useEffect, useMemo, useState } from 'react'
import type { ReactNode } from 'react'

import { useSystemColorScheme, type ColorSchemeType } from '@maxhub/max-ui'

import { STORAGE_KEY, ThemeContext } from './theme-context'

export interface ThemeProviderProps {
  children: ReactNode
}

const readStoredColorScheme = (): ColorSchemeType | null => {
  const stored = window.localStorage.getItem(STORAGE_KEY)

  return stored === 'light' || stored === 'dark' ? stored : null
}

export const ThemeProvider = ({ children }: ThemeProviderProps) => {
  const systemColorScheme = useSystemColorScheme({ listenChanges: true })
  const [preferredColorScheme, setPreferredColorScheme] = useState(readStoredColorScheme)

  const colorScheme = preferredColorScheme ?? systemColorScheme

  const toggleColorScheme = useCallback(() => {
    setPreferredColorScheme(colorScheme === 'dark' ? 'light' : 'dark')
  }, [colorScheme])

  useEffect(() => {
    document.documentElement.style.colorScheme = colorScheme
  }, [colorScheme])

  useEffect(() => {
    if (preferredColorScheme) {
      window.localStorage.setItem(STORAGE_KEY, preferredColorScheme)
    }
  }, [preferredColorScheme])

  const value = useMemo(
    () => ({ colorScheme, toggleColorScheme }),
    [colorScheme, toggleColorScheme],
  )

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}
