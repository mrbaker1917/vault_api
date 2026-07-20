import { apiFetch, apiJson } from './client'
import type { MeResponse, SignupResponse, TokenPair } from './types'
import { clearTokens, setTokens } from '../auth/tokens'

export async function signup(email: string, password: string): Promise<SignupResponse> {
  return apiJson<SignupResponse>(
    '/api/v1/auth/signup',
    {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    },
    false,
  )
}

export async function login(
  email: string,
  password: string,
  totpCode?: string,
  deviceName = 'vault-web',
): Promise<TokenPair> {
  const body: Record<string, string> = {
    email,
    password,
    device_name: deviceName,
  }
  if (totpCode) {
    body.totp_code = totpCode
  }

  const tokens = await apiJson<TokenPair>(
    '/api/v1/auth/login',
    {
      method: 'POST',
      body: JSON.stringify(body),
    },
    false,
  )

  setTokens(tokens.access_token, tokens.refresh_token)
  return tokens
}

export async function fetchMe(): Promise<MeResponse> {
  return apiJson<MeResponse>('/api/v1/me')
}

export async function logout(): Promise<void> {
  try {
    await apiFetch('/api/v1/auth/logout', { method: 'POST' })
  } finally {
    clearTokens()
  }
}

export async function changePassword(
  currentPassword: string,
  newPassword: string,
  totpCode?: string,
): Promise<void> {
  const body: Record<string, string> = {
    current_password: currentPassword,
    new_password: newPassword,
  }
  if (totpCode) {
    body.totp_code = totpCode
  }
  await apiFetch('/api/v1/auth/change-password', {
    method: 'POST',
    body: JSON.stringify(body),
  })
}
