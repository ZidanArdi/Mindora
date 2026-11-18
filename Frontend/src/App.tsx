import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import {
  UserDashboardPage,
  AdminDashboardPage,
  LoginPage,
  RegisterPage,
  HomePage,
  ForbiddenPage,
} from './pages';
import { AuthProvider } from './context/AuthContext';
import { ProtectedRoute } from './components/auth/ProtectedRoute';

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Public Guest Routes */}
          <Route path="/" element={<HomePage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/register" element={<RegisterPage />} />
          <Route path="/forbidden" element={<ForbiddenPage />} />

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

          {/* Catch-all route: redirect unknown paths to landing page */}
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
