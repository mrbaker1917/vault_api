import { bytesToBase64, base64ToBytes } from './encoding'

const SALT_PREFIX = 'vault_salt_'

export function hasStoredSalt(userId: string): boolean {
  return localStorage.getItem(`${SALT_PREFIX}${userId}`) != null
}

export function loadSalt(userId: string): Uint8Array {
  const stored = localStorage.getItem(`${SALT_PREFIX}${userId}`)
  if (!stored) {
    throw new Error('vault salt not configured')
  }
  return base64ToBytes(stored)
}

export function createAndStoreSalt(userId: string): Uint8Array {
  const salt = crypto.getRandomValues(new Uint8Array(16))
  localStorage.setItem(`${SALT_PREFIX}${userId}`, bytesToBase64(salt))
  return salt
}

export function clearStoredSalt(userId: string): void {
  localStorage.removeItem(`${SALT_PREFIX}${userId}`)
}
