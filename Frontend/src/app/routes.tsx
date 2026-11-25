import { Routes, Route, Navigate } from 'react-router-dom';
import {
  UserDashboardPage,
  AdminDashboardPage,
  LoginPage,
  RegisterPage,
  HomePage,
  ForbiddenPage,
  TasksPage,
  CalendarPage,
  StudyTrackingPage,
  FocusTrackingPage,
  LifestyleTrackingPage,
  AnalyticsPage,
  RecommendationsPage,
  ReportsPage,
  MessagesPage,
  ProfilePage,
  SettingsPage,
} from '../features';
import { ProtectedRoute } from '../components/auth/ProtectedRoute';
import { useAuth } from '../hooks';

function UnknownRouteRedirect() {
  const { user, isAuthenticated } = useAuth();
  if (isAuthenticated && user) {
    return <Navigate to={user.role === 'admin' ? '/admin' : '/dashboard'} replace />;
  }
  return <Navigate to="/" replace />;
}

export function AppRoutes() {
  return (
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
          <ProtectedRoute allowedRoles={['user']}>
            <UserDashboardPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/tasks"
        element={
          <ProtectedRoute allowedRoles={['user']}>
            <TasksPage />
          </ProtectedRoute>
        }
      />
      {/* Legacy & Shortcut Alias for Tasks */}
      <Route path="/list" element={<Navigate to="/tasks" replace />} />

      <Route
        path="/calendar"
        element={
          <ProtectedRoute allowedRoles={['user']}>
            <CalendarPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/study"
        element={
          <ProtectedRoute allowedRoles={['user']}>
            <StudyTrackingPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/focus"
        element={
          <ProtectedRoute allowedRoles={['user']}>
            <FocusTrackingPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/lifestyle"
        element={
          <ProtectedRoute allowedRoles={['user']}>
            <LifestyleTrackingPage />
          </ProtectedRoute>
        }
      />

      {/* Phase 3 Analytics & ML Prediction (Guarded: User & Admin) */}
      <Route
        path="/analytics"
        element={
          <ProtectedRoute allowedRoles={['user']}>
            <AnalyticsPage />
          </ProtectedRoute>
        }
      />

      {/* Phase 4 Modules Placeholders (Guarded: User & Admin) */}
      <Route
        path="/messages"
        element={
          <ProtectedRoute allowedRoles={['user']}>
            <MessagesPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/reports"
        element={
          <ProtectedRoute allowedRoles={['user']}>
            <ReportsPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/recommendations"
        element={
          <ProtectedRoute allowedRoles={['user']}>
            <RecommendationsPage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/profile"
        element={
          <ProtectedRoute allowedRoles={['user', 'admin']}>
            <ProfilePage />
          </ProtectedRoute>
        }
      />
      <Route
        path="/settings"
        element={
          <ProtectedRoute allowedRoles={['user', 'admin']}>
            <SettingsPage />
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
      <Route
        path="/admin/*"
        element={
          <ProtectedRoute allowedRoles={['admin']}>
            <AdminDashboardPage />
          </ProtectedRoute>
        }
      />

      {/* Catch-all route: smart redirect based on session status */}
      <Route path="*" element={<UnknownRouteRedirect />} />
    </Routes>
  );
}
