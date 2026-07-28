import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import * as auditApi from '../api/audit'
import { listAllAuditLogs } from '../api/audit'
import { formatRequestError } from '../api/client'
import type { AuditLogEntry } from '../api/types'
import { formatAuditAction } from '../utils/audit-labels'
import {
  AUDIT_ACTIONS,
  AUDIT_CATEGORIES,
  buildAuditSince,
  DATE_RANGE_OPTIONS,
  downloadAuditCsv,
  getAuditEventLink,
  groupAuditEntriesByDate,
  parseAuditMetadata,
} from '../utils/audit-log'

const PAGE_SIZE = 50

export function AuditLogPage() {
  const [entries, setEntries] = useState<AuditLogEntry[]>([])
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [exporting, setExporting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasMore, setHasMore] = useState(false)
  const [category, setCategory] = useState('')
  const [action, setAction] = useState('')
  const [days, setDays] = useState<number>(30)

  const queryParams = useMemo(
    () => ({
      category: category || undefined,
      action: action || undefined,
      since: buildAuditSince(days),
    }),
    [category, action, days],
  )

  const loadEntries = useCallback(
    async (offset = 0, append = false) => {
      if (append) {
        setLoadingMore(true)
      } else {
        setLoading(true)
      }
      setError(null)
      try {
        const batch = await auditApi.listAuditLogs({
          ...queryParams,
          limit: PAGE_SIZE,
          offset,
        })
        setEntries((current) => (append ? [...current, ...batch] : batch))
        setHasMore(batch.length === PAGE_SIZE)
      } catch (err) {
        setError(formatRequestError(err, 'Failed to load audit log'))
      } finally {
        setLoading(false)
        setLoadingMore(false)
      }
    },
    [queryParams],
  )

  useEffect(() => {
    void loadEntries()
  }, [loadEntries])

  const groupedEntries = useMemo(() => groupAuditEntriesByDate(entries), [entries])

  async function handleExport() {
    setExporting(true)
    setError(null)
    try {
      const allEntries = await listAllAuditLogs(queryParams)
      downloadAuditCsv(allEntries)
    } catch (err) {
      setError(formatRequestError(err, 'Failed to export audit log'))
    } finally {
      setExporting(false)
    }
  }

  const visibleActions = AUDIT_ACTIONS.filter(
    (option) =>
      !option.value ||
      !category ||
      option.value.startsWith(`${category}.`),
  )

  return (
    <div className="space-y-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <h1 className="text-3xl font-semibold text-white">Audit log</h1>
          <p className="mt-2 text-sm text-slate-400">
            Account and vault activity with filters, grouping, and CSV export. Vault secrets are
            never logged.
          </p>
        </div>
        <button
          type="button"
          disabled={exporting || loading}
          onClick={() => void handleExport()}
          className="rounded-md border border-slate-700 px-4 py-2 text-sm hover:bg-slate-800 disabled:opacity-50"
        >
          {exporting ? 'Exporting…' : 'Export CSV'}
        </button>
      </div>

      <div className="grid gap-3 sm:grid-cols-3">
        <label className="block text-sm">
          <span className="text-slate-400">Category</span>
          <select
            value={category}
            onChange={(e) => {
              setCategory(e.target.value)
              setAction('')
            }}
            className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-emerald-500"
          >
            {AUDIT_CATEGORIES.map((option) => (
              <option key={option.value || 'all'} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <label className="block text-sm">
          <span className="text-slate-400">Action</span>
          <select
            value={action}
            onChange={(e) => setAction(e.target.value)}
            className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-emerald-500"
          >
            {visibleActions.map((option) => (
              <option key={option.value || 'all-actions'} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>

        <label className="block text-sm">
          <span className="text-slate-400">Date range</span>
          <select
            value={days}
            onChange={(e) => setDays(Number(e.target.value))}
            className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-emerald-500"
          >
            {DATE_RANGE_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
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
          No activity matches these filters.
        </div>
      ) : (
        <div className="space-y-8">
          {groupedEntries.map((group) => (
            <section key={group.label} className="space-y-3">
              <h2 className="text-sm font-medium uppercase tracking-wide text-slate-500">
                {group.label}
              </h2>
              {group.entries.map((entry) => (
                <AuditEntry key={entry.id} entry={entry} />
              ))}
            </section>
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
  const metadata = parseAuditMetadata(entry.metadata)
  const when = new Date(entry.created_at).toLocaleString()
  const eventLink = getAuditEventLink(entry)

  return (
    <article className="rounded-xl border border-slate-800 bg-slate-900 p-4">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <h3 className="font-medium text-white">{formatAuditAction(entry.action)}</h3>
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
        {entry.user_agent && (
          <>
            <dt className="text-slate-500">User agent</dt>
            <dd className="break-all text-slate-300">{entry.user_agent}</dd>
          </>
        )}
      </dl>

      {metadata.length > 0 && (
        <dl className="mt-3 grid gap-2 rounded-md border border-slate-800 bg-slate-950/60 p-3 text-sm sm:grid-cols-2">
          {metadata.map((field) => (
            <div key={`${field.label}-${field.value}`}>
              <dt className="text-slate-500">{field.label}</dt>
              <dd className="text-slate-200">{field.value}</dd>
            </div>
          ))}
        </dl>
      )}

      {eventLink && (
        <Link
          to={eventLink.to}
          className="mt-3 inline-block text-sm text-emerald-400 hover:underline"
        >
          {eventLink.label}
        </Link>
      )}
    </article>
  )
}
