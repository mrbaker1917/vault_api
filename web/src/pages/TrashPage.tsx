import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import * as vaultApi from '../api/vault'
import { formatRequestError } from '../api/client'
import type { VaultItem } from '../api/types'

const PAGE_SIZE = 50

export function TrashPage() {
  const [items, setItems] = useState<VaultItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [restoringId, setRestoringId] = useState<string | null>(null)
  const [message, setMessage] = useState<string | null>(null)

  const loadItems = useCallback(async (offset = 0, append = false) => {
    if (append) {
      setLoadingMore(true)
    } else {
      setLoading(true)
    }
    setError(null)
    try {
      const response = await vaultApi.listDeletedVaultItems({ limit: PAGE_SIZE, offset })
      setItems((current) => (append ? [...current, ...response.items] : response.items))
      setTotal(response.total)
    } catch (err) {
      setError(formatRequestError(err, 'Failed to load deleted items'))
    } finally {
      setLoading(false)
      setLoadingMore(false)
    }
  }, [])

  useEffect(() => {
    void loadItems()
  }, [loadItems])

  async function handleRestore(item: VaultItem) {
    const title = item.Title || 'Untitled'
    if (!window.confirm(`Restore "${title}" to your vault?`)) return

    setRestoringId(item.ID)
    setError(null)
    setMessage(null)
    try {
      await vaultApi.restoreVaultItem(item.ID, { version: item.Version })
      setMessage(`Restored "${title}".`)
      await loadItems()
    } catch (err) {
      setError(formatRequestError(err, 'Failed to restore item'))
    } finally {
      setRestoringId(null)
    }
  }

  const hasMore = items.length < total

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-semibold text-white">Recently deleted</h1>
        <p className="mt-2 text-sm text-slate-400">
          Soft-deleted vault items can be restored here. Items are hidden from your vault until
          restored or permanently purged by a future retention job.
        </p>
      </div>

      {message && (
        <p className="rounded-md border border-emerald-900/50 bg-emerald-950/30 px-3 py-2 text-sm text-emerald-300">
          {message}
        </p>
      )}

      {error && (
        <p className="rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-300">
          {error}
        </p>
      )}

      {loading ? (
        <p className="text-slate-400">Loading deleted items…</p>
      ) : items.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-700 bg-slate-900/50 p-8 text-center text-sm text-slate-400">
          No deleted items. When you delete something from the vault, it appears here.
        </div>
      ) : (
        <div className="grid gap-3">
          {items.map((item) => (
            <div
              key={item.ID}
              className="flex flex-col gap-4 rounded-xl border border-slate-800 bg-slate-900 p-4 sm:flex-row sm:items-center sm:justify-between"
            >
              <div>
                <h2 className="font-medium text-white">{item.Title || 'Untitled'}</h2>
                <p className="mt-1 text-sm text-slate-400">
                  {item.ItemType}
                  {item.Folder ? ` · ${item.Folder}` : ''}
                </p>
                {item.DeletedAt && (
                  <p className="mt-2 text-xs text-slate-500">
                    Deleted {new Date(item.DeletedAt).toLocaleString()}
                  </p>
                )}
              </div>
              <button
                type="button"
                disabled={restoringId === item.ID}
                onClick={() => void handleRestore(item)}
                className="rounded-md bg-emerald-500 px-4 py-2 text-sm font-medium text-slate-950 hover:bg-emerald-400 disabled:opacity-50"
              >
                {restoringId === item.ID ? 'Restoring…' : 'Restore'}
              </button>
            </div>
          ))}
        </div>
      )}

      {hasMore && (
        <button
          type="button"
          disabled={loadingMore}
          onClick={() => void loadItems(items.length, true)}
          className="rounded-md border border-slate-700 px-4 py-2 text-sm hover:bg-slate-800 disabled:opacity-50"
        >
          {loadingMore ? 'Loading…' : 'Load more'}
        </button>
      )}

      <p className="text-sm text-slate-500">
        <Link to="/" className="text-emerald-400 hover:underline">
          Back to vault
        </Link>
      </p>
    </div>
  )
}
