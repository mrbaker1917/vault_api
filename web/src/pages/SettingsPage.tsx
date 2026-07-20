import { FormEvent, useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import * as authApi from '../api/auth'
import * as mfaApi from '../api/mfa'
import * as recoveryApi from '../api/recovery'
import * as sessionsApi from '../api/sessions'
import { ApiError, formatRequestError } from '../api/client'
import type { Session } from '../api/types'
import { setMfaEnabledHint } from '../auth/mfa-hint'
import { useAuth } from '../auth/AuthContext'
import { MfaQrCode } from '../components/MfaQrCode'

export function SettingsPage() {
  const { user, refreshUser } = useAuth()

  useEffect(() => {
    void refreshUser()
  }, [refreshUser])

  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-3xl font-semibold text-white">Settings</h1>
        <p className="mt-2 text-sm text-slate-400">
          Manage account security, MFA, recovery codes, and active sessions.
        </p>
      </div>

      <MFASection mfaEnabled={user?.mfa_enabled ?? false} onUpdated={() => void refreshUser()} />
      <RecoverySection mfaEnabled={user?.mfa_enabled ?? false} />
      <PasswordSection mfaEnabled={user?.mfa_enabled ?? false} />
      <SessionsSection />

      <p className="text-sm text-slate-500">
        <Link to="/" className="text-emerald-400 hover:underline">
          Back to vault
        </Link>
      </p>
    </div>
  )
}

function Section({
  title,
  description,
  children,
}: {
  title: string
  description: string
  children: React.ReactNode
}) {
  return (
    <section className="rounded-xl border border-slate-800 bg-slate-900 p-6">
      <h2 className="text-lg font-medium text-white">{title}</h2>
      <p className="mt-1 text-sm text-slate-400">{description}</p>
      <div className="mt-4">{children}</div>
    </section>
  )
}

