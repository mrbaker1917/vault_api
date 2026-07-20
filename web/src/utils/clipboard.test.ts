import { describe, expect, it } from 'vitest'
import { normalizeUrl } from './clipboard'

describe('normalizeUrl', () => {
  it('adds https when scheme is missing', () => {
    expect(normalizeUrl('example.com')).toBe('https://example.com')
  })

  it('preserves existing https URLs', () => {
    expect(normalizeUrl('https://example.com/path')).toBe('https://example.com/path')
  })
})
