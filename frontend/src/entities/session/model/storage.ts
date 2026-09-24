import { useSyncExternalStore } from 'react'

import {
  clearAccessToken,
  setAccessToken,
  subscribeAccessToken,
} from '@/shared/lib/auth'
import { ukSessionSchema, type UkSession } from './schema'

const STORAGE_KEY = 'app.uk-session'

let cache: UkSession | null | undefined

const read = (): UkSession | null => {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)
    const parsed = raw ? ukSessionSchema.safeParse(JSON.parse(raw)) : null

    if (!parsed?.success) {
      return null
    }

    return Date.parse(parsed.data.expiresAt) > Date.now() ? parsed.data : null
  } catch {
    return null
  }
}

export const getUkSession = (): UkSession | null => {
  if (cache === undefined) {
    cache = read()
  }

  return cache
}

export const saveUkSession = (session: UkSession): void => {
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
  cache = session
  setAccessToken({ token: session.token, expiresAt: session.expiresAt })
}

export const clearUkSession = (): void => {
  window.localStorage.removeItem(STORAGE_KEY)
  cache = null
  clearAccessToken()
}

const subscribe = (listener: () => void): (() => void) =>
  subscribeAccessToken(() => {
    cache = undefined
    listener()
  })

export const useUkSession = (): UkSession | null =>
  useSyncExternalStore(subscribe, getUkSession, () => null)
