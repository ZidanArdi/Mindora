import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import {
  UserDashboardPage,
  AdminDashboardPage,
  LoginPage,
  RegisterPage,
  HomePage,
  ForbiddenPage,
  TasksPage,
  StudyTrackingPage,
  FocusTrackingPage,
  LifestyleTrackingPage,
  PlaceholderModulePage,
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

          {/* Student Productivity Core Modules (Guarded: User & Admin) */}
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <UserDashboardPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/tasks"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <TasksPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/study"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <StudyTrackingPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/focus"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <FocusTrackingPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/lifestyle"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <LifestyleTrackingPage />
              </ProtectedRoute>
            }
          />

          {/* Future Modules Placeholders (Guarded: User & Admin) */}
          <Route
            path="/analytics"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <PlaceholderModulePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/reports"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <PlaceholderModulePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/recommendations"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <PlaceholderModulePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/profile"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <PlaceholderModulePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/settings"
            element={
              <ProtectedRoute allowedRoles={['user', 'admin']}>
                <PlaceholderModulePage />
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
