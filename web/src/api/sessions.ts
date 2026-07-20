import { apiFetch, apiJson } from './client'
import type { Session } from './types'

export async function listSessions(): Promise<Session[]> {
  return apiJson<Session[]>('/api/v1/auth/sessions')
}

export async function revokeSession(sessionId: string): Promise<void> {
  await apiFetch(`/api/v1/auth/sessions/${sessionId}`, { method: 'DELETE' })
}
