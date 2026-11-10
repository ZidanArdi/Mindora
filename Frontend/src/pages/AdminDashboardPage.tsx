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

export default function AdminDashboardPage() {
  return (
    <DashboardTemplate role="admin" currentUser={mockAdminUser}>
      <div className="space-y-6">
        {/* Top Header & 4 KPI Cards */}
        <AdminOverviewRow
          userName={mockAdminUser.name.split(' ')[0]}
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
