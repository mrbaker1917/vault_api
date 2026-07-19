import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import { decryptPayload, encryptPayload } from '../crypto/blob'
import { deriveVaultKey } from '../crypto/kdf'
import {
  createAndStoreSalt,
  hasStoredSalt,
  loadSalt,
} from '../crypto/salt'
import { createVerifier, verifyMasterPassword } from '../crypto/verifier'
import type { VaultItemPayload } from '../crypto/types'
import { bytesToBase64, base64ToBytes } from '../crypto/encoding'
import { useAuth } from './AuthContext'

type VaultContextValue = {
  unlocked: boolean
  needsSetup: boolean
  unlock: (masterPassword: string) => Promise<void>
  setupMasterPassword: (masterPassword: string) => Promise<void>
  lock: () => void
  encryptItemPayload: (payload: VaultItemPayload) => Promise<string>
  decryptItemPayload: (encryptedData: string) => Promise<VaultItemPayload>
}

const VaultContext = createContext<VaultContextValue | null>(null)

export function VaultProvider({ children }: { children: ReactNode }) {
  const { user } = useAuth()
  const [vaultKey, setVaultKey] = useState<CryptoKey | null>(null)
  const [needsSetup, setNeedsSetup] = useState(false)

  useEffect(() => {
    setVaultKey(null)
    if (!user) {
      setNeedsSetup(false)
      return
    }
    setNeedsSetup(!hasStoredSalt(user.id))
  }, [user])

  const lock = useCallback(() => {
    setVaultKey(null)
  }, [])

  const unlock = useCallback(
    async (masterPassword: string) => {
      if (!user) {
        throw new Error('not authenticated')
      }
      const valid = await verifyMasterPassword(user.id, masterPassword)
      if (!valid) {
        throw new Error('incorrect master password')
      }
      const salt = loadSalt(user.id)
      const key = await deriveVaultKey(masterPassword, salt)
      setVaultKey(key)
    },
    [user],
  )

  const setupMasterPassword = useCallback(
    async (masterPassword: string) => {
      if (!user) {
        throw new Error('not authenticated')
      }
      createAndStoreSalt(user.id)
      await createVerifier(user.id, masterPassword)
      const salt = loadSalt(user.id)
      const key = await deriveVaultKey(masterPassword, salt)
      setVaultKey(key)
      setNeedsSetup(false)
    },
    [user],
  )

  const requireKey = useCallback((): CryptoKey => {
    if (!vaultKey) {
      throw new Error('vault is locked')
    }
    return vaultKey
  }, [vaultKey])

  const encryptItemPayload = useCallback(
    async (payload: VaultItemPayload): Promise<string> => {
      const key = requireKey()
      const blob = await encryptPayload(key, payload)
      return bytesToBase64(blob)
    },
    [requireKey],
  )

  const decryptItemPayload = useCallback(
    async (encryptedData: string): Promise<VaultItemPayload> => {
      const key = requireKey()
      return decryptPayload(key, base64ToBytes(encryptedData))
    },
    [requireKey],
  )

  const value = useMemo(
    () => ({
      unlocked: vaultKey != null,
      needsSetup,
      unlock,
      setupMasterPassword,
      lock,
      encryptItemPayload,
      decryptItemPayload,
    }),
    [
      vaultKey,
      needsSetup,
      unlock,
      setupMasterPassword,
      lock,
      encryptItemPayload,
      decryptItemPayload,
    ],
  )

  return <VaultContext.Provider value={value}>{children}</VaultContext.Provider>
}

export function useVault(): VaultContextValue {
  const ctx = useContext(VaultContext)
  if (!ctx) {
    throw new Error('useVault must be used within VaultProvider')
  }
  return ctx
}
