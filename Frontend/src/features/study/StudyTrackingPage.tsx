import { useState, useEffect, useRef } from 'react';
import { DashboardTemplate } from '../../components/templates';
import { Button, Input, ClockIcon, BookOpenIcon } from '../../components/atoms';
import { productivityService } from '../../services/productivityService';
import type { StudySession, StudyDailyStats, TaskItem } from '../../types';
import { useAuth } from '../../hooks';
import { formatDateTime, translateStatus, formatStudyDuration, formatTimer } from '../../utils/i18n';

export default function StudyTrackingPage() {
  const { user } = useAuth();
  const [activeSession, setActiveSession] = useState<StudySession | null>(null);
  const [history, setHistory] = useState<StudySession[]>([]);
  const [dailyStats, setDailyStats] = useState<StudyDailyStats>({
    total_completed_seconds: 0,
    completed_sessions_count: 0,
  });
  const [availableTasks, setAvailableTasks] = useState<TaskItem[]>([]);
  const [loading, setLoading] = useState(true);

  // Form State
  const [subject, setSubject] = useState('');
  const [selectedTaskID, setSelectedTaskID] = useState<string>('');
  const [submitting, setSubmitting] = useState(false);

  // Live Elapsed Timer State (client counter synchronized with server start)
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const timerRef = useRef<number | null>(null);

  const applyStudyData = (
    active: StudySession | null,
    hist: { sessions: StudySession[] },
    stats: StudyDailyStats | null,
    taskData: { tasks: TaskItem[] }
  ) => {
    setActiveSession(active);
    setHistory(hist.sessions);
    if (stats) setDailyStats(stats);
    setAvailableTasks(taskData.tasks);

    if (active) {
      if (active.status === 'active') {
        const startedAtMs = new Date(active.started_at).getTime();
        const initialElapsed = Math.max(
          active.duration_seconds,
          Math.floor((Date.now() - startedAtMs) / 1000)
        );
        setElapsedSeconds(initialElapsed);
      } else {
        setElapsedSeconds(active.duration_seconds);
      }
    } else {
      setElapsedSeconds(0);
    }
  };

  const loadData = async () => {
    try {
      setLoading(true);
      const [active, hist, stats, taskData] = await Promise.all([
        productivityService.getActiveStudySession(),
        productivityService.listStudySessions({ limit: 10 }),
        productivityService.getDailyStudyStats(),
        productivityService.listTasks({ status: 'pending', limit: 50 }),
      ]);
      applyStudyData(active, hist, stats, taskData);
    } catch (err: unknown) {
      console.error('Failed to load study data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    let isMounted = true;
    Promise.all([
      productivityService.getActiveStudySession(),
      productivityService.listStudySessions({ limit: 10 }),
      productivityService.getDailyStudyStats(),
      productivityService.listTasks({ status: 'pending', limit: 50 }),
    ])
      .then(([active, hist, stats, taskData]) => {
        if (isMounted) {
          applyStudyData(active, hist, stats, taskData);
        }
      })
      .catch((err: unknown) => {
        console.error('Failed to load study data:', err);
      })
      .finally(() => {
        if (isMounted) {
          setLoading(false);
        }
      });

    return () => {
      isMounted = false;
    };
  }, []);

  // Timer ticker
  useEffect(() => {
    if (activeSession && activeSession.status === 'active') {
      timerRef.current = window.setInterval(() => {
        setElapsedSeconds((prev) => prev + 1);
      }, 1000);
    } else {
      if (timerRef.current) clearInterval(timerRef.current);
    }
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [activeSession]);

  const handleStart = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!subject.trim()) return;

    try {
      setSubmitting(true);
      const session = await productivityService.startStudySession({
        subject: subject.trim(),
        task_id: selectedTaskID || undefined,
      });
      setActiveSession(session);
      setSubject('');
      setSelectedTaskID('');
      setElapsedSeconds(0);
      loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Failed to start session');
    } finally {
      setSubmitting(false);
    }
  };

  const handlePause = async () => {
    if (!activeSession) return;
    try {
      const updated = await productivityService.pauseStudySession(activeSession.id);
      setActiveSession(updated);
      setElapsedSeconds(updated.duration_seconds);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Pause failed');
    }
  };

  const handleResume = async () => {
    if (!activeSession) return;
    try {
      const updated = await productivityService.resumeStudySession(activeSession.id);
      setActiveSession(updated);
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Resume failed');
    }
  };

  const handleFinish = async () => {
    if (!activeSession) return;
    try {
      await productivityService.finishStudySession(activeSession.id);
      setActiveSession(null);
      setElapsedSeconds(0);
      loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Finish failed');
    }
  };

  const handleCancel = async () => {
    if (!activeSession) return;
    if (!confirm('Apakah Anda yakin ingin membatalkan sesi ini? Durasi yang belum selesai tidak akan disimpan.')) return;
    try {
      await productivityService.cancelStudySession(activeSession.id);
      setActiveSession(null);
      setElapsedSeconds(0);
      loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Cancel failed');
    }
  };

  const currentUserProfile = user
    ? {
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        avatarText: user.name.slice(0, 2).toUpperCase(),
      }
    : { id: '', name: 'Mahasiswa', email: '', role: 'user' as const, avatarText: 'MH' };

  return (
    <DashboardTemplate role="user" currentUser={currentUserProfile}>
      <div className="space-y-6 max-w-7xl mx-auto">
        {/* Header */}
        <div className="pb-2 border-b border-border">
          <h1 className="text-2xl font-bold text-text-primary tracking-tight">
            Pelacakan Belajar
          </h1>
          <p className="text-xs text-text-secondary mt-1">
            Catat sesi belajar akademik berkelanjutan, hubungkan sesi dengan tugas, dan ukur total jam belajar harian Anda.
          </p>
        </div>

        {/* Daily Stats KPI Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-emerald-100 text-emerald-800 flex items-center justify-center">
              <ClockIcon className="w-6 h-6 text-emerald-800" />
            </div>
            <div>
              <p className="text-xs font-semibold text-text-secondary">Jam Belajar Hari Ini</p>
              <h3 className="text-2xl font-black text-text-primary mt-0.5">
                {(dailyStats.total_completed_seconds / 3600).toFixed(1)} jam
              </h3>
              <p className="text-[10px] text-text-muted mt-0.5">
                {dailyStats.total_completed_seconds === 0
                  ? '0 menit terselesaikan tercatat'
                  : dailyStats.total_completed_seconds < 60
                  ? `${dailyStats.total_completed_seconds} detik terselesaikan tercatat`
                  : `${Math.floor(dailyStats.total_completed_seconds / 60)} menit terselesaikan tercatat`}
              </p>
            </div>
          </div>

          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-primary-100 text-primary flex items-center justify-center">
              <BookOpenIcon className="w-6 h-6 text-primary" />
            </div>
            <div>
              <p className="text-xs font-semibold text-text-secondary">Sesi Selesai</p>
              <h3 className="text-2xl font-black text-text-primary mt-0.5">
                {dailyStats.completed_sessions_count}
              </h3>
              <p className="text-[10px] text-text-muted mt-0.5">Sesi selesai yang berkontribusi pada total harian</p>
            </div>
          </div>
        </div>

        {/* Active Session OR Start New Session Card */}
        {activeSession ? (
          <div className="p-8 rounded-3xl bg-gradient-to-br from-[#123C32] to-[#0E2721] text-white shadow-soft-lg space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div>
                <span className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-500/20 text-emerald-300 text-xs font-bold border border-emerald-400/30">
                  <span className={`w-2 h-2 rounded-full ${activeSession.status === 'active' ? 'bg-emerald-400 animate-ping' : 'bg-amber-400'}`} />
                  {activeSession.status === 'active' ? 'SESI SEDANG BERJALAN' : 'SESI DIJEDA'}
                </span>
                <h2 className="text-2xl font-bold tracking-tight mt-2">{activeSession.subject}</h2>
                {activeSession.task && (
                  <p className="text-xs text-emerald-200 mt-1">
                    Terhubung dengan tugas: <strong className="underline">{activeSession.task.title}</strong>
                  </p>
                )}
              </div>

              {/* Running Time Display */}
              <div className="text-center sm:text-right">
                <span className="text-4xl sm:text-5xl font-mono font-bold tracking-widest text-emerald-300">
                  {formatTimer(elapsedSeconds)}
                </span>
                <p className="text-[10px] text-emerald-200/80 mt-1 uppercase tracking-wider">
                  Waktu Berjalan Tersinkronisasi Server
                </p>
              </div>
            </div>

            {/* Action Buttons */}
            <div className="flex flex-wrap items-center gap-3 pt-4 border-t border-emerald-700/50">
              {activeSession.status === 'active' ? (
                <Button
                  variant="outline"
                  size="md"
                  onClick={handlePause}
                  className="bg-white/10 hover:bg-white/20 text-white border-white/20 text-xs font-semibold"
                >
                  Jeda Sesi
                </Button>
              ) : (
                <Button
                  variant="primary"
                  size="md"
                  onClick={handleResume}
                  className="!bg-emerald-500 hover:!bg-emerald-600 text-white text-xs font-semibold"
                >
                  Lanjutkan Sesi
                </Button>
              )}

              <Button
                variant="primary"
                size="md"
                onClick={handleFinish}
                className="!bg-[#138F78] hover:!bg-[#0e6f5d] text-white text-xs font-bold"
              >
                Selesaikan & Catat Sesi
              </Button>

              <button
                type="button"
                onClick={handleCancel}
                className="px-4 py-2 rounded-xl text-xs font-semibold text-rose-300 hover:text-rose-100 hover:bg-rose-900/30 transition-colors ml-auto"
              >
                Batalkan Sesi
              </button>
            </div>
          </div>
        ) : (
          <div className="p-6 rounded-3xl bg-white border border-border shadow-soft-xs space-y-4">
            <h2 className="text-base font-bold text-text-primary">Mulai Sesi Belajar Baru</h2>
            <form onSubmit={handleStart} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold text-text-primary mb-1">
                    Mata Kuliah / Topik Belajar *
                  </label>
                  <Input
                    required
                    placeholder="cth. Kuliah Algoritma Terdistribusi Pertemuan 5"
                    value={subject}
                    onChange={(e) => setSubject(e.target.value)}
                    className="w-full text-xs"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-text-primary mb-1">
                    Hubungkan dengan Tugas (Opsional)
                  </label>
                  <select
                    value={selectedTaskID}
                    onChange={(e) => setSelectedTaskID(e.target.value)}
                    className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    <option value="">Tanpa tautan tugas (Belajar mandiri)</option>
                    {availableTasks.map((t) => (
                      <option key={t.id} value={t.id}>
                        [{t.category}] {t.title}
                      </option>
                    ))}
                  </select>
                </div>
              </div>

              <div className="flex justify-end pt-2">
                <Button
                  type="submit"
                  variant="primary"
                  size="md"
                  disabled={submitting || !subject.trim()}
                  className="text-xs font-bold gap-2"
                >
                  <span>Mulai Sesi Belajar</span>
                </Button>
              </div>
            </form>
          </div>
        )}

        {/* Study History Table */}
        <div className="bg-white rounded-3xl border border-border shadow-soft-xs p-6 space-y-4">
          <h2 className="text-base font-bold text-text-primary">Riwayat Sesi Belajar Terkini</h2>

          {loading ? (
            <p className="text-xs text-text-secondary animate-pulse">Memuat sesi belajar...</p>
          ) : history.length === 0 ? (
            <div className="p-8 text-center text-xs text-text-muted border-2 border-dashed border-border rounded-2xl">
              Belum ada sesi belajar tercatat. Mulai sesi di atas untuk mulai memantau fokus belajar Anda.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs text-text-primary">
                <thead>
                  <tr className="border-b border-border text-[11px] font-bold text-text-secondary uppercase tracking-wider">
                    <th className="py-3 px-4">Mata Kuliah / Topik</th>
                    <th className="py-3 px-4">Durasi</th>
                    <th className="py-3 px-4">Status</th>
                    <th className="py-3 px-4">Tugas Terkait</th>
                    <th className="py-3 px-4">Tanggal & Waktu</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {history.map((sess) => (
                    <tr key={sess.id} className="hover:bg-surface-muted/50 transition-colors">
                      <td className="py-3 px-4 font-semibold">{sess.subject}</td>
                      <td className="py-3 px-4">
                        <div className="font-medium text-text-primary">
                          {formatStudyDuration(sess.duration_seconds)}
                        </div>
                        <div className="text-[10px] text-text-muted font-mono">
                          {formatTimer(sess.duration_seconds)}
                        </div>
                      </td>
                      <td className="py-3 px-4">
                        <span
                          className={`px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider ${
                            sess.status === 'completed'
                              ? 'bg-emerald-100 text-emerald-800'
                              : sess.status === 'active'
                              ? 'bg-sky-100 text-sky-800'
                              : sess.status === 'paused'
                              ? 'bg-amber-100 text-amber-800'
                              : 'bg-slate-100 text-slate-700'
                          }`}
                        >
                          {translateStatus(sess.status)}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-text-secondary">
                        {sess.task && sess.task.title ? (
                          <span className="font-medium text-text-primary">{sess.task.title}</span>
                        ) : (
                          <span className="text-text-muted italic">Belajar mandiri</span>
                        )}
                      </td>
                      <td className="py-3 px-4 text-text-muted">
                        {formatDateTime(sess.started_at)}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </DashboardTemplate>
  );
}
