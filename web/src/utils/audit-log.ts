import type { AuditLogEntry } from '../api/types'
import { formatAuditAction } from './audit-labels'

export const AUDIT_CATEGORIES = [
  { value: '', label: 'All categories' },
  { value: 'auth', label: 'Authentication' },
  { value: 'vault', label: 'Vault' },
  { value: 'mfa', label: 'MFA' },
  { value: 'recovery', label: 'Recovery' },
] as const

export const AUDIT_ACTIONS = [
  { value: '', label: 'All actions' },
  { value: 'auth.signup', label: 'Account created' },
  { value: 'auth.login', label: 'Signed in' },
  { value: 'auth.logout', label: 'Signed out' },
  { value: 'auth.password.change', label: 'Password changed' },
  { value: 'auth.session.revoke', label: 'Session revoked' },
  { value: 'vault.item.create', label: 'Vault item created' },
  { value: 'vault.item.update', label: 'Vault item updated' },
  { value: 'vault.item.delete', label: 'Vault item deleted' },
  { value: 'vault.item.restore', label: 'Vault item restored' },
  { value: 'vault.item.share', label: 'Vault item shared' },
  { value: 'vault.item.share.revoke', label: 'Vault share revoked' },
  { value: 'mfa.enable', label: 'MFA enabled' },
  { value: 'mfa.verify', label: 'MFA verified' },
  { value: 'mfa.disable', label: 'MFA disabled' },
  { value: 'recovery.generate', label: 'Recovery codes generated' },
  { value: 'recovery.verify', label: 'Recovery sign-in' },
] as const

export const DATE_RANGE_OPTIONS = [
  { value: 7, label: 'Last 7 days' },
  { value: 30, label: 'Last 30 days' },
  { value: 90, label: 'Last 90 days' },
  { value: 0, label: 'All time' },
] as const

export type AuditListParams = {
  category?: string
  action?: string
  since?: string
  limit?: number
  offset?: number
}

export function buildAuditSince(days: number): string | undefined {
  if (days <= 0) {
    return undefined
  }
  const since = new Date()
  since.setDate(since.getDate() - days)
  return since.toISOString()
}

const METADATA_LABELS: Record<string, string> = {
  title: 'Title',
  item_type: 'Type',
  version: 'Version',
  email: 'Email',
  device_name: 'Device',
  mfa_used: 'MFA used',
  code_count: 'Codes generated',
  mfa_reset: 'MFA reset',
}

export type AuditMetadataField = {
  label: string
  value: string
}

export function parseAuditMetadata(metadata: unknown): AuditMetadataField[] {
  if (metadata == null) {
    return []
  }

  let record: Record<string, unknown>
  if (typeof metadata === 'string') {
    try {
      record = JSON.parse(metadata) as Record<string, unknown>
    } catch {
      return [{ label: 'Details', value: metadata }]
    }
  } else if (typeof metadata === 'object') {
    record = metadata as Record<string, unknown>
  } else {
    return [{ label: 'Details', value: String(metadata) }]
  }

  return Object.entries(record).map(([key, value]) => ({
    label: METADATA_LABELS[key] ?? key.replace(/_/g, ' '),
    value: formatMetadataValue(value),
  }))
}

function formatMetadataValue(value: unknown): string {
  if (typeof value === 'boolean') {
    return value ? 'Yes' : 'No'
  }
  if (value == null) {
    return ''
  }
  return String(value)
}

export type AuditEventLink = {
  to: string
  label: string
}

export function getAuditEventLink(entry: AuditLogEntry): AuditEventLink | null {
  if (entry.resource_type !== 'vault_item' || !entry.resource_id) {
    return null
  }

  switch (entry.action) {
    case 'vault.item.delete':
      return { to: '/trash', label: 'View in trash' }
    case 'vault.item.restore':
      return { to: '/', label: 'View in vault' }
    case 'vault.item.create':
    case 'vault.item.update':
    case 'vault.item.share':
    case 'vault.item.share.revoke':
      return { to: `/?item=${entry.resource_id}`, label: 'View item' }
    default:
      return null
  }
}

export function groupAuditEntriesByDate(
  entries: AuditLogEntry[],
  now = new Date(),
): Array<{ label: string; entries: AuditLogEntry[] }> {
  const groups = new Map<string, AuditLogEntry[]>()
  const order: string[] = []

  for (const entry of entries) {
    const label = getAuditDateGroupLabel(new Date(entry.created_at), now)
    if (!groups.has(label)) {
      groups.set(label, [])
      order.push(label)
    }
    groups.get(label)!.push(entry)
  }

  return order.map((label) => ({ label, entries: groups.get(label)! }))
}

export function getAuditDateGroupLabel(date: Date, now = new Date()): string {
  const startOfToday = startOfDay(now)
  const startOfYesterday = new Date(startOfToday)
  startOfYesterday.setDate(startOfYesterday.getDate() - 1)
  const startOfWeek = new Date(startOfToday)
  startOfWeek.setDate(startOfWeek.getDate() - 7)
  const startOfMonth = new Date(startOfToday)
  startOfMonth.setDate(startOfMonth.getDate() - 30)

  if (date >= startOfToday) {
    return 'Today'
  }
  if (date >= startOfYesterday) {
    return 'Yesterday'
  }
  if (date >= startOfWeek) {
    return 'Last 7 days'
  }
  if (date >= startOfMonth) {
    return 'Last 30 days'
  }
  return 'Older'
}

function startOfDay(date: Date): Date {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate())
}

export function auditEntriesToCsv(entries: AuditLogEntry[]): string {
  const header = ['Time', 'Action', 'Resource type', 'Resource ID', 'IP address', 'User agent', 'Details']
  const rows = entries.map((entry) => {
    const metadata = parseAuditMetadata(entry.metadata)
      .map((field) => `${field.label}: ${field.value}`)
      .join('; ')
    return [
      entry.created_at,
      formatAuditAction(entry.action),
      entry.resource_type ?? '',
      entry.resource_id ?? '',
      entry.ip_address ?? '',
      entry.user_agent ?? '',
      metadata,
    ]
  })

  return [header, ...rows].map((row) => row.map(escapeCsvCell).join(',')).join('\n')
}

function escapeCsvCell(value: string): string {
  if (/[",\n]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`
  }
  return value
}

export function downloadAuditCsv(entries: AuditLogEntry[], filename = 'audit-log.csv'): void {
  const blob = new Blob([auditEntriesToCsv(entries)], { type: 'text/csv;charset=utf-8' })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = filename
  anchor.click()
  URL.revokeObjectURL(url)
}
