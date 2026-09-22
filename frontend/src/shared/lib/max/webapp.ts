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

const START_PARAM_PATTERN = /^[A-Za-z0-9_-]{1,512}$/

/** Диплинк мини-приложения: https://max.ru/{botName}?startapp={payload} */
export const buildMiniAppLink = (startParam: string): string | null =>
  env.maxBotName && START_PARAM_PATTERN.test(startParam)
    ? `https://max.ru/${env.maxBotName}?startapp=${startParam}`
    : null

/** Диплинк экрана «Отправить в MAX»: https://max.ru/:share?text={text} */
export const buildShareDeepLink = (text: string): string =>
  `https://max.ru/:share?text=${encodeURIComponent(text)}`
