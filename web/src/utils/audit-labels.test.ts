import { describe, expect, it } from 'vitest'
import { formatAuditAction } from './audit-labels'

describe('formatAuditAction', () => {
  it('maps known actions to readable labels', () => {
    expect(formatAuditAction('vault.item.delete')).toBe('Vault item deleted')
    expect(formatAuditAction('auth.login')).toBe('Signed in')
  })

  it('returns the raw action for unknown values', () => {
    expect(formatAuditAction('custom.action')).toBe('custom.action')
  })
})
