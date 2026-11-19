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

// =========================================================================
// Phase 2: Core Productivity Types
// =========================================================================

export type TaskPriority = 'low' | 'medium' | 'high' | 'urgent';
export type TaskStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled';

export interface TaskItem {
  id: string;
  user_id?: string;
  title: string;
  description?: string;
  category: string;
  priority?: TaskPriority;
  status?: TaskStatus;
  deadline?: string;
  completed_at?: string;
  created_at?: string;
  updated_at?: string;

  // Backward compatibility fields with prototype/mock UI:
  categoryColor?: string;
  time?: string;
  completed?: boolean;
}

export interface CreateTaskPayload {
  title: string;
  description?: string;
  category?: string;
  priority?: TaskPriority;
  deadline?: string;
}

export interface UpdateTaskPayload {
  title?: string;
  description?: string;
  category?: string;
  priority?: TaskPriority;
  status?: TaskStatus;
  deadline?: string;
}

export type StudySessionStatus = 'active' | 'paused' | 'completed' | 'cancelled';

export interface StudySession {
  id: string;
  user_id: string;
  task_id?: string;
  subject: string;
  duration_seconds: number;
  status: StudySessionStatus;
  started_at: string;
  ended_at?: string;
  created_at: string;
  task?: TaskItem;
}

export interface StudyDailyStats {
  total_completed_seconds: number;
  completed_sessions_count: number;
}

export type FocusSessionType = 'focus' | 'break';
export type FocusSessionStatus = 'active' | 'completed' | 'abandoned';

export interface FocusSession {
  id: string;
  user_id: string;
  task_id?: string;
  session_type: FocusSessionType;
  started_at: string;
  ended_at?: string;
  duration_minutes: number;
  status: FocusSessionStatus;
  interruption_count: number;
  created_at: string;
  updated_at?: string;
  task?: TaskItem;

  // Helper calculation field
  target_minutes?: number;
}

export interface FocusDailyStats {
  completed_focus_minutes: number;
  completed_focus_count: number;
  completed_breaks_count: number;
  total_interruptions: number;
}

export interface LifestyleRecord {
  id: string;
  user_id: string;
  record_date: string;
  sleep_hours?: number;
  phone_usage_hours?: number;
  social_media_hours?: number;
  youtube_hours?: number;
  gaming_hours?: number;
  coffee_intake_mg?: number;
  exercise_minutes?: number;
  stress_level?: number;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface UpsertLifestylePayload {
  record_date: string;
  sleep_hours?: number;
  phone_usage_hours?: number;
  social_media_hours?: number;
  youtube_hours?: number;
  gaming_hours?: number;
  coffee_intake_mg?: number;
  exercise_minutes?: number;
  stress_level?: number;
  notes?: string;
}

export interface DashboardSummary {
  total_tasks: number;
  completed_tasks: number;
  pending_tasks: number;
  in_progress_tasks: number;
  completion_percentage: number;
  total_study_seconds: number;
  total_study_formatted: string;
  completed_study_sessions: number;
  completed_focus_minutes: number;
  completed_focus_count: number;
  completed_breaks_count: number;
  total_interruptions: number;
  recent_tasks: TaskItem[];
  upcoming_deadlines: TaskItem[];
  today_lifestyle_logged: boolean;
  today_lifestyle?: LifestyleRecord;
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

export interface UpcomingSchedule {
  id: string;
  title: string;
  timeRange: string;
  description?: string;
  type?: 'meeting' | 'study' | 'review';
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
