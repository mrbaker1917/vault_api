import { Link } from 'react-router-dom'
import { useCallback, useEffect, useState } from 'react'
import * as vaultApi from '../api/vault'
import type { VaultItem } from '../api/types'
import { formatRequestError } from '../api/client'
import { useVault } from '../auth/VaultContext'
import type { VaultItemPayload } from '../crypto/types'
import { VaultItemDetail } from '../components/VaultItemDetail'
import { VaultItemForm, type VaultItemFormValues } from '../components/VaultItemForm'

type DecryptedVaultItem = VaultItem & {
  payload?: VaultItemPayload
  decryptFailed?: boolean
}

export function VaultPage() {
  const { encryptItemPayload, decryptItemPayload } = useVault()
  const [items, setItems] = useState<DecryptedVaultItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [viewing, setViewing] = useState<DecryptedVaultItem | null>(null)
  const [editing, setEditing] = useState<DecryptedVaultItem | null>(null)

  const loadItems = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const response = await vaultApi.listVaultItems({
        title: search.trim() || undefined,
        limit: 100,
      })
      const decrypted = await Promise.all(
        response.items.map(async (item) => {
          try {
            const payload = await decryptItemPayload(item.EncryptedData)
            return { ...item, payload }
          } catch {
            return { ...item, payload: undefined }
          }
        }),
      )
      setItems(decrypted.filter((item) => !item.DeletedAt))
    } catch (err) {
      setError(formatRequestError(err, 'Failed to load vault items'))
    } finally {
      setLoading(false)
    }
  }, [decryptItemPayload, search])

  useEffect(() => {
    void loadItems()
  }, [loadItems])

  async function handleCreate(values: VaultItemFormValues) {
    const encrypted_data = await encryptItemPayload(values.payload)
    await vaultApi.createVaultItem({
      encrypted_data,
      item_type: values.itemType,
      title: values.title,
      folder: values.folder,
      tags: parseTags(values.tags),
    })
    setShowCreate(false)
    await loadItems()
  }

  async function handleUpdate(values: VaultItemFormValues) {
    if (!editing || editing.payload == null) return
    const encrypted_data = await encryptItemPayload(values.payload)
    await vaultApi.updateVaultItem(editing.ID, {
      encrypted_data,
      item_type: values.itemType,
      title: values.title,
      folder: values.folder,
      tags: parseTags(values.tags),
      version: editing.Version,
    })
    setEditing(null)
    await loadItems()
  }

  async function handleDelete(item: DecryptedVaultItem) {
    if (!window.confirm(`Delete "${item.Title || 'Untitled'}"?`)) return
    await vaultApi.deleteVaultItem(item.ID, { version: item.Version })
    setViewing(null)
    setEditing(null)
    await loadItems()
  }

  function openItem(item: DecryptedVaultItem) {
    if (item.payload == null) {
      setViewing({ ...item, decryptFailed: true })
      return
    }
    setViewing(item)
  }

  function startEdit(item: DecryptedVaultItem) {
    setViewing(null)
    setEditing(item)
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-3xl font-semibold text-white">Vault</h1>
          <p className="mt-2 text-sm text-slate-400">
            Click an item to view secrets, copy fields, or open links.
          </p>
        </div>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className="rounded-md bg-emerald-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-emerald-400"
        >
          Add item
        </button>
      </div>

      <p className="text-sm text-slate-500">
        Deleted items can be restored from{' '}
        <Link to="/trash" className="text-emerald-400 hover:underline">
          Trash
        </Link>
        .
      </p>

      <div className="flex gap-3">
        <input
          type="search"
          placeholder="Search by title…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full max-w-md rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-emerald-500"
        />
        <button
          type="button"
          onClick={() => void loadItems()}
          className="rounded-md border border-slate-700 px-3 py-2 text-sm hover:bg-slate-800"
        >
          Search
        </button>
      </div>

      {error && (
        <p className="rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-300">
          {error}
        </p>
      )}

      {loading ? (
        <p className="text-slate-400">Loading items…</p>
      ) : items.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-700 bg-slate-900/50 p-8 text-center text-sm text-slate-400">
          No vault items yet. Add your first login or note to get started.
        </div>
      ) : (
        <div className="grid gap-3">
          {items.map((item) => (
            <button
              key={item.ID}
              type="button"
              onClick={() => openItem(item)}
              className="rounded-xl border border-slate-800 bg-slate-900 p-4 text-left transition hover:border-emerald-500/40 hover:bg-slate-900/80"
            >
              <div className="flex items-start justify-between gap-4">
                <div>
                  <h2 className="font-medium text-white">{item.Title || 'Untitled'}</h2>
                  <p className="mt-1 text-sm text-slate-400">
                    {item.ItemType}
                    {item.Folder ? ` · ${item.Folder}` : ''}
                  </p>
                </div>
                <span className="text-xs text-slate-500">v{item.Version}</span>
              </div>
              {item.payload == null && (
                <p className="mt-2 text-xs text-amber-400">Could not decrypt this item</p>
              )}
              {item.Tags?.length > 0 && (
                <div className="mt-3 flex flex-wrap gap-2">
                  {item.Tags.map((tag) => (
                    <span
                      key={tag}
                      className="rounded-full bg-slate-800 px-2 py-0.5 text-xs text-slate-300"
                    >
                      {tag}
                    </span>
                  ))}
                </div>
              )}
            </button>
          ))}
        </div>
      )}

      {showCreate && (
        <Modal title="Add vault item" onClose={() => setShowCreate(false)}>
          <VaultItemForm submitLabel="Create item" onSubmit={handleCreate} onCancel={() => setShowCreate(false)} />
        </Modal>
      )}

      {viewing && viewing.decryptFailed && (
        <Modal title="Cannot decrypt item" onClose={() => setViewing(null)}>
          <p className="text-sm text-slate-300">
            This item could not be decrypted with your current master password. You can delete it
            or unlock with the correct master password.
          </p>
          <button
            type="button"
            onClick={() => void handleDelete(viewing)}
            className="mt-4 text-sm text-red-400 hover:underline"
          >
            Delete item
          </button>
        </Modal>
      )}

      {viewing && !viewing.decryptFailed && viewing.payload != null && (
        <Modal title="Vault item" onClose={() => setViewing(null)}>
          <VaultItemDetail
            title={viewing.Title}
            itemType={viewing.ItemType as VaultItemFormValues['itemType']}
            folder={viewing.Folder}
            tags={viewing.Tags ?? []}
            payload={viewing.payload}
            onEdit={() => startEdit(viewing)}
            onDelete={() => void handleDelete(viewing)}
          />
        </Modal>
      )}

      {editing && editing.payload != null && (
        <Modal title="Edit vault item" onClose={() => setEditing(null)}>
          <VaultItemForm
            initial={{
              itemType: editing.ItemType as VaultItemFormValues['itemType'],
              title: editing.Title,
              folder: editing.Folder,
              tags: editing.Tags?.join(', ') ?? '',
              payload: editing.payload,
            }}
            submitLabel="Save changes"
            onSubmit={handleUpdate}
            onCancel={() => setEditing(null)}
          />
        </Modal>
      )}
    </div>
  )
}

function Modal({
  title,
  onClose,
  children,
}: {
  title: string
  onClose: () => void
  children: React.ReactNode
}) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 p-4">
      <div className="max-h-[90vh] w-full max-w-2xl overflow-y-auto rounded-xl border border-slate-800 bg-slate-900 p-6 shadow-xl">
        <div className="mb-6 flex items-center justify-between">
          <h2 className="text-xl font-semibold text-white">{title}</h2>
          <button
            type="button"
            onClick={onClose}
            className="rounded-md border border-slate-700 px-2 py-1 text-sm hover:bg-slate-800"
          >
            Close
          </button>
        </div>
        {children}
      </div>
    </div>
  )
}

function parseTags(value: string): string[] {
  return value
    .split(',')
    .map((tag) => tag.trim())
    .filter(Boolean)
}
