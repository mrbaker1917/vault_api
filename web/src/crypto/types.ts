export type VaultItemType = 'login' | 'note' | 'card' | 'identity'

export type LoginPayload = {
  username?: string
  password?: string
  url?: string
  notes?: string
}

export type NotePayload = {
  notes?: string
}

export type CardPayload = {
  cardholder?: string
  number?: string
  expiry?: string
  cvv?: string
  notes?: string
}

export type IdentityPayload = {
  name?: string
  email?: string
  phone?: string
  address?: string
  notes?: string
}

export type VaultItemPayload =
  | LoginPayload
  | NotePayload
  | CardPayload
  | IdentityPayload
