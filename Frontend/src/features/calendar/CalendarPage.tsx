import { useState, useEffect, useMemo } from 'react';
import { DashboardTemplate } from '../../components/templates';
import { Button, ChevronLeftIcon, ChevronRightIcon, ClockIcon, PlusIcon, CloseIcon, CalendarIcon } from '../../components/atoms';
import { productivityService } from '../../services/productivityService';
import { mockUpcomingSchedules } from '../../mocks';
import type { UserProfile, TaskItem } from '../../types';
import { useAuth } from '../../hooks';
import { cn, translatePriority } from '../../utils';

interface CalendarEventItem {
  id: string;
  title: string;
  date: string; // YYYY-MM-DD
  timeRange: string;
  type: 'meeting' | 'study' | 'deadline' | 'lecture';
}

export default function CalendarPage() {
  const { user } = useAuth();

  const displayUser: UserProfile = user
    ? {
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        avatarText: user.name.slice(0, 2).toUpperCase(),
        workspaceName: 'Ruang Kerja Akademik',
      }
    : {
        id: 'usr_guest',
        name: 'Mahasiswa',
        email: 'student@mindora.app',
        role: 'user',
        avatarText: 'MH',
        workspaceName: 'Ruang Kerja Akademik',
      };

  // Current view date (year and month)
  const [currentDate, setCurrentDate] = useState(() => new Date());
  const [todayStr] = useState(() => new Date().toISOString().split('T')[0]);
  const [selectedDateStr, setSelectedDateStr] = useState(() => new Date().toISOString().split('T')[0]);
  const [filterType, setFilterType] = useState<string>('all');
  const [tasks, setTasks] = useState<TaskItem[]>([]);
  const [customEvents, setCustomEvents] = useState<CalendarEventItem[]>([]);
  const [isAddOpen, setIsAddOpen] = useState(false);

  // New Event Form
  const [newTitle, setNewTitle] = useState('');
  const [newDate, setNewDate] = useState(() => new Date().toISOString().split('T')[0]);
  const [newTime, setNewTime] = useState('10:00 - 11:30');
  const [newType, setNewType] = useState<'meeting' | 'study' | 'deadline' | 'lecture'>('study');

  // Load real tasks for deadline integration
  useEffect(() => {
    let isMounted = true;
    const fetchDeadlines = async () => {
      try {
        const res = await productivityService.listTasks({ limit: 100 });
        if (isMounted) {
          setTasks(res.tasks);
        }
      } catch (err) {
        console.error('Failed to load task deadlines for calendar:', err);
      }
    };
    fetchDeadlines();
    return () => {
      isMounted = false;
    };
  }, []);

  const daysOfWeek = ['Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab', 'Min'];

  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();

  const monthNames = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'
  ];

  // Base mock events combined with task deadlines and custom events
  const allEvents: CalendarEventItem[] = useMemo(() => {
    const baseEvents: CalendarEventItem[] = mockUpcomingSchedules.map((s, idx) => ({
      id: s.id,
      title: s.title,
      date: todayStr,
      timeRange: s.timeRange,
      type: (s.type === 'meeting' ? (idx % 2 === 0 ? 'lecture' : 'meeting') : 'study') as 'meeting' | 'study' | 'deadline' | 'lecture',
    }));

    // Add task deadlines as calendar items
    const taskDeadlines: CalendarEventItem[] = tasks
      .filter((t) => t.deadline)
      .map((t) => {
        const d = t.deadline ? t.deadline.split('T')[0] : todayStr;
        const timePart = t.deadline && t.deadline.includes('T') ? t.deadline.split('T')[1].slice(0, 5) : '23:59';
        return {
          id: `task_${t.id}`,
          title: `Tenggat: ${t.title} (${translatePriority(t.priority || 'normal')})`,
          date: d,
          timeRange: `Batas waktu ${timePart}`,
          type: 'deadline' as const,
        };
      });

    return [...baseEvents, ...taskDeadlines, ...customEvents];
  }, [tasks, customEvents, todayStr]);

  // Calendar matrix calculation
  const calendarGrid = useMemo(() => {
    const firstDayOfMonth = new Date(year, month, 1);
    const lastDayOfMonth = new Date(year, month + 1, 0);

    // Day of week: 0 (Sun) to 6 (Sat) -> Convert to 0 (Mon) to 6 (Sun)
    let startDayOfWeek = firstDayOfMonth.getDay() - 1;
    if (startDayOfWeek === -1) startDayOfWeek = 6;

    const daysInMonth = lastDayOfMonth.getDate();
    const daysInPrevMonth = new Date(year, month, 0).getDate();

    const days = [];

    // Previous month padding
    for (let i = startDayOfWeek - 1; i >= 0; i--) {
      const d = daysInPrevMonth - i;
      const prevDate = new Date(year, month - 1, d);
      const dStr = prevDate.toISOString().split('T')[0];
      days.push({
        dayNumber: d,
        dateStr: dStr,
        isCurrentMonth: false,
        hasEvents: allEvents.some((e) => e.date === dStr),
      });
    }

    // Current month days
    for (let d = 1; d <= daysInMonth; d++) {
      // Format as YYYY-MM-DD
      const mm = String(month + 1).padStart(2, '0');
      const dd = String(d).padStart(2, '0');
      const dStr = `${year}-${mm}-${dd}`;
      days.push({
        dayNumber: d,
        dateStr: dStr,
        isCurrentMonth: true,
        hasEvents: allEvents.some((e) => e.date === dStr),
      });
    }

    // Next month padding to complete 35 or 42 cells
    const remaining = 7 - (days.length % 7);
    if (remaining < 7) {
      for (let d = 1; d <= remaining; d++) {
        const nextDate = new Date(year, month + 1, d);
        const dStr = nextDate.toISOString().split('T')[0];
        days.push({
          dayNumber: d,
          dateStr: dStr,
          isCurrentMonth: false,
          hasEvents: allEvents.some((e) => e.date === dStr),
        });
      }
    }

    return days;
  }, [year, month, allEvents]);

  const handlePrevMonth = () => {
    setCurrentDate(new Date(year, month - 1, 1));
  };

  const handleNextMonth = () => {
    setCurrentDate(new Date(year, month + 1, 1));
  };

  const handleToday = () => {
    const today = new Date();
    setCurrentDate(today);
    setSelectedDateStr(today.toISOString().split('T')[0]);
  };

  const handleAddEvent = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newTitle.trim()) return;

    const event: CalendarEventItem = {
      id: `evt_${Date.now()}`,
      title: newTitle.trim(),
      date: newDate,
      timeRange: newTime.trim() || 'Sepanjang Hari',
      type: newType,
    };

    setCustomEvents((prev) => [...prev, event]);
    setNewTitle('');
    setIsAddOpen(false);
  };

  const getEventTypeLabel = (type: string) => {
    switch (type) {
      case 'deadline':
        return 'Tenggat';
      case 'lecture':
        return 'Kuliah';
      case 'study':
        return 'Belajar';
      case 'meeting':
        return 'Pertemuan';
      default:
        return type;
    }
  };

  const selectedDayEvents = allEvents.filter((e) => {
    const dateMatch = e.date === selectedDateStr;
    const typeMatch = filterType === 'all' || e.type === filterType;
    return dateMatch && typeMatch;
  });

  return (
    <DashboardTemplate role="user" currentUser={displayUser}>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-[#123C32]">
              Kalender & Jadwal Akademik
            </h1>
            <p className="text-xs text-text-secondary mt-1">
              Koordinasikan jadwal kuliah, tenggat pengumpulan tugas, dan sesi belajar terstruktur.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={handleToday}
              className="text-xs font-semibold"
            >
              Hari Ini
            </Button>
            <Button
              variant="primary"
              size="sm"
              onClick={() => {
                setNewDate(selectedDateStr);
                setIsAddOpen(true);
              }}
              className="text-xs font-bold gap-1.5 shadow-soft-xs"
            >
              <PlusIcon className="w-4 h-4" />
              <span>Tambah Agenda</span>
            </Button>
          </div>
        </div>

        {/* 2-Column Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          {/* Calendar Grid (8 cols) */}
          <div className="lg:col-span-8 bg-white rounded-3xl border border-border p-6 shadow-soft-xs space-y-4">
            {/* Month Navigation Bar */}
            <div className="flex items-center justify-between pb-3 border-b border-border">
              <div className="flex items-center gap-3">
                <h2 className="text-lg font-bold text-[#123C32]">
                  {monthNames[month]} {year}
                </h2>
              </div>
              <div className="flex items-center gap-1.5">
                <button
                  type="button"
                  onClick={handlePrevMonth}
                  className="p-1.5 rounded-xl border border-border hover:bg-surface-muted text-text-secondary transition-colors"
                  aria-label="Bulan sebelumnya"
                >
                  <ChevronLeftIcon className="w-4 h-4" />
                </button>
                <button
                  type="button"
                  onClick={handleNextMonth}
                  className="p-1.5 rounded-xl border border-border hover:bg-surface-muted text-text-secondary transition-colors"
                  aria-label="Bulan berikutnya"
                >
                  <ChevronRightIcon className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Days of week header */}
            <div className="grid grid-cols-7 gap-1 text-center font-bold text-xs text-[#526B60] py-1">
              {daysOfWeek.map((day) => (
                <div key={day} className="py-1">
                  {day}
                </div>
              ))}
            </div>

            {/* Days Matrix */}
            <div className="grid grid-cols-7 gap-1.5">
              {calendarGrid.map((item, idx) => {
                const isSelected = item.dateStr === selectedDateStr;
                const isToday = item.dateStr === todayStr;
                return (
                  <button
                    key={`${item.dateStr}_${idx}`}
                    type="button"
                    onClick={() => setSelectedDateStr(item.dateStr)}
                    className={cn(
                      'h-14 sm:h-16 p-1.5 rounded-2xl flex flex-col justify-between items-start transition-all text-left relative cursor-pointer border',
                      item.isCurrentMonth
                        ? 'bg-white hover:border-emerald-300'
                        : 'bg-slate-50/50 text-slate-400 opacity-60 border-transparent',
                      isSelected
                        ? 'border-[#138F78] ring-2 ring-[#138F78]/20 bg-emerald-50/40 shadow-soft-xs'
                        : 'border-border/60',
                      isToday && !isSelected && 'border-amber-300 bg-amber-50/30'
                    )}
                  >
                    <span
                      className={cn(
                        'text-xs font-semibold rounded-md px-1.5 py-0.5',
                        isSelected ? 'bg-[#138F78] text-white' : isToday ? 'text-amber-800 font-bold' : 'text-text-primary'
                      )}
                    >
                      {item.dayNumber}
                    </span>

                    {item.hasEvents && (
                      <div className="flex items-center gap-1 w-full pl-0.5 pb-0.5">
                        <span className="w-2 h-2 rounded-full bg-[#138F78]" />
                        <span className="text-[10px] text-text-secondary hidden sm:inline truncate">
                          Agenda
                        </span>
                      </div>
                    )}
                  </button>
                );
              })}
            </div>
          </div>

          {/* Agenda & Selected Day Detail (4 cols) */}
          <div className="lg:col-span-4 space-y-6">
            <div className="bg-white rounded-3xl border border-border p-6 shadow-soft-xs space-y-4">
              <div className="flex items-center justify-between pb-2 border-b border-border">
                <div>
                  <h3 className="text-sm font-bold text-[#123C32]">Jadwal Tanggal Ini</h3>
                  <span className="text-xs text-text-secondary">{selectedDateStr}</span>
                </div>
                <span className="text-[11px] font-bold text-[#138F78] bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200">
                  {selectedDayEvents.length} agenda
                </span>
              </div>

              {/* Filter Pills */}
              <div className="flex items-center gap-1.5 flex-wrap">
                {[
                  { label: 'Semua', value: 'all' },
                  { label: 'Kuliah', value: 'lecture' },
                  { label: 'Tenggat', value: 'deadline' },
                  { label: 'Belajar', value: 'study' },
                ].map((pill) => (
                  <button
                    key={pill.value}
                    type="button"
                    onClick={() => setFilterType(pill.value)}
                    className={cn(
                      'px-2.5 py-1 rounded-lg text-[11px] font-semibold transition-all',
                      filterType === pill.value
                        ? 'bg-[#123C32] text-white shadow-soft-xs'
                        : 'bg-surface-muted text-text-secondary hover:text-text-primary'
                    )}
                  >
                    {pill.label}
                  </button>
                ))}
              </div>

              {/* Event List */}
              {selectedDayEvents.length === 0 ? (
                <div className="py-8 text-center space-y-2 bg-surface-muted/30 rounded-2xl border border-dashed border-border">
                  <div className="w-10 h-10 rounded-xl bg-emerald-50 text-primary flex items-center justify-center mx-auto mb-1"><CalendarIcon className="w-5 h-5" /></div>
                  <p className="text-xs font-semibold text-text-primary">Tidak ada agenda untuk tanggal ini</p>
                  <p className="text-[11px] text-text-secondary max-w-[220px] mx-auto">
                    Klik 'Tambah Agenda' di atas atau buat tugas dengan tenggat pada tanggal ini.
                  </p>
                </div>
              ) : (
                <div className="space-y-2.5">
                  {selectedDayEvents.map((evt) => (
                    <div
                      key={evt.id}
                      className={cn(
                        'p-3.5 rounded-2xl border transition-all space-y-1.5',
                        evt.type === 'deadline'
                          ? 'bg-rose-50/60 border-rose-200 text-rose-950'
                          : evt.type === 'lecture'
                          ? 'bg-blue-50/60 border-blue-200 text-blue-950'
                          : 'bg-emerald-50/60 border-emerald-200 text-emerald-950'
                      )}
                    >
                      <div className="flex items-center justify-between gap-2">
                        <span
                          className={cn(
                            'text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full',
                            evt.type === 'deadline'
                              ? 'bg-rose-200 text-rose-900'
                              : evt.type === 'lecture'
                              ? 'bg-blue-200 text-blue-900'
                              : 'bg-emerald-200 text-emerald-900'
                          )}
                        >
                          {getEventTypeLabel(evt.type)}
                        </span>
                        <div className="flex items-center gap-1 text-[11px] font-medium text-text-secondary">
                          <ClockIcon className="w-3.5 h-3.5" />
                          <span>{evt.timeRange}</span>
                        </div>
                      </div>
                      <h4 className="text-xs font-bold leading-snug">{evt.title}</h4>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {/* Quick Helper Card */}
            <div className="p-5 rounded-3xl bg-gradient-to-br from-emerald-50 to-teal-50/40 border border-emerald-200/60 space-y-2">
              <span className="text-xs font-bold text-[#123C32] block">Tips Akademik</span>
              <p className="text-[11px] text-[#526B60] leading-relaxed">
                Sinkronkan target belajar dengan tenggat tugas. Sesi belajar yang diselesaikan di Mindora akan terintegrasi langsung ke Analisis Produktivitas Anda.
              </p>
            </div>
          </div>
        </div>

        {/* Add Event Modal */}
        {isAddOpen && (
          <div className="fixed inset-0 z-50 bg-slate-900/40 backdrop-blur-xs flex items-center justify-center p-4">
            <div className="bg-white rounded-3xl border border-border max-w-md w-full p-6 shadow-soft-lg space-y-4 animate-scaleUp">
              <div className="flex items-center justify-between">
                <h3 className="text-base font-bold text-[#123C32]">Jadwalkan Agenda Akademik</h3>
                <button
                  type="button"
                  onClick={() => setIsAddOpen(false)}
                  className="p-1 rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-muted"
                >
                  <CloseIcon className="w-5 h-5" />
                </button>
              </div>

              <form onSubmit={handleAddEvent} className="space-y-3.5">
                <div>
                  <label className="text-xs font-bold text-text-primary block mb-1">
                    Judul Agenda *
                  </label>
                  <input
                    type="text"
                    required
                    placeholder="Contoh: Kuliah Sistem Terdistribusi, Ujian Matematika"
                    value={newTitle}
                    onChange={(e) => setNewTitle(e.target.value)}
                    className="w-full h-10 px-3 rounded-xl border border-border text-xs focus:ring-2 focus:ring-[#138F78]/20 focus:outline-none"
                  />
                </div>

                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="text-xs font-bold text-text-primary block mb-1">
                      Tanggal Agenda *
                    </label>
                    <input
                      type="date"
                      required
                      value={newDate}
                      onChange={(e) => setNewDate(e.target.value)}
                      className="w-full h-10 px-3 rounded-xl border border-border text-xs focus:ring-2 focus:ring-[#138F78]/20 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="text-xs font-bold text-text-primary block mb-1">
                      Kategori
                    </label>
                    <select
                      value={newType}
                      onChange={(e) => setNewType(e.target.value as 'meeting' | 'study' | 'deadline' | 'lecture')}
                      className="w-full h-10 px-3 rounded-xl border border-border text-xs focus:ring-2 focus:ring-[#138F78]/20 focus:outline-none"
                    >
                      <option value="lecture">Kuliah / Kelas</option>
                      <option value="study">Sesi Belajar</option>
                      <option value="deadline">Tenggat Tugas</option>
                      <option value="meeting">Pertemuan / Diskusi</option>
                    </select>
                  </div>
                </div>

                <div>
                  <label className="text-xs font-bold text-text-primary block mb-1">
                    Rentang Waktu
                  </label>
                  <input
                    type="text"
                    placeholder="Contoh: 14:00 - 16:00"
                    value={newTime}
                    onChange={(e) => setNewTime(e.target.value)}
                    className="w-full h-10 px-3 rounded-xl border border-border text-xs focus:ring-2 focus:ring-[#138F78]/20 focus:outline-none"
                  />
                </div>

                <div className="flex items-center justify-end gap-2 pt-2">
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => setIsAddOpen(false)}
                    className="text-xs"
                  >
                    Batal
                  </Button>
                  <Button
                    type="submit"
                    variant="primary"
                    size="sm"
                    className="text-xs font-bold shadow-soft-xs"
                  >
                    Simpan Agenda
                  </Button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </DashboardTemplate>
  );
}

