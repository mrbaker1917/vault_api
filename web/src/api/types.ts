export type TokenPair = {
  access_token: string
  refresh_token: string
}

export type MeResponse = {
  id: string
}

export type SignupResponse = {
  id: string
  email: string
}

export type MFARequiredBody = {
  error: string
  mfa_required: true
}
