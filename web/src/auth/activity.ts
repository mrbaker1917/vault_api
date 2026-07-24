const activityListeners = new Set<() => void>()

/** Notify idle timers that the user performed an action. */
export function markUserActivity(): void {
  for (const listener of activityListeners) {
    listener()
  }
}

export function subscribeUserActivity(listener: () => void): () => void {
  activityListeners.add(listener)
  return () => {
    activityListeners.delete(listener)
  }
}

const ACTIVITY_EVENTS = ['mousedown', 'keydown', 'touchstart', 'scroll'] as const
const THROTTLE_MS = 1000

export function attachDocumentActivityListeners(onActivity: () => void): () => void {
  let lastFired = 0

  function handleActivity() {
    const now = Date.now()
    if (now - lastFired < THROTTLE_MS) {
      return
    }
    lastFired = now
    onActivity()
  }

  for (const eventName of ACTIVITY_EVENTS) {
    document.addEventListener(eventName, handleActivity, { passive: true })
  }

  return () => {
    for (const eventName of ACTIVITY_EVENTS) {
      document.removeEventListener(eventName, handleActivity)
    }
  }
}
