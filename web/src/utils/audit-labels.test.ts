import { describe, expect, it } from 'vitest'
import { formatAuditAction, formatAuditMetadata } from './audit-labels'

describe('formatAuditAction', () => {
  it('maps known actions to readable labels', () => {
    expect(formatAuditAction('vault.item.delete')).toBe('Vault item deleted')
    expect(formatAuditAction('auth.login')).toBe('Signed in')
  })

  it('returns the raw action for unknown values', () => {
    expect(formatAuditAction('custom.action')).toBe('custom.action')
  })
})

describe('formatAuditMetadata', () => {
  it('pretty-prints object metadata', () => {
    expect(formatAuditMetadata({ title: 'GitHub', version: 2 })).toBe(
      '{\n  "title": "GitHub",\n  "version": 2\n}',
    )
  })

  it('returns null for empty metadata', () => {
    expect(formatAuditMetadata(null)).toBeNull()
  })
})
