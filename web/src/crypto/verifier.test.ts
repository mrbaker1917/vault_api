import { beforeEach, describe, expect, it, vi } from 'vitest'
import { resolveSaltForPassword } from './verifier'
import { getPrimarySalt } from './salt'

describe('resolveSaltForPassword', () => {
  beforeEach(() => {
    vi.stubGlobal('localStorage', {
      getItem: () => null,
      setItem: () => undefined,
      removeItem: () => undefined,
    })
  })

  it('rejects any password when there is no verifier and no ciphertext sample', async () => {
    const userId = '00000000-0000-0000-0000-000000000099'
    const salt = await resolveSaltForPassword(userId, 'any-password')
    expect(salt).toBeNull()
  })

  it('returns primary salt for setup only via createVerifier flow', async () => {
    const salt = await getPrimarySalt('00000000-0000-0000-0000-000000000001')
    expect(salt).toHaveLength(16)
  })
})
