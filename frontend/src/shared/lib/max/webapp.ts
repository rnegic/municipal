import { env } from '@/shared/config/env'
import type {
  MaxWebAppApi,
  MaxWebAppInitData,
  MaxWebAppShareContentParams,
  MaxWebAppUser,
} from './types'

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

const LAUNCH_DATA_PARAM = 'WebAppData'

export const normalizeLaunchHash = (): void => {
  if (typeof window === 'undefined' || !window.location.hash.includes(`${LAUNCH_DATA_PARAM}=`)) {
    return
  }

  window.history.replaceState(null, '', `${window.location.href.split('#')[0]}#/`)
}

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

export const openMaxLink = (url: string): boolean => {
  const webApp = getWebApp()

  if (!webApp?.openMaxLink) {
    return false
  }

  webApp.openMaxLink(url)
  return true
}

export const shareMaxContent = (params: MaxWebAppShareContentParams): boolean => {
  const webApp = getWebApp()

  if (!webApp?.shareMaxContent) {
    return false
  }

  void Promise.resolve(webApp.shareMaxContent(params)).catch(() => undefined)
  return true
}

export const openLink = (url: string): void => {
  const webApp = getWebApp()

  if (webApp?.openLink) {
    webApp.openLink(url)
    return
  }

  window.open(url, '_blank', 'noopener,noreferrer')
}

const START_PARAM_PATTERN = /^[A-Za-z0-9_-]{1,512}$/

export const buildMiniAppLink = (startParam: string): string | null =>
  env.maxBotName && START_PARAM_PATTERN.test(startParam)
    ? `https://max.ru/${env.maxBotName}?startapp=${startParam}`
    : null

export const buildShareDeepLink = (text: string): string =>
  `https://max.ru/:share?text=${encodeURIComponent(text)}`
