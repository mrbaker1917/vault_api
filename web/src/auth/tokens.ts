const ACCESS_TOKEN_KEY = 'vault_access_token'
const REFRESH_TOKEN_KEY = 'vault_refresh_token'
const EXPIRES_AT_KEY = 'vault_token_expires_at'

/** Access token lifetime on the server (15 minutes). */
const ACCESS_TOKEN_TTL_MS = 15 * 60 * 1000
/** Refresh one minute before expiry. */
const REFRESH_BUFFER_MS = 60 * 1000

export function getAccessToken(): string | null {
  return sessionStorage.getItem(ACCESS_TOKEN_KEY)
}

export function getRefreshToken(): string | null {
  return sessionStorage.getItem(REFRESH_TOKEN_KEY)
}

export function setTokens(accessToken: string, refreshToken: string): void {
  sessionStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  sessionStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  sessionStorage.setItem(
    EXPIRES_AT_KEY,
    String(Date.now() + ACCESS_TOKEN_TTL_MS - REFRESH_BUFFER_MS),
  )
}

export function clearTokens(): void {
  sessionStorage.removeItem(ACCESS_TOKEN_KEY)
  sessionStorage.removeItem(REFRESH_TOKEN_KEY)
  sessionStorage.removeItem(EXPIRES_AT_KEY)
}

export function shouldRefreshAccessToken(): boolean {
  const expiresAt = sessionStorage.getItem(EXPIRES_AT_KEY)
  if (!expiresAt || !getRefreshToken()) {
    return false
  }
  return Date.now() >= Number(expiresAt)
}

export function markAccessTokenRefreshed(accessToken: string): void {
  sessionStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  sessionStorage.setItem(
    EXPIRES_AT_KEY,
    String(Date.now() + ACCESS_TOKEN_TTL_MS - REFRESH_BUFFER_MS),
  )
}
