import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider } from './auth/AuthContext'
import { VaultProvider } from './auth/VaultContext'
import { AppLayout } from './components/AppLayout'
import { ProtectedRoute } from './components/ProtectedRoute'
import { VaultUnlockGate } from './components/VaultUnlockGate'
import { LoginPage } from './pages/LoginPage'
import { AuditLogPage } from './pages/AuditLogPage'
import { TrashPage } from './pages/TrashPage'
import { RecoveryLoginPage } from './pages/RecoveryLoginPage'
import { SettingsPage } from './pages/SettingsPage'
import { SignupPage } from './pages/SignupPage'
import { VaultPage } from './pages/VaultPage'

export default function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="/recovery" element={<RecoveryLoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
          <Route element={<ProtectedRoute />}>
            <Route
              element={
                <VaultProvider>
                  <AppLayout />
                </VaultProvider>
              }
            >
              <Route path="/settings" element={<SettingsPage />} />
              <Route path="/audit" element={<AuditLogPage />} />
              <Route path="/trash" element={<TrashPage />} />
              <Route
                path="/"
                element={
                  <VaultUnlockGate>
                    <VaultPage />
                  </VaultUnlockGate>
                }
              />
            </Route>
          </Route>
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  )
}
