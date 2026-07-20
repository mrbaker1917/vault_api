const ACTION_LABELS: Record<string, string> = {
  'auth.signup': 'Account created',
  'auth.login': 'Signed in',
  'auth.logout': 'Signed out',
  'auth.password.change': 'Password changed',
  'auth.session.revoke': 'Session revoked',
  'vault.item.create': 'Vault item created',
  'vault.item.update': 'Vault item updated',
  'vault.item.delete': 'Vault item deleted',
  'vault.item.restore': 'Vault item restored',
  'vault.item.share': 'Vault item shared',
  'vault.item.share.revoke': 'Vault share revoked',
  'mfa.enable': 'MFA enabled',
  'mfa.verify': 'MFA verified',
  'mfa.disable': 'MFA disabled',
  'recovery.generate': 'Recovery codes generated',
  'recovery.verify': 'Recovery sign-in',
}

export function formatAuditAction(action: string): string {
  return ACTION_LABELS[action] ?? action
}

export function formatAuditMetadata(metadata: unknown): string | null {
  if (metadata == null) return null
  if (typeof metadata === 'string') {
    try {
      return JSON.stringify(JSON.parse(metadata), null, 2)
    } catch {
      return metadata
    }
  }
  try {
    return JSON.stringify(metadata, null, 2)
  } catch {
    return String(metadata)
  }
}
