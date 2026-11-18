export interface ApiResponse<T = unknown> {
  success: boolean;
  message: string;
  data?: T;
}

export type UserRole = 'user' | 'admin';

export interface UserResponse {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  status: string;
  created_at: string;
  permissions?: string[];
}

export interface AuthResult {
  access_token: string;
  token_type: string;
  expires_in: number;
  user: UserResponse;
}

export interface LoginPayload {
  email: string;
  password: string;
  turnstile_token?: string;
}

export interface RegisterPayload {
  name: string;
  email: string;
  password: string;
  turnstile_token?: string;
}

export interface UserProfile {
  id: string;
  name: string;
  email: string;
  role: UserRole;
  avatarText: string;
  workspaceName?: string;
}

export interface TaskItem {
  id: string;
  title: string;
  category: string;
  categoryColor: string;
  time: string;
  completed: boolean;
  priority?: 'low' | 'medium' | 'high';
}

export interface UpcomingSchedule {
  id: string;
  title: string;
  timeRange: string;
  description?: string;
  type?: 'meeting' | 'study' | 'review';
}

export interface NavCategory {
  id: string;
  name: string;
  color: string;
  count?: number;
}

export interface WorkspaceItem {
  id: string;
  name: string;
  iconType: 'palette' | 'hash' | 'folder';
}

export interface KPICardData {
  id: string;
  title: string;
  value: string;
  change: string;
  isPositive: boolean;
  badgeText?: string;
  iconType: 'users' | 'tasks' | 'rate' | 'workspaces';
}

export interface ActivityItem {
  id: string;
  title: string;
  meta: string;
  type: 'member_joined' | 'tasks_completed' | 'workspace_created' | 'permissions_updated';
}

export interface TeamMember {
  id: string;
  name: string;
  email: string;
  avatarText: string;
  role: 'Super admin' | 'Admin' | 'Manager' | 'Member';
  tasksCompleted: number;
  status: 'Active' | 'Away' | 'Inactive';
}

export interface ProductivityChartData {
  day: string;
  completed: number;
  created: number;
}
