import { useEffect, useState } from 'react'
import type { VaultItemType, VaultItemPayload } from '../crypto/types'

export type VaultItemFormValues = {
  itemType: VaultItemType
  title: string
  folder: string
  tags: string
  payload: VaultItemPayload
}

type VaultItemFormProps = {
  initial?: Partial<VaultItemFormValues>
  submitLabel: string
  onSubmit: (values: VaultItemFormValues) => Promise<void>
  onCancel?: () => void
}

function emptyPayload(itemType: VaultItemType): VaultItemPayload {
  switch (itemType) {
    case 'login':
      return { username: '', password: '', url: '', notes: '' }
    case 'note':
      return { notes: '' }
    case 'card':
      return { cardholder: '', number: '', expiry: '', cvv: '', notes: '' }
    case 'identity':
      return { name: '', email: '', phone: '', address: '', notes: '' }
  }
}

export function VaultItemForm({
  initial,
  submitLabel,
  onSubmit,
  onCancel,
}: VaultItemFormProps) {
  const [itemType, setItemType] = useState<VaultItemType>(initial?.itemType ?? 'login')
  const [title, setTitle] = useState(initial?.title ?? '')
  const [folder, setFolder] = useState(initial?.folder ?? '')
  const [tags, setTags] = useState(initial?.tags ?? '')
  const [payload, setPayload] = useState<VaultItemPayload>(
    initial?.payload ?? emptyPayload(initial?.itemType ?? 'login'),
  )
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!initial?.payload) {
      setPayload(emptyPayload(itemType))
    }
  }, [itemType, initial?.payload])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await onSubmit({ itemType, title, folder, tags, payload })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Save failed')
    } finally {
      setSubmitting(false)
    }
  }

  function updatePayload(field: string, value: string) {
    setPayload((current) => ({ ...current, [field]: value }))
  }

  return (
    <form onSubmit={(e) => void handleSubmit(e)} className="space-y-4">
      <div className="grid gap-4 sm:grid-cols-2">
        <label className="block text-sm">
          <span className="text-slate-300">Type</span>
          <select
            value={itemType}
            onChange={(e) => setItemType(e.target.value as VaultItemType)}
            className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 outline-none focus:border-emerald-500"
          >
            <option value="login">Login</option>
            <option value="note">Note</option>
            <option value="card">Card</option>
            <option value="identity">Identity</option>
          </select>
        </label>

        <label className="block text-sm">
          <span className="text-slate-300">Title</span>
          <input
            required
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 outline-none focus:border-emerald-500"
          />
        </label>
      </div>

      <div className="grid gap-4 sm:grid-cols-2">
        <label className="block text-sm">
          <span className="text-slate-300">Folder</span>
          <input
            value={folder}
            onChange={(e) => setFolder(e.target.value)}
            className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 outline-none focus:border-emerald-500"
          />
        </label>

        <label className="block text-sm">
          <span className="text-slate-300">Tags (comma-separated)</span>
          <input
            value={tags}
            onChange={(e) => setTags(e.target.value)}
            className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 outline-none focus:border-emerald-500"
          />
        </label>
      </div>

      <div className="rounded-lg border border-slate-800 bg-slate-950/60 p-4">
        <h3 className="text-sm font-medium text-slate-300">Encrypted fields</h3>
        <div className="mt-3 space-y-3">
          {itemType === 'login' && (
            <>
              <Field label="Username" value={(payload as { username?: string }).username ?? ''} onChange={(v) => updatePayload('username', v)} />
              <Field label="Password" value={(payload as { password?: string }).password ?? ''} onChange={(v) => updatePayload('password', v)} secret />
              <Field label="URL" value={(payload as { url?: string }).url ?? ''} onChange={(v) => updatePayload('url', v)} />
              <Field label="Notes" value={(payload as { notes?: string }).notes ?? ''} onChange={(v) => updatePayload('notes', v)} multiline />
            </>
          )}
          {itemType === 'note' && (
            <Field label="Notes" value={(payload as { notes?: string }).notes ?? ''} onChange={(v) => updatePayload('notes', v)} multiline />
          )}
          {itemType === 'card' && (
            <>
              <Field label="Cardholder" value={(payload as { cardholder?: string }).cardholder ?? ''} onChange={(v) => updatePayload('cardholder', v)} />
              <Field label="Number" value={(payload as { number?: string }).number ?? ''} onChange={(v) => updatePayload('number', v)} secret />
              <Field label="Expiry" value={(payload as { expiry?: string }).expiry ?? ''} onChange={(v) => updatePayload('expiry', v)} />
              <Field label="CVV" value={(payload as { cvv?: string }).cvv ?? ''} onChange={(v) => updatePayload('cvv', v)} secret />
              <Field label="Notes" value={(payload as { notes?: string }).notes ?? ''} onChange={(v) => updatePayload('notes', v)} multiline />
            </>
          )}
          {itemType === 'identity' && (
            <>
              <Field label="Name" value={(payload as { name?: string }).name ?? ''} onChange={(v) => updatePayload('name', v)} />
              <Field label="Email" value={(payload as { email?: string }).email ?? ''} onChange={(v) => updatePayload('email', v)} />
              <Field label="Phone" value={(payload as { phone?: string }).phone ?? ''} onChange={(v) => updatePayload('phone', v)} />
              <Field label="Address" value={(payload as { address?: string }).address ?? ''} onChange={(v) => updatePayload('address', v)} multiline />
              <Field label="Notes" value={(payload as { notes?: string }).notes ?? ''} onChange={(v) => updatePayload('notes', v)} multiline />
            </>
          )}
        </div>
      </div>

      {error && (
        <p className="rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-300">
          {error}
        </p>
      )}

      <div className="flex justify-end gap-3">
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            className="rounded-md border border-slate-700 px-4 py-2 text-sm hover:bg-slate-800"
          >
            Cancel
          </button>
        )}
        <button
          type="submit"
          disabled={submitting}
          className="rounded-md bg-emerald-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-emerald-400 disabled:opacity-60"
        >
          {submitting ? 'Saving…' : submitLabel}
        </button>
      </div>
    </form>
  )
}

function Field({
  label,
  value,
  onChange,
  secret = false,
  multiline = false,
}: {
  label: string
  value: string
  onChange: (value: string) => void
  secret?: boolean
  multiline?: boolean
}) {
  return (
    <label className="block text-sm">
      <span className="text-slate-400">{label}</span>
      {multiline ? (
        <textarea
          value={value}
          onChange={(e) => onChange(e.target.value)}
          rows={3}
          className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 outline-none focus:border-emerald-500"
        />
      ) : (
        <input
          type={secret ? 'password' : 'text'}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 outline-none focus:border-emerald-500"
        />
      )}
    </label>
  )
}
