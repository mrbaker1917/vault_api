import { useCallback, useEffect, useRef, useState } from 'react'
import {
  attachDocumentActivityListeners,
  subscribeUserActivity,
} from './activity'
import {
  getIdleTimeoutPrefs,
  minutesToMs,
  subscribeIdleTimeoutPrefs,
  type IdleTimeoutPrefs,
} from './idle-timeout-prefs'

const CHECK_INTERVAL_MS = 5_000

type IdleTimeoutOptions = {
  enabled: boolean
  unlocked: boolean
  onVaultLock: () => void
  onLogout: () => void
}

export function useIdleTimeout({
  enabled,
  unlocked,
  onVaultLock,
  onLogout,
}: IdleTimeoutOptions): void {
  const [prefs, setPrefs] = useState<IdleTimeoutPrefs>(() => getIdleTimeoutPrefs())
  const lastActivityRef = useRef(Date.now())
  const vaultLockedForIdleRef = useRef(false)
  const onVaultLockRef = useRef(onVaultLock)
  const onLogoutRef = useRef(onLogout)

  onVaultLockRef.current = onVaultLock
  onLogoutRef.current = onLogout

  const bumpActivity = useCallback(() => {
    lastActivityRef.current = Date.now()
    vaultLockedForIdleRef.current = false
  }, [])

  useEffect(() => {
    return subscribeIdleTimeoutPrefs(() => {
      setPrefs(getIdleTimeoutPrefs())
      bumpActivity()
    })
  }, [bumpActivity])

  useEffect(() => {
    if (!enabled) {
      return
    }

    bumpActivity()
    const detachDocument = attachDocumentActivityListeners(bumpActivity)
    const detachActivity = subscribeUserActivity(bumpActivity)

    const intervalId = window.setInterval(() => {
      const idleMs = Date.now() - lastActivityRef.current
      const logoutMs = minutesToMs(prefs.logoutMinutes)
      const vaultLockMs = minutesToMs(prefs.vaultLockMinutes)

      if (logoutMs > 0 && idleMs >= logoutMs) {
        bumpActivity()
        onLogoutRef.current()
        return
      }

      if (
        vaultLockMs > 0 &&
        unlocked &&
        !vaultLockedForIdleRef.current &&
        idleMs >= vaultLockMs
      ) {
        vaultLockedForIdleRef.current = true
        onVaultLockRef.current()
      }
    }, CHECK_INTERVAL_MS)

    return () => {
      detachDocument()
      detachActivity()
      window.clearInterval(intervalId)
    }
  }, [enabled, unlocked, prefs, bumpActivity])
}

export { markUserActivity } from './activity'
