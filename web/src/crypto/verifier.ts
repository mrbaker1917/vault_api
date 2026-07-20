import { bytesToBase64, base64ToBytes } from './encoding'
import { encryptPayload, decryptPayload } from './blob'
import { deriveVaultKey } from './kdf'
import {
  bytesEqual,
  deriveDeterministicSalt,
  getSaltCandidates,
  getPrimarySalt,
  loadLegacySalt,
} from './salt'
import type { NotePayload } from './types'

const VERIFIER_PREFIX = 'vault_verifier_'
const VERIFIER_MARKER = '__vault_verifier__'

export function hasStoredVerifier(userId: string): boolean {
  return localStorage.getItem(`${VERIFIER_PREFIX}${userId}`) != null
}

export function storeVerifier(userId: string, encrypted: string): void {
  localStorage.setItem(`${VERIFIER_PREFIX}${userId}`, encrypted)
}

export function loadVerifier(userId: string): string | null {
  return localStorage.getItem(`${VERIFIER_PREFIX}${userId}`)
}

export async function createVerifier(userId: string, masterPassword: string): Promise<void> {
  const salt = await getPrimarySalt(userId)
  const key = await deriveVaultKey(masterPassword, salt)
  const payload: NotePayload = { notes: VERIFIER_MARKER }
  const blob = await encryptPayload(key, payload)
  storeVerifier(userId, bytesToBase64(blob))
}

async function verifyWithSalt(
  masterPassword: string,
  salt: Uint8Array,
  storedVerifier: string,
): Promise<boolean> {
  try {
    const key = await deriveVaultKey(masterPassword, salt)
    const payload = (await decryptPayload(key, base64ToBytes(storedVerifier))) as NotePayload
    return payload.notes === VERIFIER_MARKER
  } catch {
    return false
  }
}

async function verifyWithSample(
  masterPassword: string,
  salt: Uint8Array,
  encryptedSample: string,
): Promise<boolean> {
  try {
    const key = await deriveVaultKey(masterPassword, salt)
    await decryptPayload(key, base64ToBytes(encryptedSample))
    return true
  } catch {
    return false
  }
}

/** Returns the salt that validates the master password, or null if none match. */
export async function resolveSaltForPassword(
  userId: string,
  masterPassword: string,
  encryptedSample?: string,
): Promise<Uint8Array | null> {
  const storedVerifier = loadVerifier(userId)
  const candidates = await getSaltCandidates(userId)

  for (const salt of candidates) {
    if (storedVerifier) {
      if (await verifyWithSalt(masterPassword, salt, storedVerifier)) {
        return salt
      }
    } else if (encryptedSample) {
      if (await verifyWithSample(masterPassword, salt, encryptedSample)) {
        return salt
      }
    }
  }

  return null
}

export async function verifyMasterPassword(
  userId: string,
  masterPassword: string,
  encryptedSample?: string,
): Promise<boolean> {
  const salt = await resolveSaltForPassword(userId, masterPassword, encryptedSample)
  return salt != null
}

export function clearVerifier(userId: string): void {
  localStorage.removeItem(`${VERIFIER_PREFIX}${userId}`)
}

export async function usesLegacySalt(userId: string): Promise<boolean> {
  const legacy = loadLegacySalt(userId)
  if (!legacy) return false
  const deterministic = await deriveDeterministicSalt(userId)
  return !bytesEqual(legacy, deterministic)
}

export async function refreshVerifier(userId: string, masterPassword: string): Promise<void> {
  await createVerifier(userId, masterPassword)
}
