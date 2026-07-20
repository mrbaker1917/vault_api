import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import * as auditApi from '../api/audit'
import { formatRequestError } from '../api/client'
import type { AuditLogEntry } from '../api/types'
import { formatAuditAction, formatAuditMetadata } from '../utils/audit-labels'

const PAGE_SIZE = 50

export function AuditLogPage() {
  const [entries, setEntries] = useState<AuditLogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasMore, setHasMore] = useState(false)

  const loadEntries = useCallback(async (offset = 0, append = false) => {
    if (append) {
      setLoadingMore(true)
    } else {
      setLoading(true)
    }
    setError(null)
    try {
      const batch = await auditApi.listAuditLogs({ limit: PAGE_SIZE, offset })
      setEntries((current) => (append ? [...current, ...batch] : batch))
      setHasMore(batch.length === PAGE_SIZE)
    } catch (err) {
      setError(formatRequestError(err, 'Failed to load audit log'))
    } finally {
      setLoading(false)
      setLoadingMore(false)
    }
  }, [])

  useEffect(() => {
    void loadEntries()
  }, [loadEntries])

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-semibold text-white">Audit log</h1>
        <p className="mt-2 text-sm text-slate-400">
          Recent account and vault activity. Metadata is stored server-side; vault secrets are never
          logged.
        </p>
      </div>

      {error && (
        <p className="rounded-md border border-red-900/50 bg-red-950/40 px-3 py-2 text-sm text-red-300">
          {error}
        </p>
      )}

      {loading ? (
        <p className="text-slate-400">Loading activity…</p>
      ) : entries.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-700 bg-slate-900/50 p-8 text-center text-sm text-slate-400">
          No activity recorded yet.
        </div>
      ) : (
        <div className="space-y-3">
          {entries.map((entry) => (
            <AuditEntry key={entry.id} entry={entry} />
          ))}
        </div>
      )}

      {hasMore && (
        <button
          type="button"
          disabled={loadingMore}
          onClick={() => void loadEntries(entries.length, true)}
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

function AuditEntry({ entry }: { entry: AuditLogEntry }) {
  const metadata = formatAuditMetadata(entry.metadata)
  const when = new Date(entry.created_at).toLocaleString()

  return (
    <article className="rounded-xl border border-slate-800 bg-slate-900 p-4">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h2 className="font-medium text-white">{formatAuditAction(entry.action)}</h2>
          <p className="mt-1 text-xs text-slate-500">{when}</p>
        </div>
        <p className="font-mono text-xs text-slate-500">{entry.action}</p>
      </div>

      <dl className="mt-3 grid gap-2 text-sm sm:grid-cols-2">
        {entry.resource_type && (
          <>
            <dt className="text-slate-500">Resource</dt>
            <dd className="text-slate-300">
              {entry.resource_type}
              {entry.resource_id ? ` · ${entry.resource_id}` : ''}
            </dd>
          </>
        )}
        {entry.ip_address && (
          <>
            <dt className="text-slate-500">IP address</dt>
            <dd className="font-mono text-slate-300">{entry.ip_address}</dd>
          </>
        )}
      </dl>

      {metadata && (
        <pre className="mt-3 overflow-x-auto rounded-md border border-slate-800 bg-slate-950/60 p-3 text-xs text-slate-300">
          {metadata}
        </pre>
      )}
    </article>
  )
}
