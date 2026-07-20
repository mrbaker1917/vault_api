import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { ApiError } from '../api/client'
import * as authApi from '../api/auth'
import type { MeResponse } from '../api/types'
import { createVerifier } from '../crypto/verifier'
import { clearMfaEnabledHint, resolveMfaEnabled, setMfaEnabledHint } from './mfa-hint'
import { clearTokens, getAccessToken } from './tokens'

type User = {
  id: string
  mfa_enabled: boolean
}

type AuthContextValue = {
  user: User | null
  loading: boolean
  login: (email: string, password: string, totpCode?: string) => Promise<void>
  signup: (email: string, password: string, masterPassword?: string) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

function toUser(me: MeResponse): User {
  return { id: me.id, mfa_enabled: resolveMfaEnabled(me.mfa_enabled) }
}

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  const refreshUser = useCallback(async () => {
    const me = await authApi.fetchMe()
    setUser(toUser(me))
  }, [])

  const loadUser = useCallback(async () => {
    if (!getAccessToken()) {
      setUser(null)
      setLoading(false)
      return
    }

    try {
      await refreshUser()
    } catch {
      clearTokens()
      setUser(null)
    } finally {
      setLoading(false)
    }
  }, [refreshUser])

  useEffect(() => {
    void loadUser()
  }, [loadUser])

  const login = useCallback(async (email: string, password: string, totpCode?: string) => {
    await authApi.login(email, password, totpCode)
    if (totpCode) {
      setMfaEnabledHint(true)
    }
    await refreshUser()
  }, [refreshUser])

  const signup = useCallback(async (email: string, password: string, masterPassword?: string) => {
    await authApi.signup(email, password)
    await authApi.login(email, password)
    const me = await authApi.fetchMe()
    if (masterPassword) {
      await createVerifier(me.id, masterPassword)
    }
    setUser(toUser(me))
  }, [])

  const logout = useCallback(async () => {
    await authApi.logout()
    clearMfaEnabledHint()
    setUser(null)
  }, [])

  const value = useMemo(
    () => ({ user, loading, login, signup, logout, refreshUser }),
    [user, loading, login, signup, logout, refreshUser],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) {
    throw new Error('useAuth must be used within AuthProvider')
  }
  return ctx
}

export { ApiError }
