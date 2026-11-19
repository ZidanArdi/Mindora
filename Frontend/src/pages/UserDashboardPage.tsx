import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { DashboardTemplate } from '../components/templates';
import { UserHeroBanner, AIAssistantWidget, FocusTimerWidget } from '../components/organisms';
import { CalendarWidget } from '../components/molecules';
import { ShieldIcon, Button, CheckIcon } from '../components/atoms';
import { productivityService } from '../services/productivityService';
import type { DashboardSummary, UserProfile, TaskStatus, ProductivityPrediction } from '../types';
import { useAuth } from '../hooks';

export default function UserDashboardPage() {
  const navigate = useNavigate();
  const { user } = useAuth();

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
            .toUpperCase() || 'ST',
        workspaceName: 'Academic Workspace',
      }
    : {
        id: 'usr_guest',
        name: 'Student',
        email: 'student@mindora.app',
        role: 'user',
        avatarText: 'ST',
        workspaceName: 'Academic Workspace',
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
      return `${h}h ${m}m`;
    }
    return `${mins}m`;
  };

  return (
    <DashboardTemplate role="user" currentUser={displayUser}>
      {isAdminViewingUserWorkspace ? (
        <div className="p-8 rounded-2xl bg-surface-card border border-border space-y-4 text-center max-w-2xl mx-auto shadow-soft-sm my-12 animate-fadeIn">
          <div className="w-12 h-12 rounded-xl bg-nature-mint text-primary flex items-center justify-center mx-auto">
            <ShieldIcon className="w-6 h-6" />
          </div>
          <h2 className="text-xl font-bold text-text-primary">
            Student Privacy Policy Enforcement
          </h2>
          <p className="text-sm text-text-secondary leading-relaxed">
            In accordance with Mindora RBAC and student data privacy rules, personal productivity records, private study tasks, and focus timer sessions are strictly isolated and cannot be accessed from an administrative account.
          </p>
          <div className="pt-2 flex justify-center">
            <Button
              variant="primary"
              size="md"
              onClick={() => navigate('/admin')}
              className="text-xs font-bold"
            >
              Go to Admin Console
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
                  Task Progress
                </span>
                <Link to="/tasks" className="text-[11px] font-bold text-primary hover:underline">
                  View →
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
                  ? `${summary.completed_tasks} of ${summary.total_tasks} tasks completed`
                  : '0 tasks created'}
              </span>
            </div>

            {/* Study Time Today */}
            <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
                  Today's Study
                </span>
                <Link to="/study" className="text-[11px] font-bold text-primary hover:underline">
                  Study Room →
                </Link>
              </div>
              <div className="my-2 text-2xl font-black text-[#123C32]">
                {summary ? formatMinutes(summary.total_study_seconds) : '0m'}
              </div>
              <span className="text-[11px] text-[#526B60]">
                {summary
                  ? `${summary.completed_study_sessions} study sessions finished`
                  : 'No active session'}
              </span>
            </div>

            {/* Focus Intervals */}
            <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
                  Deep Focus Blocks
                </span>
                <Link to="/focus" className="text-[11px] font-bold text-primary hover:underline">
                  Intervals →
                </Link>
              </div>
              <div className="my-2 text-2xl font-black text-[#138F78]">
                {summary ? summary.completed_focus_count : 0}
              </div>
              <span className="text-[11px] text-[#526B60]">
                {summary
                  ? `${summary.completed_breaks_count} breaks · ${summary.total_interruptions} distractions`
                  : '0 intervals'}
              </span>
            </div>

            {/* Lifestyle Habit Status */}
            <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
                  Habit Calibration
                </span>
                <Link to="/lifestyle" className="text-[11px] font-bold text-primary hover:underline">
                  Calibrate →
                </Link>
              </div>
              <div className="my-2">
                {summary?.today_lifestyle_logged ? (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-emerald-100 text-[#123C32] text-xs font-bold border border-emerald-300">
                    <span className="w-2 h-2 rounded-full bg-[#138F78]" />
                    Calibrated Today
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-amber-100 text-amber-900 text-xs font-bold border border-amber-300">
                    <span className="w-2 h-2 rounded-full bg-amber-500 animate-pulse" />
                    Pending Entry
                  </span>
                )}
              </div>
              <span className="text-[11px] text-[#526B60]">
                {summary?.today_lifestyle_logged
                  ? 'All 8 variables recorded'
                  : 'Log sleep & screen time today'}
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
                    <h2 className="text-base font-bold text-[#123C32]">Active Study Tasks</h2>
                    <p className="text-xs text-[#526B60]">
                      Your upcoming academic deliverables and assignments.
                    </p>
                  </div>
                  <Link
                    to="/tasks"
                    className="px-3.5 py-1.5 rounded-xl bg-emerald-50 hover:bg-emerald-100 text-[#123C32] text-xs font-bold border border-emerald-200 transition-colors"
                  >
                    Manage All Tasks →
                  </Link>
                </div>

                {loading ? (
                  <div className="py-8 text-center text-xs text-text-muted">Loading tasks...</div>
                ) : !summary || summary.recent_tasks.length === 0 ? (
                  <div className="py-10 text-center space-y-3 bg-surface-muted/30 rounded-2xl border border-dashed border-border">
                    <span className="text-3xl block">📝</span>
                    <p className="text-sm font-semibold text-text-primary">
                      No tasks created yet
                    </p>
                    <p className="text-xs text-text-secondary max-w-sm mx-auto">
                      Organize your homework, course readings, and project deadlines with our task manager.
                    </p>
                    <div className="pt-2">
                      <Link
                        to="/tasks"
                        className="inline-flex px-4 py-2 rounded-xl text-xs font-bold text-white bg-[#138F78] hover:bg-[#123C32] transition-colors shadow-soft-xs"
                      >
                        + Create Your First Task
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
                                {task.category}
                              </span>
                              {task.deadline && (
                                <span className="text-[10px] font-medium text-amber-700">
                                  Due: {new Date(task.deadline).toLocaleDateString([], { month: 'short', day: 'numeric' })}
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
                            {task.priority}
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
                  <span className="text-[10px] font-bold uppercase tracking-wider text-emerald-300 bg-white/10 px-2 py-0.5 rounded-full">
                    ML Productivity Estimate
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
                      {latestPrediction ? `For ${latestPrediction.prediction_date}` : 'No estimate recorded today'}
                    </span>
                  </div>
                  <Link
                    to="/analytics"
                    className="text-xs font-bold text-emerald-300 hover:text-white transition-colors"
                  >
                    Full Analytics →
                  </Link>
                </div>
                <p className="text-[10px] text-white/60 border-t border-white/10 pt-2 leading-relaxed">
                  Estimated from habit patterns (study, sleep, phone, stress). Strictly not an objective grade.
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
