import type { UserProfile, TaskItem, UpcomingSchedule, NavCategory, WorkspaceItem } from '../types';

export const mockCurrentUser: UserProfile = {
  id: 'usr_01',
  name: 'John',
  email: 'john@mindora.app',
  role: 'user',
  avatarText: 'JD',
  workspaceName: 'Personal Workspace',
};

export const mockUserTasks: TaskItem[] = [
  {
    id: 'tsk_01',
    title: 'Cleaning up desk',
    category: 'Inbox',
    categoryColor: '#3B82F6',
    time: '08.20 AM',
    completed: false,
    priority: 'medium',
  },
  {
    id: 'tsk_02',
    title: 'Client meeting prep',
    category: 'Work',
    categoryColor: '#10B981',
    time: '10.20 AM',
    completed: false,
    priority: 'high',
  },
  {
    id: 'tsk_03',
    title: 'Small yoga',
    category: 'Health',
    categoryColor: '#EC4899',
    time: '02.30 PM',
    completed: false,
    priority: 'low',
  },
  {
    id: 'tsk_04',
    title: 'Review Machine Learning research paper',
    category: 'Study',
    categoryColor: '#F59E0B',
    time: '04.15 PM',
    completed: false,
    priority: 'high',
  },
  {
    id: 'tsk_05',
    title: 'Semester break travel itinerary planning',
    category: 'Holiday trip',
    categoryColor: '#8B5CF6',
    time: '07.00 PM',
    completed: true,
    priority: 'low',
  },
  {
    id: 'tsk_06',
    title: 'Project 3 documentation final check',
    category: 'Study',
    categoryColor: '#F59E0B',
    time: '09.00 PM',
    completed: false,
    priority: 'medium',
  },
];

export const mockUpcomingSchedules: UpcomingSchedule[] = [
  {
    id: 'sch_01',
    title: 'Design Team Meeting - Discuss new project wireframes',
    timeRange: '11:30 AM - 01:00 PM',
    type: 'meeting',
  },
  {
    id: 'sch_02',
    title: 'Client Meeting - Review project updates and seek approval',
    timeRange: '03:00 PM - 05:00 PM',
    type: 'meeting',
  },
];

export const mockMyLists: NavCategory[] = [
  { id: 'cat_inbox', name: 'Inbox', color: '#3B82F6', count: 4 },
  { id: 'cat_work', name: 'Work', color: '#10B981', count: 2 },
  { id: 'cat_study', name: 'Study', color: '#F59E0B', count: 5 },
  { id: 'cat_health', name: 'Health', color: '#EC4899', count: 1 },
  { id: 'cat_holiday', name: 'Holiday trip', color: '#8B5CF6', count: 1 },
];

export const mockWorkspaces: WorkspaceItem[] = [
  { id: 'ws_ui', name: 'UI Design', iconType: 'palette' },
  { id: 'ws_ateam', name: 'A-team Class', iconType: 'hash' },
];
