import type { MaxWebAppApi, MaxWebAppInitData, MaxWebAppUser } from './types'

declare global {
  interface Window {
    WebApp?: MaxWebAppApi
  }
}

export const getWebApp = (): MaxWebAppApi | null => {
  if (typeof window === 'undefined') {
    return null
  }

  return window.WebApp ?? null
}

export const isInsideMax = (): boolean => getWebApp() !== null

export const getInitData = (): string => getWebApp()?.initData ?? ''

export const getInitDataUnsafe = (): MaxWebAppInitData => getWebApp()?.initDataUnsafe ?? {}

export const getUser = (): MaxWebAppUser | null => getInitDataUnsafe().user ?? null

export const getStartParam = (): string | null => getInitDataUnsafe().start_param ?? null

export const ready = (): void => {
  getWebApp()?.ready?.()
}

export const expand = (): void => {
  getWebApp()?.expand?.()
}
