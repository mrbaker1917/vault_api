import { useAuth } from '../auth/AuthContext'

export function DashboardPage() {
  const { user } = useAuth()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-semibold text-white">Welcome</h1>
        <p className="mt-2 text-slate-400">
          Phase 1 complete — you are authenticated against the Vault API.
        </p>
      </div>

      <div className="rounded-xl border border-slate-800 bg-slate-900 p-6">
        <h2 className="text-sm font-medium uppercase tracking-wide text-slate-500">
          Signed in as
        </h2>
        <p className="mt-2 font-mono text-sm text-emerald-300">{user?.id}</p>
      </div>

      <div className="rounded-xl border border-dashed border-slate-700 bg-slate-900/50 p-6 text-sm text-slate-400">
        Phase 2 will add client-side encryption and vault item management here.
      </div>
    </div>
  )
}
