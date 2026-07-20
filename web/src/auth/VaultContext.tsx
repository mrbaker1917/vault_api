import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
import * as vaultApi from '../api/vault'
import { decryptPayload, encryptPayload } from '../crypto/blob'
import { migrateIfNeeded, MigrationError } from '../crypto/migrate'
import { deriveVaultKey } from '../crypto/kdf'
import { getPrimarySalt } from '../crypto/salt'
import { createVerifier, hasStoredVerifier, resolveSaltForPassword } from '../crypto/verifier'
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
  const userId = user?.id
  const [vaultKey, setVaultKey] = useState<CryptoKey | null>(null)
  const [needsSetup, setNeedsSetup] = useState(false)

  // Only clear the derived key when the signed-in account changes (logout / different user).
  useEffect(() => {
    setVaultKey(null)
  }, [userId])

  useEffect(() => {
    async function detectSetupRequired() {
      if (!userId) {
        setNeedsSetup(false)
        return
      }

      if (hasStoredVerifier(userId)) {
        setNeedsSetup(false)
        return
      }

      try {
        const { total } = await vaultApi.listVaultItems({ limit: 1 })
        setNeedsSetup(total === 0)
      } catch {
        // If the API is unreachable, prefer unlock over setup to avoid clobbering existing vaults.
        setNeedsSetup(false)
      }
    }

    void detectSetupRequired()
  }, [userId])

  const lock = useCallback(() => {
    setVaultKey(null)
  }, [])

  const unlock = useCallback(
    async (masterPassword: string) => {
      if (!user) {
        throw new Error('not authenticated')
      }

      let encryptedSample: string | undefined
      if (!hasStoredVerifier(user.id)) {
        const { items } = await vaultApi.listVaultItems({ limit: 1 })
        encryptedSample = items[0]?.EncryptedData
      }

      const salt = await resolveSaltForPassword(user.id, masterPassword, encryptedSample)
      if (!salt) {
        throw new Error('incorrect master password')
      }

      try {
        await migrateIfNeeded(user.id, masterPassword, salt)
      } catch (err) {
        if (err instanceof MigrationError) {
          throw err
        }
        throw new Error('vault migration failed')
      }

      const primarySalt = await getPrimarySalt(user.id)
      const key = await deriveVaultKey(masterPassword, primarySalt)
      setVaultKey(key)
    },
    [user],
  )

  const setupMasterPassword = useCallback(
    async (masterPassword: string) => {
      if (!user) {
        throw new Error('not authenticated')
      }
      await createVerifier(user.id, masterPassword)
      const salt = await getPrimarySalt(user.id)
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
