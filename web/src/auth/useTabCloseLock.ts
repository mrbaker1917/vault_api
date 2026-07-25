import { useEffect, useRef, useState } from 'react'
import { getIdleTimeoutPrefs, subscribeIdleTimeoutPrefs } from './idle-timeout-prefs'

/** Lock the vault when the tab unloads or is restored from the back-forward cache. */
export function useTabCloseLock(enabled: boolean, onLock: () => void): void {
  const onLockRef = useRef(onLock)
  onLockRef.current = onLock

  useEffect(() => {
    if (!enabled) {
      return
    }

    function handlePageHide() {
      onLockRef.current()
    }

    function handlePageShow(event: PageTransitionEvent) {
      if (event.persisted) {
        onLockRef.current()
      }
    }

    window.addEventListener('pagehide', handlePageHide)
    window.addEventListener('pageshow', handlePageShow)
    return () => {
      window.removeEventListener('pagehide', handlePageHide)
      window.removeEventListener('pageshow', handlePageShow)
    }
  }, [enabled])
}

export function useLockOnTabClosePref(): boolean {
  const [lockOnTabClose, setLockOnTabClose] = useState(
    () => getIdleTimeoutPrefs().lockOnTabClose,
  )

  useEffect(() => {
    return subscribeIdleTimeoutPrefs(() => {
      setLockOnTabClose(getIdleTimeoutPrefs().lockOnTabClose)
    })
  }, [])

  return lockOnTabClose
}
