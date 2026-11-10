import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { UserDashboardPage, AdminDashboardPage, LoginPage } from './pages';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Main User Productivity Dashboard */}
        <Route path="/" element={<UserDashboardPage />} />
        <Route path="/dashboard" element={<Navigate to="/" replace />} />

        {/* Authentication / Login Page */}
        <Route path="/login" element={<LoginPage />} />

        {/* Admin Console Dashboard */}
        <Route path="/admin" element={<AdminDashboardPage />} />

        {/* Catch-all route */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
