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
import { createAndStoreSalt } from '../crypto/salt'
import { createVerifier } from '../crypto/verifier'
import { clearTokens, getAccessToken } from './tokens'

type User = {
  id: string
}

type AuthContextValue = {
  user: User | null
  loading: boolean
  login: (email: string, password: string, totpCode?: string) => Promise<void>
  signup: (email: string, password: string, masterPassword?: string) => Promise<void>
  logout: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  const loadUser = useCallback(async () => {
    if (!getAccessToken()) {
      setUser(null)
      setLoading(false)
      return
    }

    try {
      const me = await authApi.fetchMe()
      setUser({ id: me.id })
    } catch {
      clearTokens()
      setUser(null)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadUser()
  }, [loadUser])

  const login = useCallback(async (email: string, password: string, totpCode?: string) => {
    await authApi.login(email, password, totpCode)
    const me = await authApi.fetchMe()
    setUser({ id: me.id })
  }, [])

  const signup = useCallback(async (email: string, password: string, masterPassword?: string) => {
    await authApi.signup(email, password)
    await authApi.login(email, password)
    const me = await authApi.fetchMe()
    if (masterPassword) {
      createAndStoreSalt(me.id)
      await createVerifier(me.id, masterPassword)
    }
    setUser({ id: me.id })
  }, [])

  const logout = useCallback(async () => {
    await authApi.logout()
    setUser(null)
  }, [])

  const value = useMemo(
    () => ({ user, loading, login, signup, logout }),
    [user, loading, login, signup, logout],
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
