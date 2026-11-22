import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { DashboardTemplate } from '../../components/templates';
import { AdminOverviewRow } from './AdminOverviewRow';
import { AdminActivityChart } from './AdminActivityChart';
import { AdminMembersTable } from './AdminMembersTable';
import {
  mockAdminUser,
  mockAdminKPIs,
  mockProductivityChart,
  mockRecentActivities,
  mockTeamMembers,
} from '../../mocks';
import { useAuth } from '../../hooks';
import type { UserProfile } from '../../types';

export default function AdminDashboardPage() {
  const { user } = useAuth();
  const location = useLocation();

  const displayUser: UserProfile = user
    ? {
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        avatarText: user.name
          .split(' ')
          .map((n) => n[0])
          .join('')
          .slice(0, 2)
          .toUpperCase() || 'AD',
        workspaceName: 'Administrasi Sistem',
      }
    : mockAdminUser;

  // Handle smooth scroll when hash changes
  useEffect(() => {
    if (location.hash) {
      const el = document.getElementById(location.hash.slice(1));
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'start' });
      }
    }
  }, [location.hash]);

  const activeHash = location.hash.replace('#', '');

  return (
    <DashboardTemplate role="admin" currentUser={displayUser}>
      <div className="space-y-6">
        {/* Top Header & 4 KPI Cards */}
        <div id="overview">
          <AdminOverviewRow
            userName={displayUser.name.split(' ')[0]}
            kpis={mockAdminKPIs}
          />
        </div>

        {/* Productivity Bar Chart & Recent Activity Feed */}
        <div id="reports">
          <AdminActivityChart
            chartData={mockProductivityChart}
            recentActivities={mockRecentActivities}
          />
        </div>

        {/* Team Members Management Table (Privacy-compliant aggregated stats) */}
        <div id="members">
          <AdminMembersTable members={mockTeamMembers} />
        </div>

        {/* Dedicated Admin Anchor View for Workspaces, Roles, and Settings */}
        {activeHash === 'workspaces' && (
          <div id="workspaces" className="p-6 rounded-3xl bg-white border border-border shadow-soft-xs space-y-3 animate-fadeIn">
            <h3 className="text-sm font-bold text-text-primary">Ruang Kerja Departemen</h3>
            <p className="text-xs text-text-secondary leading-relaxed">
              Manajemen multi-penyewa ruang kerja Mindora. Kohor akademik, departemen universitas, dan kelompok belajar diisolasi di bawah kebijakan RBAC.
            </p>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2">
              {['Angkatan Teknik 2026', 'Fakultas Sains', 'Kelompok Riset Pascasarjana'].map((ws) => (
                <div key={ws} className="p-4 rounded-2xl bg-surface-muted/50 border border-border">
                  <span className="text-xs font-bold text-text-primary block">{ws}</span>
                  <span className="text-[10px] text-text-secondary">Penyewa Aktif Â· 24 anggota</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeHash === 'roles' && (
          <div id="roles" className="p-6 rounded-3xl bg-white border border-border shadow-soft-xs space-y-3 animate-fadeIn">
            <h3 className="text-sm font-bold text-text-primary">Peran & Kebijakan Kontrol Akses</h3>
            <p className="text-xs text-text-secondary leading-relaxed">
              Definisi peran yang ditegakkan melalui klaim JWT: <code className="text-emerald-700 bg-emerald-50 px-1 py-0.5 rounded font-mono">user</code> (mahasiswa) dan <code className="text-emerald-700 bg-emerald-50 px-1 py-0.5 rounded font-mono">admin</code> (pengelola konsol).
            </p>
          </div>
        )}

        {activeHash === 'settings' && (
          <div id="settings" className="p-6 rounded-3xl bg-white border border-border shadow-soft-xs space-y-3 animate-fadeIn">
            <h3 className="text-sm font-bold text-text-primary">Konfigurasi Sistem & Keamanan</h3>
            <p className="text-xs text-text-secondary leading-relaxed">
              Konfigurasi lingkungan: verifikasi token otentikasi, rotasi kunci rahasia JWT, dan pemantauan koneksi PostgreSQL.
            </p>
          </div>
        )}
      </div>
    </DashboardTemplate>
  );
}

