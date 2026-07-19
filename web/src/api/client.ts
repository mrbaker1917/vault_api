import {
  clearTokens,
  getAccessToken,
  getRefreshToken,
  markAccessTokenRefreshed,
  shouldRefreshAccessToken,
} from '../auth/tokens'
import type { MFARequiredBody } from './types'

export const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8081'

export class ApiError extends Error {
  status: number
  mfaRequired: boolean

  constructor(status: number, message: string, mfaRequired = false) {
    super(message)
    this.status = status
    this.mfaRequired = mfaRequired
  }
}

export function formatRequestError(err: unknown, fallback: string): string {
  if (err instanceof ApiError) {
    return err.message
  }
  if (err instanceof TypeError) {
    return `Could not reach the API at ${API_URL}. Check that the server is running.`
  }
  if (err instanceof Error && err.message) {
    return err.message
  }
  return fallback
}

let refreshPromise: Promise<void> | null = null

async function parseApiError(res: Response): Promise<ApiError> {
  const text = await res.text()
  if (text) {
    try {
      const body = JSON.parse(text) as MFARequiredBody
      if (body.mfa_required) {
        return new ApiError(res.status, body.error ?? 'mfa required', true)
      }
    } catch {
      // plain-text error body
    }
    return new ApiError(res.status, text)
  }
  return new ApiError(res.status, res.statusText)
}

async function refreshAccessToken(): Promise<void> {
  if (!refreshPromise) {
    refreshPromise = (async () => {
      const refreshToken = getRefreshToken()
      if (!refreshToken) {
        clearTokens()
        throw new ApiError(401, 'session expired')
      }

      const res = await fetch(`${API_URL}/api/v1/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      })

      if (!res.ok) {
        clearTokens()
        throw await parseApiError(res)
      }

      const body = (await res.json()) as { access_token: string }
      markAccessTokenRefreshed(body.access_token)
    })().finally(() => {
      refreshPromise = null
    })
  }

  await refreshPromise
}

async function ensureFreshAccessToken(): Promise<void> {
  if (shouldRefreshAccessToken()) {
    await refreshAccessToken()
  }
}

export async function apiFetch(
  path: string,
  options: RequestInit = {},
  auth = true,
): Promise<Response> {
  if (auth) {
    await ensureFreshAccessToken()
  }

  const headers = new Headers(options.headers)
  if (options.body != null && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  if (auth) {
    const token = getAccessToken()
    if (token) {
      headers.set('Authorization', `Bearer ${token}`)
    }
  }

  let res = await fetch(`${API_URL}${path}`, { ...options, headers })

  if (res.status === 401 && auth && getRefreshToken()) {
    await refreshAccessToken()
    const token = getAccessToken()
    if (token) {
      headers.set('Authorization', `Bearer ${token}`)
    }
    res = await fetch(`${API_URL}${path}`, { ...options, headers })
  }

  if (!res.ok) {
    throw await parseApiError(res)
  }

  return res
}

export async function apiJson<T>(
  path: string,
  options: RequestInit = {},
  auth = true,
): Promise<T> {
  const res = await apiFetch(path, options, auth)
  if (res.status === 204) {
    return undefined as T
  }
  return (await res.json()) as T
}
