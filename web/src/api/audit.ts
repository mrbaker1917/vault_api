import { apiJson } from './client'
import type { AuditLogEntry } from './types'
import type { AuditListParams } from '../utils/audit-log'

export async function listAuditLogs(params: AuditListParams = {}): Promise<AuditLogEntry[]> {
  const query = new URLSearchParams()
  if (params.category) query.set('category', params.category)
  if (params.action) query.set('action', params.action)
  if (params.since) query.set('since', params.since)
  if (params.limit != null) query.set('limit', String(params.limit))
  if (params.offset != null) query.set('offset', String(params.offset))

  const suffix = query.size > 0 ? `?${query.toString()}` : ''
  return apiJson<AuditLogEntry[]>(`/api/v1/audit/logs${suffix}`)
}

export async function listAllAuditLogs(
  params: Omit<AuditListParams, 'limit' | 'offset'>,
  maxEntries = 1000,
): Promise<AuditLogEntry[]> {
  const pageSize = 100
  const collected: AuditLogEntry[] = []

  for (let offset = 0; offset < maxEntries; offset += pageSize) {
    const batch = await listAuditLogs({
      ...params,
      limit: pageSize,
      offset,
    })
    collected.push(...batch)
    if (batch.length < pageSize) {
      break
    }
  }

  return collected.slice(0, maxEntries)
}
