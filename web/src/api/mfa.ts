import { apiFetch, apiJson } from './client'

export async function enableMFA(password: string): Promise<{ secret: string; otpauth_url: string }> {
  return apiJson('/api/v1/mfa/enable', {
    method: 'POST',
    body: JSON.stringify({ password }),
  })
}

export async function verifyMFA(code: string): Promise<void> {
  await apiFetch('/api/v1/mfa/verify', {
    method: 'POST',
    body: JSON.stringify({ code }),
  })
}

export async function disableMFA(code: string): Promise<void> {
  await apiFetch('/api/v1/mfa/disable', {
    method: 'POST',
    body: JSON.stringify({ code }),
  })
}
