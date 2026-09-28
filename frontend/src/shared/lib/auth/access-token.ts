const STORAGE_KEY = 'app.access-token'

export interface AccessToken {
  token: string
  expiresAt: string
}

type Listener = () => void

const listeners = new Set<Listener>()

let cache: AccessToken | null | undefined

const read = (): AccessToken | null => {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY)

    if (!raw) {
      return null
    }

    const parsed: unknown = JSON.parse(raw)

    if (
      typeof parsed !== 'object' ||
      parsed === null ||
      typeof (parsed as AccessToken).token !== 'string' ||
      typeof (parsed as AccessToken).expiresAt !== 'string'
    ) {
      return null
    }

    return parsed as AccessToken
  } catch {
    return null
  }
}

const notify = (): void => {
  cache = undefined
  listeners.forEach((listener) => listener())
}

export const getAccessToken = (): AccessToken | null => {
  cache ??= read()

  return cache
}

export const getValidAccessToken = (): string | null => {
  const stored = getAccessToken()

  if (!stored) {
    return null
  }

  if (Date.parse(stored.expiresAt) <= Date.now()) {
    clearAccessToken()
    return null
  }

  return stored.token
}

export const setAccessToken = (value: AccessToken): void => {
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(value))
  notify()
}

export const clearAccessToken = (): void => {
  window.localStorage.removeItem(STORAGE_KEY)
  notify()
}

export const subscribeAccessToken = (listener: Listener): (() => void) => {
  listeners.add(listener)

  const onStorage = (event: StorageEvent) => {
    if (event.key === STORAGE_KEY || event.key === null) {
      notify()
    }
  }

  window.addEventListener('storage', onStorage)

  return () => {
    listeners.delete(listener)
    window.removeEventListener('storage', onStorage)
  }
}