function MFASection({
  mfaEnabled,
  onUpdated,
}: {
  mfaEnabled: boolean
  onUpdated: () => void
}) {
  const [setup, setSetup] = useState<{ secret: string; otpauth_url: string } | null>(null)
  const [verifyCode, setVerifyCode] = useState('')
  const [disableCode, setDisableCode] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [message, setMessage] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleEnable() {
    setError(null)
    setMessage(null)
    setSubmitting(true)
    try {
      const result = await mfaApi.enableMFA()
      setSetup(result)
    } catch (err) {
      setError(formatRequestError(err, 'Failed to start MFA setup'))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleVerify(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await mfaApi.verifyMFA(verifyCode)
      setSetup(null)
      setVerifyCode('')
      setMfaEnabledHint(true)
      setMessage('Two-factor authentication is now enabled.')
      onUpdated()
    } catch (err) {
      setError(formatRequestError(err, 'Invalid authenticator code'))
    } finally {
      setSubmitting(false)
    }
  }

  async function handleDisable(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setMessage(null)
    setSubmitting(true)
    try {
      await mfaApi.disableMFA(disableCode)
      setDisableCode('')
      setMfaEnabledHint(false)
      setMessage('Two-factor authentication has been disabled.')
      onUpdated()
    } catch (err) {
      setError(formatRequestError(err, 'Failed to disable MFA'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Section
      title="Two-factor authentication"
      description="Protect your account with a TOTP authenticator app."
    >
      <p className="text-sm text-slate-300">
        Status:{' '}
        <span className={mfaEnabled ? 'text-emerald-400' : 'text-slate-400'}>
          {mfaEnabled ? 'Enabled' : 'Disabled'}
        </span>
      </p>

      {!mfaEnabled && !setup && (
        <button
          type="button"
          onClick={() => void handleEnable()}
          disabled={submitting}
          className="mt-4 rounded-md bg-emerald-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-emerald-400 disabled:opacity-60"
        >
          Enable MFA
        </button>
      )}

      {setup && (
        <div className="mt-4 space-y-4 rounded-lg border border-slate-800 bg-slate-950/60 p-4">
          <p className="text-sm text-slate-300">
            Scan this QR code or enter the secret in your authenticator app, then confirm with a
            code.
          </p>
          <MfaQrCode otpauthUrl={setup.otpauth_url} />
          <p className="break-all font-mono text-xs text-slate-400">{setup.secret}</p>
          <p className="text-xs text-slate-500">
            Or enter the secret above manually in your authenticator app.
          </p>
          <form onSubmit={(e) => void handleVerify(e)} className="flex gap-3">
            <input
              type="text"
              inputMode="numeric"
              required
              placeholder="6-digit code"
              value={verifyCode}
              onChange={(e) => setVerifyCode(e.target.value)}
              className="flex-1 rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-emerald-500"
            />
            <button
              type="submit"
              disabled={submitting}
              className="rounded-md bg-emerald-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-emerald-400 disabled:opacity-60"
            >
              Verify
            </button>
          </form>
        </div>
      )}

      {mfaEnabled && (
        <form onSubmit={(e) => void handleDisable(e)} className="mt-4 flex max-w-md gap-3">
          <input
            type="text"
            inputMode="numeric"
            required
            placeholder="Code to disable MFA"
            value={disableCode}
            onChange={(e) => setDisableCode(e.target.value)}
            className="flex-1 rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-emerald-500"
          />
          <button
            type="submit"
            disabled={submitting}
            className="rounded-md border border-red-800 px-4 py-2 text-sm text-red-300 hover:bg-red-950/40 disabled:opacity-60"
          >
            Disable
          </button>
        </form>
      )}

      {message && (
        <p className="mt-3 rounded-md border border-emerald-900/50 bg-emerald-950/30 px-3 py-2 text-sm text-emerald-300">
          {message}
        </p>
      )}
      {error && (
        <p className="mt-3 rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-300">
          {error}
        </p>
      )}
    </Section>
  )
}

function RecoverySection({ mfaEnabled }: { mfaEnabled: boolean }) {
  const [codes, setCodes] = useState<string[] | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleGenerate() {
    if (!window.confirm('This replaces any existing recovery codes. Continue?')) return
    setError(null)
    setSubmitting(true)
    try {
      const result = await recoveryApi.generateRecoveryCodes()
      setCodes(result.recovery_codes)
    } catch (err) {
      setError(formatRequestError(err, 'Failed to generate recovery codes'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Section
      title="Recovery codes"
      description="One-time codes to sign in when you cannot use your authenticator. MFA must be enabled."
    >
      {!mfaEnabled ? (
        <p className="text-sm text-slate-400">Enable MFA first to generate recovery codes.</p>
      ) : (
        <button
          type="button"
          onClick={() => void handleGenerate()}
          disabled={submitting}
          className="rounded-md border border-slate-700 px-4 py-2 text-sm hover:bg-slate-800 disabled:opacity-60"
        >
          {submitting ? 'Generating…' : 'Generate new codes'}
        </button>
      )}

      {codes && (
        <div className="mt-4 rounded-lg border border-amber-900/50 bg-amber-950/20 p-4">
          <p className="text-sm font-medium text-amber-200">Save these codes now — they won&apos;t be shown again.</p>
          <ul className="mt-3 grid gap-2 sm:grid-cols-2">
            {codes.map((code) => (
              <li key={code} className="rounded bg-slate-950 px-3 py-2 font-mono text-sm text-slate-200">
                {code}
              </li>
            ))}
          </ul>
        </div>
      )}

      {error && (
        <p className="mt-3 rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-300">
          {error}
        </p>
      )}
    </Section>
  )
}

function PasswordSection({ mfaEnabled }: { mfaEnabled: boolean }) {
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [needsTotp, setNeedsTotp] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [message, setMessage] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setMessage(null)

    if (newPassword !== confirmPassword) {
      setError('New passwords do not match')
      return
    }

    setSubmitting(true)
    try {
      await authApi.changePassword(
        currentPassword,
        newPassword,
        needsTotp || mfaEnabled ? totpCode : undefined,
      )
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      setTotpCode('')
      setNeedsTotp(false)
      setMessage('Password changed. Other sessions were signed out.')
    } catch (err) {
      if (err instanceof ApiError && err.mfaRequired) {
        setNeedsTotp(true)
        setError('Enter your authenticator code to confirm.')
      } else {
        setError(formatRequestError(err, 'Failed to change password'))
      }
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Section
      title="Account password"
      description="Change the password used to sign in to the API. This does not change your vault master password."
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="max-w-md space-y-3">
        <PasswordField label="Current password" value={currentPassword} onChange={setCurrentPassword} />
        <PasswordField label="New password" value={newPassword} onChange={setNewPassword} />
        <PasswordField label="Confirm new password" value={confirmPassword} onChange={setConfirmPassword} />
        {(mfaEnabled || needsTotp) && (
          <label className="block text-sm">
            <span className="text-slate-300">Authenticator code</span>
            <input
              type="text"
              inputMode="numeric"
              required
              value={totpCode}
              onChange={(e) => setTotpCode(e.target.value)}
              className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 outline-none focus:border-emerald-500"
            />
          </label>
        )}
        <button
          type="submit"
          disabled={submitting}
          className="rounded-md bg-emerald-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-emerald-400 disabled:opacity-60"
        >
          {submitting ? 'Updating…' : 'Change password'}
        </button>
      </form>
      {message && (
        <p className="mt-3 rounded-md border border-emerald-900/50 bg-emerald-950/30 px-3 py-2 text-sm text-emerald-300">
          {message}
        </p>
      )}
      {error && (
        <p className="mt-3 rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-300">
          {error}
        </p>
      )}
    </Section>
  )
}

function PasswordField({
  label,
  value,
  onChange,
}: {
  label: string
  value: string
  onChange: (value: string) => void
}) {
  return (
    <label className="block text-sm">
      <span className="text-slate-300">{label}</span>
      <input
        type="password"
        required
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 outline-none focus:border-emerald-500"
      />
    </label>
  )
}

function SessionsSection() {
  const [sessions, setSessions] = useState<Session[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadSessions = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      setSessions(await sessionsApi.listSessions())
    } catch (err) {
      setError(formatRequestError(err, 'Failed to load sessions'))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void loadSessions()
  }, [loadSessions])

  async function handleRevoke(session: Session) {
    if (session.is_current) return
    if (!window.confirm(`Revoke session on ${session.device_name || 'unknown device'}?`)) return
    try {
      await sessionsApi.revokeSession(session.id)
      await loadSessions()
    } catch (err) {
      setError(formatRequestError(err, 'Failed to revoke session'))
    }
  }

  return (
    <Section
      title="Active sessions"
      description="Devices currently signed in to your account."
    >
      {loading ? (
        <p className="text-sm text-slate-400">Loading sessions…</p>
      ) : sessions.length === 0 ? (
        <p className="text-sm text-slate-400">No active sessions.</p>
      ) : (
        <ul className="space-y-3">
          {sessions.map((session) => (
            <li
              key={session.id}
              className="flex flex-col gap-2 rounded-lg border border-slate-800 bg-slate-950/60 p-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <p className="font-medium text-white">
                  {session.device_name || 'Unknown device'}
                  {session.is_current && (
                    <span className="ml-2 text-xs text-emerald-400">(this device)</span>
                  )}
                </p>
                <p className="mt-1 text-xs text-slate-500">
                  {session.ip_address} · {new Date(session.created_at).toLocaleString()}
                </p>
              </div>
              {!session.is_current && (
                <button
                  type="button"
                  onClick={() => void handleRevoke(session)}
                  className="rounded-md border border-slate-700 px-3 py-1.5 text-sm hover:bg-slate-800"
                >
                  Revoke
                </button>
              )}
            </li>
          ))}
        </ul>
      )}
      {error && (
        <p className="mt-3 rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-300">
          {error}
        </p>
      )}
    </Section>
  )
}
