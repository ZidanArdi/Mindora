import { useState, useEffect } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { DashboardTemplate } from '../../components/templates';
import { UserHeroBanner } from './UserHeroBanner';
import { AIAssistantWidget } from './AIAssistantWidget';
import { FocusTimerWidget } from '../focus/FocusTimerWidget';
import { CalendarWidget } from '../../components/molecules';
import { ShieldIcon, Button, CheckIcon } from '../../components/atoms';
import { productivityService } from '../../services/productivityService';
import type { DashboardSummary, UserProfile, TaskStatus, ProductivityPrediction } from '../../types';
import { useAuth } from '../../hooks';
import { translatePriority, translateCategory, formatDate } from '../../utils/i18n';

export default function UserDashboardPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { user } = useAuth();

  // Redirect legacy hash anchors to dedicated routes so the dashboard is not stuck
  useEffect(() => {
    if (location.hash === '#list') {
      navigate('/tasks', { replace: true });
    } else if (location.hash === '#calendar') {
      navigate('/calendar', { replace: true });
    } else if (location.hash === '#messages') {
      navigate('/messages', { replace: true });
    }
  }, [location.hash, navigate]);

  const isAdminViewingUserWorkspace = user?.role === 'admin';
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [latestPrediction, setLatestPrediction] = useState<ProductivityPrediction | null>(null);
  const [loading, setLoading] = useState(() => !isAdminViewingUserWorkspace);

  const displayUser: UserProfile = user
    ? {
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        avatarText:
          user.name
            .split(' ')
            .map((n) => n[0])
            .join('')
            .slice(0, 2)
            .toUpperCase() || 'MH',
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

  useEffect(() => {
    let isMounted = true;
    const fetchSummary = async () => {
      try {
        setLoading(true);
        const [data, predRes] = await Promise.all([
          productivityService.getDashboardSummary(),
          productivityService.listPredictions({ page: 1, limit: 1 }).catch(() => null),
        ]);
        if (isMounted) {
          setSummary(data);
          if (predRes && predRes.items && predRes.items.length > 0) {
            setLatestPrediction(predRes.items[0]);
          }
        }
      } catch (err: unknown) {
        console.error('Failed to load dashboard summary:', err);
      } finally {
        if (isMounted) setLoading(false);
      }
    };

    if (!isAdminViewingUserWorkspace) {
      fetchSummary();
    }

    return () => {
      isMounted = false;
    };
  }, [isAdminViewingUserWorkspace]);

  const handleToggleTask = async (taskId: string, currentStatus: string) => {
    try {
      const nextStatus: TaskStatus = currentStatus === 'completed' ? 'pending' : 'completed';
      await productivityService.updateTaskStatus(taskId, nextStatus);
      // Refresh summary
      const updated = await productivityService.getDashboardSummary();
      setSummary(updated);
    } catch (err: unknown) {
      console.error('Failed to update task:', err);
    }
  };

  const formatMinutes = (seconds: number) => {
    const mins = Math.round(seconds / 60);
    if (mins >= 60) {
      const h = Math.floor(mins / 60);
      const m = mins % 60;
      return `${h} jam ${m} mnt`;
    }
    return `${mins} mnt`;
  };

  return (
    <DashboardTemplate role="user" currentUser={displayUser}>
      {isAdminViewingUserWorkspace ? (
        <div className="p-8 rounded-2xl bg-surface-card border border-border space-y-4 text-center max-w-2xl mx-auto shadow-soft-sm my-12 animate-fadeIn">
          <div className="w-12 h-12 rounded-xl bg-nature-mint text-primary flex items-center justify-center mx-auto">
            <ShieldIcon className="w-6 h-6" />
          </div>
          <h2 className="text-xl font-bold text-text-primary">
            Penegakan Kebijakan Privasi Mahasiswa
          </h2>
          <p className="text-sm text-text-secondary leading-relaxed">
            Sesuai dengan RBAC Mindora dan aturan privasi data mahasiswa, catatan produktivitas pribadi, tugas belajar, dan sesi timer fokus terisolasi secara ketat dan tidak dapat diakses dari akun administrator.
          </p>
          <div className="pt-2 flex justify-center">
            <Button
              variant="primary"
              size="md"
              onClick={() => navigate('/admin')}
              className="text-xs font-bold"
            >
              Buka Konsol Admin
            </Button>
          </div>
        </div>
      ) : (
        <div className="space-y-6">
          {/* Top Hero Banner */}
          <UserHeroBanner
            userName={displayUser.name}
            taskCount={summary ? summary.pending_tasks : 0}
          />

          {/* Quick Metrics Bar (Real Data, NO Fake ML Score) */}
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Task Completion Rate */}
            <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
                  Progres Tugas
                </span>
                <Link to="/tasks" className="text-[11px] font-bold text-primary hover:underline">
                  Lihat â†’
                </Link>
              </div>
              <div className="my-2">
                <div className="text-2xl font-black text-[#123C32]">
                  {summary ? `${Math.round(summary.completion_percentage)}%` : '0%'}
                </div>
                <div className="w-full h-1.5 rounded-full bg-slate-100 mt-2 overflow-hidden">
                  <div
                    className="h-full bg-[#138F78] transition-all duration-500 rounded-full"
                    style={{
                      width: `${summary ? Math.min(100, summary.completion_percentage) : 0}%`,
                    }}
                  />
                </div>
              </div>
              <span className="text-[11px] text-[#526B60]">
                {summary
                  ? `${summary.completed_tasks} dari ${summary.total_tasks} tugas selesai`
                  : '0 tugas dibuat'}
              </span>
            </div>

            {/* Study Time Today */}
            <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
                  Belajar Hari Ini
                </span>
                <Link to="/study" className="text-[11px] font-bold text-primary hover:underline">
                  Ruang Belajar â†’
                </Link>
              </div>
              <div className="my-2 text-2xl font-black text-[#123C32]">
                {summary ? formatMinutes(summary.total_study_seconds) : '0 mnt'}
              </div>
              <span className="text-[11px] text-[#526B60]">
                {summary
                  ? `${summary.completed_study_sessions} sesi belajar selesai`
                  : 'Tidak ada sesi aktif'}
              </span>
            </div>

            {/* Focus Intervals */}
            <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
                  Blok Fokus Mendalam
                </span>
                <Link to="/focus" className="text-[11px] font-bold text-primary hover:underline">
                  Interval â†’
                </Link>
              </div>
              <div className="my-2 text-2xl font-black text-[#138F78]">
                {summary ? summary.completed_focus_count : 0}
              </div>
              <span className="text-[11px] text-[#526B60]">
                {summary
                  ? `${summary.completed_breaks_count} istirahat Â· ${summary.total_interruptions} distraksi`
                  : '0 interval'}
              </span>
            </div>

            {/* Lifestyle Habit Status */}
            <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
                  Kalibrasi Kebiasaan
                </span>
                <Link to="/lifestyle" className="text-[11px] font-bold text-primary hover:underline">
                  Kalibrasi â†’
                </Link>
              </div>
              <div className="my-2">
                {summary?.today_lifestyle_logged ? (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-emerald-100 text-[#123C32] text-xs font-bold border border-emerald-300">
                    <span className="w-2 h-2 rounded-full bg-[#138F78]" />
                    Terkalibrasi Hari Ini
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-amber-100 text-amber-900 text-xs font-bold border border-amber-300">
                    <span className="w-2 h-2 rounded-full bg-amber-500 animate-pulse" />
                    Menunggu Input
                  </span>
                )}
              </div>
              <span className="text-[11px] text-[#526B60]">
                {summary?.today_lifestyle_logged
                  ? 'Semua 8 variabel tercatat'
                  : 'Catat tidur & waktu layar hari ini'}
              </span>
            </div>
          </div>

          {/* Main 2-Column Grid */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
            {/* Left Column (8 cols): Recent Tasks & AI Assistant */}
            <div className="lg:col-span-8 space-y-6">
              {/* AI Assistant Quick Tool */}
              <AIAssistantWidget />

              {/* Real Tasks Card */}
              <div className="bg-white rounded-3xl border border-border p-6 shadow-soft-xs space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h2 className="text-base font-bold text-[#123C32]">Tugas Belajar Aktif</h2>
                    <p className="text-xs text-[#526B60]">
                      Target akademik dan tugas kuliah Anda yang akan datang.
                    </p>
                  </div>
                  <Link
                    to="/tasks"
                    className="px-3.5 py-1.5 rounded-xl bg-emerald-50 hover:bg-emerald-100 text-[#123C32] text-xs font-bold border border-emerald-200 transition-colors"
                  >
                    Kelola Semua Tugas â†’
                  </Link>
                </div>

                {loading ? (
                  <div className="py-8 text-center text-xs text-text-muted">Memuat tugas...</div>
                ) : !summary || summary.recent_tasks.length === 0 ? (
                  <div className="py-10 text-center space-y-3 bg-surface-muted/30 rounded-2xl border border-dashed border-border">
                    <span className="text-3xl block">ðŸ“</span>
                    <p className="text-sm font-semibold text-text-primary">
                      Belum ada tugas dibuat
                    </p>
                    <p className="text-xs text-text-secondary max-w-sm mx-auto">
                      Atur PR, materi kuliah, dan tenggat waktu proyek Anda dengan pengelola tugas kami.
                    </p>
                    <div className="pt-2">
                      <Link
                        to="/tasks"
                        className="inline-flex px-4 py-2 rounded-xl text-xs font-bold text-white bg-[#138F78] hover:bg-[#123C32] transition-colors shadow-soft-xs"
                      >
                        + Buat Tugas Pertama Anda
                      </Link>
                    </div>
                  </div>
                ) : (
                  <div className="space-y-2">
                    {summary.recent_tasks.map((task) => (
                      <div
                        key={task.id}
                        className="flex items-center justify-between p-3.5 rounded-2xl bg-surface-muted/40 hover:bg-surface-muted border border-border/60 transition-all group"
                      >
                        <div className="flex items-center gap-3 min-w-0">
                          <button
                            type="button"
                            onClick={() => handleToggleTask(task.id, task.status || 'pending')}
                            className={`w-5 h-5 rounded-lg border flex items-center justify-center transition-all cursor-pointer ${
                              task.status === 'completed'
                                ? 'bg-[#138F78] border-[#138F78] text-white'
                                : 'border-slate-300 hover:border-[#138F78] bg-white'
                            }`}
                          >
                            {task.status === 'completed' && <CheckIcon className="w-3.5 h-3.5" />}
                          </button>
                          <div className="truncate">
                            <span
                              className={`text-xs font-semibold block truncate ${
                                task.status === 'completed'
                                    ? 'line-through text-text-muted'
                                    : 'text-text-primary group-hover:text-primary transition-colors'
                              }`}
                            >
                              {task.title}
                            </span>
                            <div className="flex items-center gap-2 mt-0.5">
                              <span className="text-[10px] font-medium text-text-secondary">
                                {translateCategory(task.category)}
                              </span>
                              {task.deadline && (
                                <span className="text-[10px] font-medium text-amber-700">
                                  Tenggat: {formatDate(task.deadline, { month: 'short', day: 'numeric' })}
                                </span>
                              )}
                            </div>
                          </div>
                        </div>

                        <div className="flex items-center gap-2 shrink-0">
                          <span
                            className={`px-2 py-0.5 rounded text-[10px] font-bold uppercase ${
                              task.priority === 'urgent'
                                ? 'bg-rose-100 text-rose-800'
                                : task.priority === 'high'
                                ? 'bg-amber-100 text-amber-800'
                                : 'bg-slate-100 text-slate-700'
                            }`}
                          >
                            {translatePriority(task.priority)}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>

            {/* Right Column (4 cols): Prediction, Calendar & Focus Quickstart */}
            <div className="lg:col-span-4 space-y-6">
              {/* ML Productivity Estimate Quick Card */}
              <div className="bg-gradient-to-br from-[#123C32] to-[#0A261F] text-white rounded-3xl p-5 shadow-soft-xs space-y-3">
                <div className="flex items-center justify-between">
                  <span
                    className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${
                      latestPrediction?.is_baseline || latestPrediction?.prediction_state === 'baseline'
                        ? 'bg-amber-400/20 text-amber-300'
                        : 'bg-white/10 text-emerald-300'
                    }`}
                  >
                    {latestPrediction?.is_baseline || latestPrediction?.prediction_state === 'baseline'
                      ? 'Estimasi Awal'
                      : (latestPrediction?.state_label || 'Estimasi Terpersonalisasi')}
                  </span>
                  <span className="text-[10px] font-mono text-white/60">Ridge v1.0.0</span>
                </div>
                <div className="flex items-baseline justify-between pt-1">
                  <div>
                    <div className="text-3xl font-black text-emerald-400">
                      {latestPrediction ? Math.round(latestPrediction.predicted_score) : '--'}
                      <span className="text-sm font-normal text-white/50">/100</span>
                    </div>
                    <span className="text-[11px] text-white/70 block mt-0.5">
                      {latestPrediction ? `Untuk ${latestPrediction.prediction_date}` : 'Belum ada estimasi hari ini'}
                    </span>
                  </div>
                  <Link
                    to="/analytics"
                    className="text-xs font-bold text-emerald-300 hover:text-white transition-colors"
                  >
                    Analitik Lengkap â†’
                  </Link>
                </div>
                <p className="text-[10px] text-white/60 border-t border-white/10 pt-2 leading-relaxed">
                  {latestPrediction?.state_description ||
                    'Diestimasi dari pola kebiasaan (belajar, tidur, ponsel, stres). Bukan merupakan nilai akademik objektif.'}
                </p>
              </div>

              <CalendarWidget />
              <FocusTimerWidget />
            </div>
          </div>
        </div>
      )}
    </DashboardTemplate>
  );
}

