const GUARD_COOLDOWN_MS = 800
const GUARD_SAFETY_MS = 60_000

let armed = false
let activeUntil = 0
let safetyTimer: ReturnType<typeof setTimeout> | undefined
let listenerAttached = false

const handleWindowFocus = () => {
  if (armed) {
    activeUntil = Date.now() + GUARD_COOLDOWN_MS
  }
}

const attachListener = () => {
  if (listenerAttached || typeof window === 'undefined') {
    return
  }

  listenerAttached = true
  window.addEventListener('focus', handleWindowFocus)
}

const disarm = () => {
  armed = false
  activeUntil = 0

  if (safetyTimer) {
    clearTimeout(safetyTimer)
    safetyTimer = undefined
  }
}

export const notifyFilePickerOpened = () => {
  if (typeof window === 'undefined') {
    return
  }

  attachListener()
  armed = true
  activeUntil = 0

  if (safetyTimer) {
    clearTimeout(safetyTimer)
  }

  safetyTimer = setTimeout(disarm, GUARD_SAFETY_MS)
}

export const notifyFilePickerClosed = () => {
  disarm()
}

export const isFilePickerGuardActive = () => {
  if (!armed) {
    return false
  }

  if (typeof document !== 'undefined' && !document.hasFocus()) {
    return true
  }

  return Date.now() < activeUntil
}
