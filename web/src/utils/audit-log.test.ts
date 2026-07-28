import { describe, expect, it } from 'vitest'
import type { AuditLogEntry } from '../api/types'
import {
  auditEntriesToCsv,
  getAuditDateGroupLabel,
  getAuditEventLink,
  groupAuditEntriesByDate,
  parseAuditMetadata,
} from './audit-log'

describe('parseAuditMetadata', () => {
  it('formats known metadata fields', () => {
    expect(parseAuditMetadata({ title: 'GitHub', version: 2, mfa_used: true })).toEqual([
      { label: 'Title', value: 'GitHub' },
      { label: 'Version', value: '2' },
      { label: 'MFA used', value: 'Yes' },
    ])
  })
})

describe('getAuditEventLink', () => {
  it('links deleted vault items to trash', () => {
    expect(
      getAuditEventLink({
        id: '1',
        action: 'vault.item.delete',
        resource_type: 'vault_item',
        resource_id: 'abc',
        created_at: '2026-07-27T00:00:00Z',
      }),
    ).toEqual({ to: '/trash', label: 'View in trash' })
  })
})

describe('groupAuditEntriesByDate', () => {
  it('groups entries under readable labels', () => {
    const now = new Date('2026-07-27T12:00:00Z')
    const entries: AuditLogEntry[] = [
      { id: '1', action: 'auth.login', created_at: '2026-07-27T11:00:00Z' },
      { id: '2', action: 'auth.logout', created_at: '2026-07-26T11:00:00Z' },
    ]

    const groups = groupAuditEntriesByDate(entries, now)
    expect(groups.map((group) => group.label)).toEqual(['Today', 'Yesterday'])
  })
})

describe('getAuditDateGroupLabel', () => {
  it('labels recent dates', () => {
    const now = new Date('2026-07-27T12:00:00Z')
    expect(getAuditDateGroupLabel(new Date('2026-07-27T10:00:00Z'), now)).toBe('Today')
    expect(getAuditDateGroupLabel(new Date('2026-06-01T10:00:00Z'), now)).toBe('Older')
  })
})

describe('auditEntriesToCsv', () => {
  it('escapes csv values', () => {
    const csv = auditEntriesToCsv([
      {
        id: '1',
        action: 'auth.login',
        created_at: '2026-07-27T10:00:00Z',
        ip_address: '127.0.0.1',
        metadata: { device_name: 'Chrome, local' },
      },
    ])
    expect(csv).toContain('"Device: Chrome, local"')
  })
})
