export type TokenPair = {
  access_token: string
  refresh_token: string
}

export type MeResponse = {
  id: string
  mfa_enabled?: boolean
}

export type SignupResponse = {
  id: string
  email: string
}

export type MFARequiredBody = {
  error: string
  mfa_required: true
}

/** Go serializes domain objects with PascalCase field names. */
export type VaultItem = {
  ID: string
  UserID: string
  EncryptedData: string
  ItemType: string
  Title: string
  Folder: string
  Tags: string[]
  CreatedAt: string
  UpdatedAt: string
  DeletedAt?: string | null
  Version: number
}

export type VaultItemListResponse = {
  items: VaultItem[]
  total: number
  limit: number
  offset: number
}

export type VaultItemCreateRequest = {
  encrypted_data: string
  item_type: string
  title?: string
  folder?: string
  tags?: string[]
}

export type VaultItemUpdateRequest = VaultItemCreateRequest & {
  version: number
}

export type VersionRequest = {
  version: number
}

export type Session = {
  id: string
  device_name: string
  ip_address: string
  user_agent: string
  created_at: string
  expires_at: string
  is_current: boolean
}

export type MFASetupResponse = {
  secret: string
  otpauth_url: string
}

export type RecoveryCodesResponse = {
  recovery_codes: string[]
}
