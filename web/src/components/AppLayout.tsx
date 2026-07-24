import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuth } from '../auth/AuthContext'
import { useVault } from '../auth/VaultContext'
import { IdleTimeoutManager } from './IdleTimeoutManager'

export function AppLayout() {
  const { user, logout } = useAuth()
  const { unlocked, lock } = useVault()
  const navigate = useNavigate()

  async function handleLogout() {
    lock()
    await logout()
    navigate('/login')
  }

  return (
    <div className="min-h-screen">
      <IdleTimeoutManager />
      <header className="border-b border-slate-800 bg-slate-900/80 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between gap-4 px-4 py-4">
          <Link to="/" className="text-lg font-semibold text-emerald-400">
            Vault
          </Link>
          <nav className="flex items-center gap-4 text-sm">
            <Link to="/" className="text-slate-300 hover:text-white">
              Vault
            </Link>
            <Link to="/settings" className="text-slate-300 hover:text-white">
              Settings
            </Link>
            <Link to="/trash" className="text-slate-300 hover:text-white">
              Trash
            </Link>
            <Link to="/audit" className="text-slate-300 hover:text-white">
              Audit
            </Link>
          </nav>
          <div className="flex items-center gap-4 text-sm text-slate-300">
            <span className="hidden font-mono sm:inline">{user?.id}</span>
            {unlocked && (
              <button
                type="button"
                onClick={lock}
                className="rounded-md border border-slate-700 px-3 py-1.5 hover:bg-slate-800"
              >
                Lock vault
              </button>
            )}
            <button
              type="button"
              onClick={() => void handleLogout()}
              className="rounded-md border border-slate-700 px-3 py-1.5 hover:bg-slate-800"
            >
              Log out
            </button>
          </div>
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-4 py-8">
        <Outlet />
      </main>
    </div>
  )
}
