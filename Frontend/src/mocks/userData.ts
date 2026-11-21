import type { UserProfile, TaskItem, UpcomingSchedule, NavCategory, WorkspaceItem } from '../types';

export const mockCurrentUser: UserProfile = {
  id: 'usr_01',
  name: 'John',
  email: 'john@mindora.app',
  role: 'user',
  avatarText: 'JD',
  workspaceName: 'Ruang Kerja Pribadi',
};

export const mockUserTasks: TaskItem[] = [
  {
    id: 'tsk_01',
    title: 'Merapikan meja belajar',
    category: 'Kotak Masuk',
    categoryColor: '#3B82F6',
    time: '08.20 WIB',
    completed: false,
    priority: 'medium',
  },
  {
    id: 'tsk_02',
    title: 'Persiapan diskusi tugas kelompok',
    category: 'Akademik',
    categoryColor: '#10B981',
    time: '10.20 WIB',
    completed: false,
    priority: 'high',
  },
  {
    id: 'tsk_03',
    title: 'Peregangan & yoga singkat',
    category: 'Kesehatan',
    categoryColor: '#EC4899',
    time: '14.30 WIB',
    completed: false,
    priority: 'low',
  },
  {
    id: 'tsk_04',
    title: 'Tinjau jurnal riset Machine Learning',
    category: 'Belajar',
    categoryColor: '#F59E0B',
    time: '16.15 WIB',
    completed: false,
    priority: 'high',
  },
  {
    id: 'tsk_05',
    title: 'Rencana jadwal liburan semester',
    category: 'Liburan',
    categoryColor: '#8B5CF6',
    time: '19.00 WIB',
    completed: true,
    priority: 'low',
  },
  {
    id: 'tsk_06',
    title: 'Pemeriksaan akhir laporan Proyek 3',
    category: 'Belajar',
    categoryColor: '#F59E0B',
    time: '21.00 WIB',
    completed: false,
    priority: 'medium',
  },
];

export const mockUpcomingSchedules: UpcomingSchedule[] = [
  {
    id: 'sch_01',
    title: 'Diskusi Tim Desain - Pembahasan kerangka proyek baru',
    timeRange: '11:30 - 13:00 WIB',
    type: 'meeting',
  },
  {
    id: 'sch_02',
    title: 'Pertemuan Dosen Pembimbing - Tinjau progres & konsultasi',
    timeRange: '15:00 - 17:00 WIB',
    type: 'meeting',
  },
];

export const mockMyLists: NavCategory[] = [
  { id: 'cat_inbox', name: 'Kotak Masuk', color: '#3B82F6', count: 4 },
  { id: 'cat_work', name: 'Akademik', color: '#10B981', count: 2 },
  { id: 'cat_study', name: 'Belajar', color: '#F59E0B', count: 5 },
  { id: 'cat_health', name: 'Kesehatan', color: '#EC4899', count: 1 },
  { id: 'cat_holiday', name: 'Liburan', color: '#8B5CF6', count: 1 },
];

export const mockWorkspaces: WorkspaceItem[] = [
  { id: 'ws_ui', name: 'Desain UI', iconType: 'palette' },
  { id: 'ws_ateam', name: 'Kelas Tim-A', iconType: 'hash' },
];
