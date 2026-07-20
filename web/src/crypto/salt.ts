import { bytesToBase64, base64ToBytes } from './encoding'

const LEGACY_SALT_PREFIX = 'vault_salt_'
const SALT_VERSION = 'v1'

export function bytesEqual(a: Uint8Array, b: Uint8Array): boolean {
  if (a.length !== b.length) return false
  for (let i = 0; i < a.length; i += 1) {
    if (a[i] !== b[i]) return false
  }
  return true
}

/** Same salt on every device for a given user (not secret — only the master password is). */
export async function deriveDeterministicSalt(userId: string): Promise<Uint8Array> {
  const data = new TextEncoder().encode(`vault_api:${SALT_VERSION}:${userId}`)
  const hash = await crypto.subtle.digest('SHA-256', data)
  return new Uint8Array(hash).slice(0, 16)
}

export function loadLegacySalt(userId: string): Uint8Array | null {
  const stored = localStorage.getItem(`${LEGACY_SALT_PREFIX}${userId}`)
  if (!stored) return null
  return base64ToBytes(stored)
}

/** Salts to try when unlocking (legacy random salt first, then portable deterministic). */
export async function getSaltCandidates(userId: string): Promise<Uint8Array[]> {
  const deterministic = await deriveDeterministicSalt(userId)
  const legacy = loadLegacySalt(userId)
  if (legacy && !bytesEqual(legacy, deterministic)) {
    return [legacy, deterministic]
  }
  return [deterministic]
}

export async function getPrimarySalt(userId: string): Promise<Uint8Array> {
  return deriveDeterministicSalt(userId)
}

export function clearLegacySalt(userId: string): void {
  localStorage.removeItem(`${LEGACY_SALT_PREFIX}${userId}`)
}

export function hasLegacySalt(userId: string): boolean {
  return localStorage.getItem(`${LEGACY_SALT_PREFIX}${userId}`) != null
}

/** @deprecated Random per-browser salt — kept for migration reads only. */
export function createLegacySalt(userId: string): Uint8Array {
  const salt = crypto.getRandomValues(new Uint8Array(16))
  localStorage.setItem(`${LEGACY_SALT_PREFIX}${userId}`, bytesToBase64(salt))
  return salt
}
