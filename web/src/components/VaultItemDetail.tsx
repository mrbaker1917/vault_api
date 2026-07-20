import { useState } from 'react'
import type {
  VaultItemType,
  VaultItemPayload,
  LoginPayload,
  CardPayload,
  IdentityPayload,
  NotePayload,
} from '../crypto/types'
import { copyToClipboard, normalizeUrl } from '../utils/clipboard'

type VaultItemDetailProps = {
  title: string
  itemType: VaultItemType
  folder: string
  tags: string[]
  payload: VaultItemPayload
  onEdit: () => void
  onDelete: () => void
}

export function VaultItemDetail({
  title,
  itemType,
  folder,
  tags,
  payload,
  onEdit,
  onDelete,
}: VaultItemDetailProps) {
  return (
    <div className="space-y-4">
      <div className="rounded-lg border border-slate-800 bg-slate-950/60 p-4">
        <h3 className="text-lg font-medium text-white">{title || 'Untitled'}</h3>
        <p className="mt-1 text-sm text-slate-400">
          {itemType}
          {folder ? ` · ${folder}` : ''}
        </p>
        {tags.length > 0 && (
          <div className="mt-3 flex flex-wrap gap-2">
            {tags.map((tag) => (
              <span
                key={tag}
                className="rounded-full bg-slate-800 px-2 py-0.5 text-xs text-slate-300"
              >
                {tag}
              </span>
            ))}
          </div>
        )}
      </div>

      <div className="space-y-3">
        {itemType === 'login' && <LoginFields payload={payload as LoginPayload} />}
        {itemType === 'note' && <NoteFields payload={payload as NotePayload} />}
        {itemType === 'card' && <CardFields payload={payload as CardPayload} />}
        {itemType === 'identity' && <IdentityFields payload={payload as IdentityPayload} />}
      </div>

      <div className="flex flex-wrap gap-3 border-t border-slate-800 pt-4">
        <button
          type="button"
          onClick={onEdit}
          className="rounded-md bg-emerald-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-emerald-400"
        >
          Edit
        </button>
        <button
          type="button"
          onClick={onDelete}
          className="rounded-md border border-red-800 px-4 py-2 text-sm text-red-300 hover:bg-red-950/40"
        >
          Delete
        </button>
      </div>
    </div>
  )
}

function LoginFields({ payload }: { payload: LoginPayload }) {
  const url = payload.url?.trim() ?? ''

  return (
    <>
      <DetailField label="Username" value={payload.username ?? ''} />
      <DetailField label="Password" value={payload.password ?? ''} secret />
      <DetailField label="URL" value={url} link={url ? normalizeUrl(url) : undefined} />
      <DetailField label="Notes" value={payload.notes ?? ''} multiline />
    </>
  )
}

function NoteFields({ payload }: { payload: NotePayload }) {
  return <DetailField label="Notes" value={payload.notes ?? ''} multiline />
}

function CardFields({ payload }: { payload: CardPayload }) {
  return (
    <>
      <DetailField label="Cardholder" value={payload.cardholder ?? ''} />
      <DetailField label="Number" value={payload.number ?? ''} secret />
      <DetailField label="Expiry" value={payload.expiry ?? ''} />
      <DetailField label="CVV" value={payload.cvv ?? ''} secret />
      <DetailField label="Notes" value={payload.notes ?? ''} multiline />
    </>
  )
}

function IdentityFields({ payload }: { payload: IdentityPayload }) {
  return (
    <>
      <DetailField label="Name" value={payload.name ?? ''} />
      <DetailField label="Email" value={payload.email ?? ''} />
      <DetailField label="Phone" value={payload.phone ?? ''} />
      <DetailField label="Address" value={payload.address ?? ''} multiline />
      <DetailField label="Notes" value={payload.notes ?? ''} multiline />
    </>
  )
}

function DetailField({
  label,
  value,
  secret = false,
  multiline = false,
  link,
}: {
  label: string
  value: string
  secret?: boolean
  multiline?: boolean
  link?: string
}) {
  const [visible, setVisible] = useState(!secret)
  const [copied, setCopied] = useState(false)
  const [copyError, setCopyError] = useState<string | null>(null)

  if (!value) {
    return null
  }

  async function handleCopy() {
    setCopyError(null)
    try {
      await copyToClipboard(value)
      setCopied(true)
      window.setTimeout(() => setCopied(false), 2000)
    } catch {
      setCopyError('Copy failed')
    }
  }

  const displayValue = secret && !visible ? '••••••••' : value

  return (
    <div className="rounded-lg border border-slate-800 bg-slate-950/60 p-3">
      <div className="flex items-center justify-between gap-3">
        <span className="text-sm font-medium text-slate-400">{label}</span>
        <div className="flex items-center gap-2">
          {secret && (
            <button
              type="button"
              onClick={() => setVisible((current) => !current)}
              className="rounded-md border border-slate-700 px-2 py-1 text-xs hover:bg-slate-800"
            >
              {visible ? 'Hide' : 'Show'}
            </button>
          )}
          <button
            type="button"
            onClick={() => void handleCopy()}
            className="rounded-md border border-slate-700 px-2 py-1 text-xs hover:bg-slate-800"
          >
            {copied ? 'Copied!' : 'Copy'}
          </button>
          {link && (
            <a
              href={link}
              target="_blank"
              rel="noopener noreferrer"
              className="rounded-md border border-slate-700 px-2 py-1 text-xs hover:bg-slate-800"
            >
              Open
            </a>
          )}
        </div>
      </div>
      <p
        className={`mt-2 break-all text-sm text-slate-200 ${multiline ? 'whitespace-pre-wrap' : ''}`}
      >
        {displayValue}
      </p>
      {copyError && <p className="mt-1 text-xs text-red-400">{copyError}</p>}
    </div>
  )
}
