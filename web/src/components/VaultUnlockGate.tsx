import { FormEvent, useEffect, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { formatRequestError } from '../api/client'
import { useVault } from '../auth/VaultContext'

type UnlockLocationState = {
  autoUnlockPassword?: string
}

export function VaultUnlockGate({ children }: { children: React.ReactNode }) {
  const { unlocked, needsSetup, unlock, setupMasterPassword } = useVault()
  const location = useLocation()
  const navigate = useNavigate()
  const [masterPassword, setMasterPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    const state = location.state as UnlockLocationState | null
    const autoUnlockPassword = state?.autoUnlockPassword
    if (!autoUnlockPassword || unlocked) {
      return
    }

    void unlock(autoUnlockPassword)
      .then(() => {
        navigate('.', { replace: true, state: {} })
      })
      .catch(() => {
        navigate('.', { replace: true, state: {} })
      })
  }, [location.state, navigate, unlock, unlocked])

  if (unlocked) {
    return <>{children}</>
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)

    if (needsSetup && masterPassword !== confirmPassword) {
      setError('Passwords do not match')
      return
    }

    setSubmitting(true)
    try {
      if (needsSetup) {
        await setupMasterPassword(masterPassword)
      } else {
        await unlock(masterPassword)
      }
      setMasterPassword('')
      setConfirmPassword('')
    } catch (err) {
      setError(
        formatRequestError(
          err,
          needsSetup ? 'Could not set master password' : 'Incorrect master password',
        ),
      )
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-[60vh] items-center justify-center px-4">
      <div className="w-full max-w-md rounded-xl border border-slate-800 bg-slate-900 p-8 shadow-xl">
        <h1 className="text-2xl font-semibold text-white">
          {needsSetup ? 'Set master password' : 'Unlock vault'}
        </h1>
        <p className="mt-2 text-sm text-slate-400">
          {needsSetup
            ? 'Choose a master password to encrypt your vault. Use the same one on every device.'
            : 'Enter your master password to decrypt vault items on this device.'}
        </p>

        <form onSubmit={(e) => void handleSubmit(e)} className="mt-8 space-y-4">
          <label className="block text-sm">
            <span className="text-slate-300">Master password</span>
            <input
              type="password"
              required
              autoComplete={needsSetup ? 'new-password' : 'current-password'}
              value={masterPassword}
              onChange={(e) => setMasterPassword(e.target.value)}
              className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 outline-none focus:border-emerald-500"
            />
          </label>

          {needsSetup && (
            <label className="block text-sm">
              <span className="text-slate-300">Confirm master password</span>
              <input
                type="password"
                required
                autoComplete="new-password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 outline-none focus:border-emerald-500"
              />
            </label>
          )}

          {error && (
            <p className="rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-300">
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={submitting}
            className="w-full rounded-md bg-emerald-500 px-4 py-2 font-medium text-slate-950 hover:bg-emerald-400 disabled:opacity-60"
          >
            {submitting
              ? needsSetup
                ? 'Setting up…'
                : 'Unlocking…'
              : needsSetup
                ? 'Create vault'
                : 'Unlock'}
          </button>
        </form>

        <p className="mt-6 text-xs text-slate-500">
          Titles, folders, and tags are stored as readable metadata on the server. Only
          secret fields are encrypted client-side.
        </p>
      </div>
    </div>
  )
}
