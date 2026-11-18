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

  return (
    <DashboardTemplate role="admin" currentUser={displayUser}>
      <div className="space-y-6">
        {/* Top Header & 4 KPI Cards */}
        <AdminOverviewRow
          userName={displayUser.name.split(' ')[0]}
          kpis={mockAdminKPIs}
        />

        {/* Productivity Bar Chart & Recent Activity Feed */}
        <AdminActivityChart
          chartData={mockProductivityChart}
          recentActivities={mockRecentActivities}
        />

        {/* Team Members Management Table (Privacy-compliant aggregated stats) */}
        <AdminMembersTable members={mockTeamMembers} />
      </div>
    </DashboardTemplate>
  );
}
