import { AES_GCM_IV_LENGTH, BLOB_VERSION } from './constants'
import type { VaultItemPayload } from './types'

export async function encryptPayload(
  key: CryptoKey,
  payload: VaultItemPayload,
): Promise<Uint8Array> {
  const iv = crypto.getRandomValues(new Uint8Array(AES_GCM_IV_LENGTH))
  const plaintext = new TextEncoder().encode(JSON.stringify(payload))
  const ciphertext = await crypto.subtle.encrypt({ name: 'AES-GCM', iv }, key, plaintext)

  const encrypted = new Uint8Array(1 + iv.length + ciphertext.byteLength)
  encrypted[0] = BLOB_VERSION
  encrypted.set(iv, 1)
  encrypted.set(new Uint8Array(ciphertext), 1 + iv.length)
  return encrypted
}

export async function decryptPayload(
  key: CryptoKey,
  blob: Uint8Array,
): Promise<VaultItemPayload> {
  if (blob.length < 1 + AES_GCM_IV_LENGTH + 1) {
    throw new Error('encrypted blob is too short')
  }
  if (blob[0] !== BLOB_VERSION) {
    throw new Error('unsupported encrypted blob version')
  }

  const iv = blob.slice(1, 1 + AES_GCM_IV_LENGTH)
  const ciphertext = blob.slice(1 + AES_GCM_IV_LENGTH)
  const plaintext = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, ciphertext)
  return JSON.parse(new TextDecoder().decode(plaintext)) as VaultItemPayload
}
