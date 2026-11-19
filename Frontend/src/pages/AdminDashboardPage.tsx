import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { DashboardTemplate } from '../components/templates';
import {
  AdminOverviewRow,
  AdminActivityChart,
  AdminMembersTable,
} from '../components/organisms';
import {
  mockAdminUser,
  mockAdminKPIs,
  mockProductivityChart,
  mockRecentActivities,
  mockTeamMembers,
} from '../mocks';
import { useAuth } from '../hooks';
import type { UserProfile } from '../types';

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
        workspaceName: 'System Administration',
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
            <h3 className="text-sm font-bold text-text-primary">Department Workspaces</h3>
            <p className="text-xs text-text-secondary leading-relaxed">
              Mindora multi-workspace tenancy management. Academic cohorts, university departments, and study pods are isolated under RBAC policies.
            </p>
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 pt-2">
              {['Engineering Cohort 2026', 'Faculty of Science', 'Graduate Research Pod'].map((ws) => (
                <div key={ws} className="p-4 rounded-2xl bg-surface-muted/50 border border-border">
                  <span className="text-xs font-bold text-text-primary block">{ws}</span>
                  <span className="text-[10px] text-text-secondary">Active Tenant · 24 members</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeHash === 'roles' && (
          <div id="roles" className="p-6 rounded-3xl bg-white border border-border shadow-soft-xs space-y-3 animate-fadeIn">
            <h3 className="text-sm font-bold text-text-primary">Roles & Access Control Policies</h3>
            <p className="text-xs text-text-secondary leading-relaxed">
              Role definitions enforced via PostgreSQL advisory locks and JWT claims: <code className="text-emerald-700 bg-emerald-50 px-1 py-0.5 rounded font-mono">user</code> (student) and <code className="text-emerald-700 bg-emerald-50 px-1 py-0.5 rounded font-mono">admin</code> (console operator).
            </p>
          </div>
        )}

        {activeHash === 'settings' && (
          <div id="settings" className="p-6 rounded-3xl bg-white border border-border shadow-soft-xs space-y-3 animate-fadeIn">
            <h3 className="text-sm font-bold text-text-primary">System & Security Configuration</h3>
            <p className="text-xs text-text-secondary leading-relaxed">
              Environment configuration: Cloudflare Turnstile token verification, JWT secret rotation, and PostgreSQL pool health monitor.
            </p>
          </div>
        )}
      </div>
    </DashboardTemplate>
  );
}
