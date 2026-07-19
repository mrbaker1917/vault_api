import { describe, expect, it } from 'vitest'
import { encryptPayload, decryptPayload } from './blob'
import { deriveVaultKey } from './kdf'
import { BLOB_VERSION } from './constants'

describe('vault crypto', () => {
  it('encrypts and decrypts a payload with version byte prefix', async () => {
    const salt = crypto.getRandomValues(new Uint8Array(16))
    const key = await deriveVaultKey('master-password-123', salt)
    const payload = { username: 'alice', password: 'secret', url: 'https://example.com' }

    const blob = await encryptPayload(key, payload)
    expect(blob[0]).toBe(BLOB_VERSION)

    const decrypted = await decryptPayload(key, blob)
    expect(decrypted).toEqual(payload)
  })

  it('rejects unsupported blob versions', async () => {
    const salt = crypto.getRandomValues(new Uint8Array(16))
    const key = await deriveVaultKey('master-password-123', salt)
    const blob = await encryptPayload(key, { notes: 'hello' })
    blob[0] = 0x02

    await expect(decryptPayload(key, blob)).rejects.toThrow(/unsupported encrypted blob version/)
  })
})
