import type { UserProfile, KPICardData, ActivityItem, TeamMember, ProductivityChartData } from '../types';

export const mockAdminUser: UserProfile = {
  id: 'adm_01',
  name: 'John Doe',
  email: 'john.admin@mindora.app',
  role: 'admin',
  avatarText: 'JD',
  workspaceName: 'Admin Console',
};

export const mockAdminKPIs: KPICardData[] = [
  {
    id: 'kpi_members',
    title: 'Total anggota',
    value: '248',
    change: '+12.5%',
    isPositive: true,
    iconType: 'users',
  },
  {
    id: 'kpi_tasks',
    title: 'Tugas aktif',
    value: '1.429',
    change: '+8.2%',
    isPositive: true,
    iconType: 'tasks',
  },
  {
    id: 'kpi_rate',
    title: 'Tingkat penyelesaian',
    value: '84.6%',
    change: '+4.1%',
    isPositive: true,
    iconType: 'rate',
  },
  {
    id: 'kpi_workspaces',
    title: 'Ruang kerja aktif',
    value: '16',
    change: '+2 baru',
    isPositive: true,
    iconType: 'workspaces',
  },
];

export const mockProductivityChart: ProductivityChartData[] = [
  { day: 'Sen', completed: 68, created: 42 },
  { day: 'Sel', completed: 86, created: 45 },
  { day: 'Rab', completed: 72, created: 52 },
  { day: 'Kam', completed: 96, created: 64 },
  { day: 'Jum', completed: 84, created: 54 },
  { day: 'Sab', completed: 52, created: 32 },
  { day: 'Min', completed: 76, created: 38 },
];

export const mockRecentActivities: ActivityItem[] = [
  {
    id: 'act_01',
    title: 'Clara bergabung',
    meta: 'Ruang Kerja Desain UI · 8 mnt lalu',
    type: 'member_joined',
  },
  {
    id: 'act_02',
    title: '12 tugas diselesaikan',
    meta: 'Tim Produk · 24 mnt lalu',
    type: 'tasks_completed',
  },
  {
    id: 'act_03',
    title: 'Ruang kerja baru dibuat',
    meta: 'Pemasaran 2026 · 1 jam lalu',
    type: 'workspace_created',
  },
  {
    id: 'act_04',
    title: 'Hak akses peran diperbarui',
    meta: 'Akses Manajer · 2 jam lalu',
    type: 'permissions_updated',
  },
];

export const mockTeamMembers: TeamMember[] = [
  {
    id: 'mem_01',
    name: 'Alicia Keys',
    email: 'alicia@mindora.app',
    avatarText: 'AK',
    role: 'Admin',
    tasksCompleted: 42,
    status: 'Active',
  },
  {
    id: 'mem_02',
    name: 'Brian Miller',
    email: 'brian@mindora.app',
    avatarText: 'BM',
    role: 'Manager',
    tasksCompleted: 36,
    status: 'Active',
  },
  {
    id: 'mem_03',
    name: 'Clara Smith',
    email: 'clara@mindora.app',
    avatarText: 'CS',
    role: 'Member',
    tasksCompleted: 28,
    status: 'Active',
  },
  {
    id: 'mem_04',
    name: 'Daniel Jones',
    email: 'daniel@mindora.app',
    avatarText: 'DJ',
    role: 'Member',
    tasksCompleted: 19,
    status: 'Away',
  },
  {
    id: 'mem_05',
    name: 'Eva Green',
    email: 'eva@mindora.app',
    avatarText: 'EG',
    role: 'Member',
    tasksCompleted: 34,
    status: 'Active',
  },
];
