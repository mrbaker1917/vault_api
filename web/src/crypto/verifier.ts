import { bytesToBase64, base64ToBytes } from './encoding'
import { encryptPayload, decryptPayload } from './blob'
import { deriveVaultKey } from './kdf'
import { loadSalt } from './salt'
import type { NotePayload } from './types'

const VERIFIER_PREFIX = 'vault_verifier_'
const VERIFIER_MARKER = '__vault_verifier__'

export function storeVerifier(userId: string, encrypted: string): void {
  localStorage.setItem(`${VERIFIER_PREFIX}${userId}`, encrypted)
}

export function loadVerifier(userId: string): string | null {
  return localStorage.getItem(`${VERIFIER_PREFIX}${userId}`)
}

export async function createVerifier(userId: string, masterPassword: string): Promise<void> {
  const salt = loadSalt(userId)
  const key = await deriveVaultKey(masterPassword, salt)
  const payload: NotePayload = { notes: VERIFIER_MARKER }
  const blob = await encryptPayload(key, payload)
  storeVerifier(userId, bytesToBase64(blob))
}

export async function verifyMasterPassword(
  userId: string,
  masterPassword: string,
): Promise<boolean> {
  const stored = loadVerifier(userId)
  if (!stored) {
    return true
  }

  try {
    const salt = loadSalt(userId)
    const key = await deriveVaultKey(masterPassword, salt)
    const payload = await decryptPayload(key, base64ToBytes(stored)) as NotePayload
    return payload.notes === VERIFIER_MARKER
  } catch {
    return false
  }
}

export function clearVerifier(userId: string): void {
  localStorage.removeItem(`${VERIFIER_PREFIX}${userId}`)
}
