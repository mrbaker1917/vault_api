export type IdleTimeoutPrefs = {
  /** Minutes of inactivity before the vault locks. 0 = never. */
  vaultLockMinutes: number
  /** Minutes of inactivity before signing out. 0 = never. */
  logoutMinutes: number
}

export const DEFAULT_IDLE_TIMEOUT_PREFS: IdleTimeoutPrefs = {
  vaultLockMinutes: 15,
  logoutMinutes: 30,
}

const STORAGE_KEY = 'vault_idle_timeout_prefs'
const PREFS_CHANGED_EVENT = 'vault-idle-timeout-prefs-changed'

export const VAULT_LOCK_OPTIONS = [
  { value: 5, label: '5 minutes' },
  { value: 15, label: '15 minutes' },
  { value: 30, label: '30 minutes' },
  { value: 60, label: '1 hour' },
  { value: 0, label: 'Never' },
] as const

export const LOGOUT_OPTIONS = [
  { value: 15, label: '15 minutes' },
  { value: 30, label: '30 minutes' },
  { value: 60, label: '1 hour' },
  { value: 120, label: '2 hours' },
  { value: 0, label: 'Never' },
] as const

function normalizePrefs(raw: Partial<IdleTimeoutPrefs>): IdleTimeoutPrefs {
  const vaultLockMinutes = Number.isFinite(raw.vaultLockMinutes)
    ? Math.max(0, Math.floor(raw.vaultLockMinutes!))
    : DEFAULT_IDLE_TIMEOUT_PREFS.vaultLockMinutes
  const logoutMinutes = Number.isFinite(raw.logoutMinutes)
    ? Math.max(0, Math.floor(raw.logoutMinutes!))
    : DEFAULT_IDLE_TIMEOUT_PREFS.logoutMinutes

  return { vaultLockMinutes, logoutMinutes }
}

export function getIdleTimeoutPrefs(): IdleTimeoutPrefs {
  try {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (!stored) {
      return { ...DEFAULT_IDLE_TIMEOUT_PREFS }
    }
    return normalizePrefs(JSON.parse(stored) as Partial<IdleTimeoutPrefs>)
  } catch {
    return { ...DEFAULT_IDLE_TIMEOUT_PREFS }
  }
}

export function setIdleTimeoutPrefs(prefs: IdleTimeoutPrefs): IdleTimeoutPrefs {
  const normalized = normalizePrefs(prefs)
  localStorage.setItem(STORAGE_KEY, JSON.stringify(normalized))
  window.dispatchEvent(new Event(PREFS_CHANGED_EVENT))
  return normalized
}

export function minutesToMs(minutes: number): number {
  return minutes * 60 * 1000
}

export function subscribeIdleTimeoutPrefs(onChange: () => void): () => void {
  const handleStorage = (event: StorageEvent) => {
    if (event.key === STORAGE_KEY) {
      onChange()
    }
  }
  const handleCustom = () => onChange()

  window.addEventListener('storage', handleStorage)
  window.addEventListener(PREFS_CHANGED_EVENT, handleCustom)
  return () => {
    window.removeEventListener('storage', handleStorage)
    window.removeEventListener(PREFS_CHANGED_EVENT, handleCustom)
  }
}
