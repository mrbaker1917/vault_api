const MFA_ENABLED_KEY = 'vault_account_mfa_enabled'

export function setMfaEnabledHint(enabled: boolean): void {
  sessionStorage.setItem(MFA_ENABLED_KEY, enabled ? 'true' : 'false')
}

export function getMfaEnabledHint(): boolean | null {
  const value = sessionStorage.getItem(MFA_ENABLED_KEY)
  if (value === 'true') return true
  if (value === 'false') return false
  return null
}

export function clearMfaEnabledHint(): void {
  sessionStorage.removeItem(MFA_ENABLED_KEY)
}

export function resolveMfaEnabled(apiValue: boolean | undefined): boolean {
  if (typeof apiValue === 'boolean') {
    setMfaEnabledHint(apiValue)
    return apiValue
  }
  return getMfaEnabledHint() ?? false
}
