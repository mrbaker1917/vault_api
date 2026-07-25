import { useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import { useIdleTimeout } from '../auth/useIdleTimeout'
import { useLockOnTabClosePref, useTabCloseLock } from '../auth/useTabCloseLock'
import { useVault } from '../auth/VaultContext'

export function IdleTimeoutManager() {
  const { user, logout } = useAuth()
  const { unlocked, lock } = useVault()
  const navigate = useNavigate()
  const lockOnTabClose = useLockOnTabClosePref()

  useTabCloseLock(user != null && lockOnTabClose, lock)

  useIdleTimeout({
    enabled: user != null,
    unlocked,
    onVaultLock: lock,
    onLogout: () => {
      void (async () => {
        lock()
        await logout()
        navigate('/login', { replace: true })
      })()
    },
  })

  return null
}
