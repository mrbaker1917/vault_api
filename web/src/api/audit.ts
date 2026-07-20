import { apiJson } from './client'
import type { AuditLogEntry } from './types'

export async function listAuditLogs(params?: {
  limit?: number
  offset?: number
}): Promise<AuditLogEntry[]> {
  const query = new URLSearchParams()
  if (params?.limit != null) query.set('limit', String(params.limit))
  if (params?.offset != null) query.set('offset', String(params.offset))

  const suffix = query.size > 0 ? `?${query.toString()}` : ''
  return apiJson<AuditLogEntry[]>(`/api/v1/audit/logs${suffix}`)
}
