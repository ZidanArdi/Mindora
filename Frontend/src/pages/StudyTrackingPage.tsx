import { useState, useEffect, useRef } from 'react';
import { DashboardTemplate } from '../components/templates';
import { Button, Input } from '../components/atoms';
import { productivityService } from '../services/productivityService';
import type { StudySession, StudyDailyStats, TaskItem } from '../types';
import { useAuth } from '../hooks';

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

  const loadData = async () => {
    try {
      setLoading(true);
      const [active, hist, stats, taskData] = await Promise.all([
        productivityService.getActiveStudySession(),
        productivityService.listStudySessions({ limit: 10 }),
        productivityService.getDailyStudyStats(),
        productivityService.listTasks({ status: 'pending', limit: 50 }),
      ]);
      setActiveSession(active);
      setHistory(hist.sessions);
      setDailyStats(stats);
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
    } catch (err: unknown) {
      console.error('Failed to load study data:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
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
    if (!confirm('Are you sure you want to cancel this session? Uncompleted duration will be discarded.')) return;
    try {
      await productivityService.cancelStudySession(activeSession.id);
      setActiveSession(null);
      setElapsedSeconds(0);
      loadData();
    } catch (err: unknown) {
      alert(err instanceof Error ? err.message : 'Cancel failed');
    }
  };

  const formatTimer = (totalSec: number) => {
    const hrs = Math.floor(totalSec / 3600);
    const mins = Math.floor((totalSec % 3600) / 60);
    const secs = totalSec % 60;
    return `${hrs.toString().padStart(2, '0')}:${mins
      .toString()
      .padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const currentUserProfile = user
    ? {
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        avatarText: user.name.slice(0, 2).toUpperCase(),
      }
    : { id: '', name: 'Student', email: '', role: 'user' as const, avatarText: 'ST' };

  return (
    <DashboardTemplate role="user" currentUser={currentUserProfile}>
      <div className="space-y-6 max-w-7xl mx-auto">
        {/* Header */}
        <div className="pb-2 border-b border-border">
          <h1 className="text-2xl font-bold text-text-primary tracking-tight">
            Study Tracking
          </h1>
          <p className="text-xs text-text-secondary mt-1">
            Record continuous academic study blocks, associate sessions with tasks, and measure daily study hours.
          </p>
        </div>

        {/* Daily Stats KPI Cards */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-emerald-100 text-emerald-800 flex items-center justify-center text-xl font-bold">
              ⏱️
            </div>
            <div>
              <p className="text-xs font-semibold text-text-secondary">Today's Study Hours</p>
              <h3 className="text-2xl font-black text-text-primary mt-0.5">
                {(dailyStats.total_completed_seconds / 3600).toFixed(1)} hrs
              </h3>
              <p className="text-[10px] text-text-muted mt-0.5">
                {Math.floor(dailyStats.total_completed_seconds / 60)} completed minutes logged
              </p>
            </div>
          </div>

          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex items-center gap-4">
            <div className="w-12 h-12 rounded-xl bg-primary-100 text-primary flex items-center justify-center text-xl font-bold">
              📚
            </div>
            <div>
              <p className="text-xs font-semibold text-text-secondary">Completed Sessions</p>
              <h3 className="text-2xl font-black text-text-primary mt-0.5">
                {dailyStats.completed_sessions_count}
              </h3>
              <p className="text-[10px] text-text-muted mt-0.5">Finished sessions contributing to daily totals</p>
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
                  {activeSession.status === 'active' ? 'SESSION IN PROGRESS' : 'SESSION PAUSED'}
                </span>
                <h2 className="text-2xl font-bold tracking-tight mt-2">{activeSession.subject}</h2>
                {activeSession.task && (
                  <p className="text-xs text-emerald-200 mt-1">
                    Linked to task: <strong className="underline">{activeSession.task.title}</strong>
                  </p>
                )}
              </div>

              {/* Running Time Display */}
              <div className="text-center sm:text-right">
                <span className="text-4xl sm:text-5xl font-mono font-bold tracking-widest text-emerald-300">
                  {formatTimer(elapsedSeconds)}
                </span>
                <p className="text-[10px] text-emerald-200/80 mt-1 uppercase tracking-wider">
                  Server Synchronized Elapsed Time
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
                  Pause Session
                </Button>
              ) : (
                <Button
                  variant="primary"
                  size="md"
                  onClick={handleResume}
                  className="!bg-emerald-500 hover:!bg-emerald-600 text-white text-xs font-semibold"
                >
                  Resume Session
                </Button>
              )}

              <Button
                variant="primary"
                size="md"
                onClick={handleFinish}
                className="!bg-[#138F78] hover:!bg-[#0e6f5d] text-white text-xs font-bold"
              >
                Finish & Log Study Block
              </Button>

              <button
                type="button"
                onClick={handleCancel}
                className="px-4 py-2 rounded-xl text-xs font-semibold text-rose-300 hover:text-rose-100 hover:bg-rose-900/30 transition-colors ml-auto"
              >
                Cancel Session
              </button>
            </div>
          </div>
        ) : (
          <div className="p-6 rounded-3xl bg-white border border-border shadow-soft-xs space-y-4">
            <h2 className="text-base font-bold text-text-primary">Start a New Study Session</h2>
            <form onSubmit={handleStart} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-semibold text-text-primary mb-1">
                    Study Subject / Topic *
                  </label>
                  <Input
                    required
                    placeholder="e.g. Distributed Consensus Algorithms Lecture 5"
                    value={subject}
                    onChange={(e) => setSubject(e.target.value)}
                    className="w-full text-xs"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-text-primary mb-1">
                    Associate with Existing Task (Optional)
                  </label>
                  <select
                    value={selectedTaskID}
                    onChange={(e) => setSelectedTaskID(e.target.value)}
                    className="w-full h-10 px-3 rounded-xl border border-border bg-white text-xs text-text-primary focus:outline-none focus:ring-2 focus:ring-primary/20"
                  >
                    <option value="">No task association (Independent study)</option>
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
                  <span>Start Study Session</span>
                </Button>
              </div>
            </form>
          </div>
        )}

        {/* Study History Table */}
        <div className="bg-white rounded-3xl border border-border shadow-soft-xs p-6 space-y-4">
          <h2 className="text-base font-bold text-text-primary">Recent Study Sessions</h2>

          {loading ? (
            <p className="text-xs text-text-secondary animate-pulse">Loading study sessions...</p>
          ) : history.length === 0 ? (
            <div className="p-8 text-center text-xs text-text-muted border-2 border-dashed border-border rounded-2xl">
              No study sessions logged yet. Start a session above to begin tracking your academic focus.
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs text-text-primary">
                <thead>
                  <tr className="border-b border-border text-[11px] font-bold text-text-secondary uppercase tracking-wider">
                    <th className="py-3 px-4">Subject</th>
                    <th className="py-3 px-4">Duration</th>
                    <th className="py-3 px-4">Status</th>
                    <th className="py-3 px-4">Linked Task</th>
                    <th className="py-3 px-4">Date</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/60">
                  {history.map((sess) => (
                    <tr key={sess.id} className="hover:bg-surface-muted/50 transition-colors">
                      <td className="py-3 px-4 font-semibold">{sess.subject}</td>
                      <td className="py-3 px-4 font-mono">
                        {Math.floor(sess.duration_seconds / 60)} mins ({formatTimer(sess.duration_seconds)})
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
                          {sess.status}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-text-secondary">
                        {sess.task ? sess.task.title : '—'}
                      </td>
                      <td className="py-3 px-4 text-text-muted">
                        {new Date(sess.started_at).toLocaleString()}
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
