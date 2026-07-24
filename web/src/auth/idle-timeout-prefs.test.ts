import { describe, expect, it, beforeEach, vi } from 'vitest'
import {
  DEFAULT_IDLE_TIMEOUT_PREFS,
  getIdleTimeoutPrefs,
  minutesToMs,
  setIdleTimeoutPrefs,
} from './idle-timeout-prefs'

function createStorage(): Storage {
  const store = new Map<string, string>()
  return {
    get length() {
      return store.size
    },
    clear: () => store.clear(),
    getItem: (key) => store.get(key) ?? null,
    key: (index) => [...store.keys()][index] ?? null,
    removeItem: (key) => {
      store.delete(key)
    },
    setItem: (key, value) => {
      store.set(key, value)
    },
  }
}

describe('idle-timeout-prefs', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', createStorage())
    vi.stubGlobal('window', {
      dispatchEvent: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    })
  })

  it('returns defaults when nothing is stored', () => {
    expect(getIdleTimeoutPrefs()).toEqual(DEFAULT_IDLE_TIMEOUT_PREFS)
  })

  it('persists updated preferences', () => {
    setIdleTimeoutPrefs({ vaultLockMinutes: 30, logoutMinutes: 60 })
    expect(getIdleTimeoutPrefs()).toEqual({ vaultLockMinutes: 30, logoutMinutes: 60 })
  })

  it('converts minutes to milliseconds', () => {
    expect(minutesToMs(15)).toBe(900_000)
  })
})
