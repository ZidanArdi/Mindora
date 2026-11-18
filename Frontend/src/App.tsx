import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { UserDashboardPage, AdminDashboardPage, LoginPage, RegisterPage } from './pages';
import { AuthProvider } from './context/AuthContext';
import { ProtectedRoute } from './components/auth/ProtectedRoute';

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Root & Authentication Routes (renders LoginPage at root / and /login) */}
          <Route path="/" element={<LoginPage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />

          {/* Student Productivity Dashboard (Guarded: User & Admin) */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <UserDashboardPage />
              </ProtectedRoute>
            }
          />

          {/* Admin Console Dashboard (Guarded: Strictly Admin) */}
          <Route
            path="/admin"
            element={
              <ProtectedRoute allowedRoles={['admin']}>
                <AdminDashboardPage />
              </ProtectedRoute>
            }
          />

          {/* Catch-all route */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
