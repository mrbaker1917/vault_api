import { apiJson } from './client'
import { setTokens } from '../auth/tokens'
import type { TokenPair } from './types'

export async function generateRecoveryCodes(
  password: string,
  code: string,
): Promise<{ recovery_codes: string[] }> {
  return apiJson('/api/v1/recovery/generate', {
    method: 'POST',
    body: JSON.stringify({ password, code }),
  })
}

export async function recoveryLogin(
  email: string,
  password: string,
  recoveryCode: string,
  deviceName = 'vault-web',
): Promise<TokenPair> {
  const tokens = await apiJson<TokenPair>(
    '/api/v1/recovery/verify',
    {
      method: 'POST',
      body: JSON.stringify({
        email,
        password,
        recovery_code: recoveryCode,
        device_name: deviceName,
      }),
    },
    false,
  )
  setTokens(tokens.access_token, tokens.refresh_token)
  return tokens
}
