import { useState, useEffect, useCallback, useId } from 'react';
import { DashboardTemplate } from '../components/templates';
import { Button } from '../components/atoms';
import { productivityService } from '../services/productivityService';
import type {
  DailyAnalytics,
  WeeklyAnalytics,
  MonthlyAnalytics,
  AnalyticsSummary,
  ProductivityPrediction,
  PredictionHistoryResponse,
  UserProfile,
} from '../types';
import { useAuth } from '../hooks';

export default function AnalyticsPage() {
  const { user } = useAuth();
  const rawDateInputId = useId();
  const dateInputId = `analytics-date-${rawDateInputId.replace(/[^a-zA-Z0-9_-]/g, '')}`;

  const [activeTab, setActiveTab] = useState<'daily' | 'weekly' | 'monthly'>('weekly');
  const [loading, setLoading] = useState<boolean>(true);
  const [predicting, setPredicting] = useState<boolean>(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Analytics data state
  const [dailyData, setDailyData] = useState<DailyAnalytics | null>(null);
  const [weeklyData, setWeeklyData] = useState<WeeklyAnalytics | null>(null);
  const [monthlyData, setMonthlyData] = useState<MonthlyAnalytics | null>(null);
  const [summaryData, setSummaryData] = useState<AnalyticsSummary | null>(null);

  // Prediction data state
  const [latestPrediction, setLatestPrediction] = useState<ProductivityPrediction | null>(null);
  const [predictionHistory, setPredictionHistory] = useState<PredictionHistoryResponse | null>(null);

  // Selected date filter (default today)
  const [selectedDate, setSelectedDate] = useState<string>(() => {
    return new Date().toISOString().split('T')[0];
  });

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

  const loadData = useCallback(async () => {
    setLoading(true);
    setErrorMessage(null);
    try {
      // 1. Load summary and predictions in parallel
      const [sum, predHist] = await Promise.all([
        productivityService.getAnalyticsSummary().catch(() => null),
        productivityService.listPredictions({ page: 1, limit: 5 }).catch(() => null),
      ]);

      setSummaryData(sum);
      setPredictionHistory(predHist);
      if (sum?.latest_prediction) {
        setLatestPrediction(sum.latest_prediction);
      } else if (predHist && predHist.items.length > 0) {
        setLatestPrediction(predHist.items[0]);
      }

      // 2. Load tab-specific analytics
      if (activeTab === 'daily') {
        const d = await productivityService.getDailyAnalytics(selectedDate);
        setDailyData(d);
      } else if (activeTab === 'weekly') {
        const w = await productivityService.getWeeklyAnalytics(selectedDate);
        setWeeklyData(w);
      } else if (activeTab === 'monthly') {
        const m = await productivityService.getMonthlyAnalytics(selectedDate);
        setMonthlyData(m);
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to load analytics data';
      setErrorMessage(msg);
    } finally {
      setLoading(false);
    }
  }, [activeTab, selectedDate]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Handle trigger new prediction
  const handleRunPrediction = async () => {
    setPredicting(true);
    setErrorMessage(null);
    try {
      const newPred = await productivityService.createPrediction({
        date: selectedDate,
      });
      setLatestPrediction(newPred);
      // Reload history
      const hist = await productivityService.listPredictions({ page: 1, limit: 5 });
      setPredictionHistory(hist);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Failed to calculate prediction';
      setErrorMessage(msg);
    } finally {
      setPredicting(false);
    }
  };

  const formatHours = (seconds: number) => {
    const hours = (seconds / 3600).toFixed(1);
    return `${hours}h`;
  };

  return (
    <DashboardTemplate role="user" currentUser={displayUser}>
      <div className="space-y-6 pb-12 animate-fadeIn max-w-7xl mx-auto">
        {/* Header Section */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border pb-5">
          <div>
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-nature-mint text-primary">
                Phase 3 Live
              </span>
              <span className="text-xs text-text-muted">Persisted Analytics & ML Engine</span>
            </div>
            <h1 className="text-2xl font-black text-primary-900 tracking-tight mt-1">
              Productivity Analytics & ML Prediction
            </h1>
            <p className="text-sm text-text-secondary">
              Objective analytics calculated strictly from verified study sessions, focus intervals, and daily habit logs.
            </p>
          </div>

          {/* Controls: Date Picker & Period Tabs */}
          <div className="flex flex-wrap items-center gap-2">
            <div className="flex items-center bg-white border border-border rounded-xl p-1 shadow-soft-xs">
              <label htmlFor={dateInputId} className="sr-only">Select analytics target date</label>
              <input
                id={dateInputId}
                type="date"
                value={selectedDate}
                onChange={(e) => setSelectedDate(e.target.value)}
                className="text-xs font-semibold text-text-primary px-2.5 py-1.5 bg-transparent border-none focus:outline-none"
              />
            </div>

            <div className="flex items-center bg-slate-100 rounded-xl p-1">
              {(['daily', 'weekly', 'monthly'] as const).map((tab) => (
                <button
                  key={tab}
                  type="button"
                  onClick={() => setActiveTab(tab)}
                  className={`px-3 py-1.5 text-xs font-bold rounded-lg transition-all capitalize ${
                    activeTab === tab
                      ? 'bg-white text-primary shadow-soft-xs'
                      : 'text-text-muted hover:text-text-primary'
                  }`}
                >
                  {tab}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Error Alert */}
        {errorMessage && (
          <div className="p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 text-xs flex items-center justify-between">
            <span>{errorMessage}</span>
            <button
              type="button"
              onClick={() => setErrorMessage(null)}
              className="text-red-500 hover:text-red-700 font-bold ml-2"
            >
              Dismiss
            </button>
          </div>
        )}

        {/* Top KPI Cards Row */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
            <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
              Study Time (7d)
            </span>
            <div className="text-2xl font-black text-primary-900 mt-2">
              {summaryData ? summaryData.week_study_formatted : '0h 0m'}
            </div>
            <span className="text-[11px] text-text-muted mt-1 block">
              {summaryData ? `${summaryData.today_study_formatted} recorded today` : 'No study today'}
            </span>
          </div>

          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
            <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
              Deep Focus (7d)
            </span>
            <div className="text-2xl font-black text-primary mt-2">
              {summaryData ? `${summaryData.week_focus_minutes}m` : '0m'}
            </div>
            <span className="text-[11px] text-text-muted mt-1 block">
              {summaryData ? `${summaryData.today_focus_minutes}m focus today` : 'No intervals'}
            </span>
          </div>

          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
            <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
              Tasks Completed (7d)
            </span>
            <div className="text-2xl font-black text-primary-900 mt-2">
              {summaryData ? summaryData.week_completed_tasks : 0}
            </div>
            <span className="text-[11px] text-text-muted mt-1 block">
              {summaryData ? `${summaryData.today_completed_tasks} tasks finished today` : '0 today'}
            </span>
          </div>

          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
            <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
              Study Consistency
            </span>
            <div className="text-2xl font-black text-[#138F78] mt-2">
              {summaryData ? `${summaryData.study_consistency_percentage}%` : '0%'}
            </div>
            <span className="text-[11px] text-text-muted mt-1 block">
              Active study days / 7-day period
            </span>
          </div>
        </div>

        {/* Middle Section: ML Prediction Engine Showcase & History */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Prediction Action Card */}
          <div className="p-6 rounded-2xl bg-gradient-to-br from-[#123C32] to-[#0A261F] text-white shadow-soft-md flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between">
                <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-white/10 text-emerald-300">
                  ML Prediction Engine
                </span>
                <span className="text-[10px] font-mono text-white/60">
                  {latestPrediction?.model_version || 'mindora-productivity-ridge-v1.0.0'}
                </span>
              </div>

              <h2 className="text-lg font-bold text-white mt-3">
                Estimated Productivity Score
              </h2>
              <p className="text-xs text-white/70 mt-1 leading-relaxed">
                Evaluated by regularized Ridge Regression based on verified study duration, sleep duration, phone usage, and self-reported stress.
              </p>

              {/* Score Display */}
              <div className="my-6 text-center py-4 rounded-xl bg-white/5 border border-white/10">
                <div className="text-5xl font-black tracking-tight text-emerald-400">
                  {latestPrediction ? Math.round(latestPrediction.predicted_score) : '--'}
                  <span className="text-xl font-normal text-white/50">/100</span>
                </div>
                <span className="text-[11px] font-medium text-white/70 mt-1 block">
                  {latestPrediction
                    ? `Estimated for date: ${latestPrediction.prediction_date}`
                    : 'No prediction calculated yet'}
                </span>
              </div>

              {/* Disclaimer */}
              <div className="p-3 rounded-lg bg-white/5 text-[11px] text-white/80 border border-white/10 leading-snug">
                <span className="font-bold text-emerald-300">Statistical Estimate:</span> This score is a probabilistic regression estimate trained on academic habit patterns. It is strictly not an objective test score or academic grade.
              </div>
            </div>

            <div className="pt-4 mt-4 border-t border-white/10">
              <Button
                variant="primary"
                size="md"
                className="w-full justify-center bg-emerald-500 hover:bg-emerald-600 text-slate-900 font-bold border-none"
                onClick={handleRunPrediction}
                disabled={predicting}
              >
                {predicting ? 'Computing Regression...' : 'Run Productivity Prediction'}
              </Button>
            </div>
          </div>

          {/* Prediction History Table */}
          <div className="lg:col-span-2 p-6 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between pb-3 border-b border-border">
                <div>
                  <h2 className="text-base font-bold text-primary-900">Prediction History & Feature Snapshot</h2>
                  <p className="text-xs text-text-muted">Audit log of previous machine learning estimates</p>
                </div>
                <span className="text-xs font-mono font-bold text-primary bg-nature-mint px-2 py-0.5 rounded-md">
                  {predictionHistory?.total || 0} total records
                </span>
              </div>

              {loading ? (
                <div className="py-12 text-center text-xs text-text-muted">Loading prediction history...</div>
              ) : predictionHistory && predictionHistory.items.length > 0 ? (
                <div className="overflow-x-auto mt-3">
                  <table className="w-full text-left text-xs">
                    <thead>
                      <tr className="border-b border-slate-100 text-[11px] uppercase tracking-wider text-text-muted font-bold">
                        <th className="py-2.5 px-3">Date</th>
                        <th className="py-2.5 px-3">Score</th>
                        <th className="py-2.5 px-3">Study</th>
                        <th className="py-2.5 px-3">Sleep</th>
                        <th className="py-2.5 px-3">Phone</th>
                        <th className="py-2.5 px-3">Stress</th>
                        <th className="py-2.5 px-3">Model</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 font-medium">
                      {predictionHistory.items.map((item) => (
                        <tr key={item.id} className="hover:bg-slate-50/60 transition-colors">
                          <td className="py-2.5 px-3 font-semibold text-text-primary">
                            {item.prediction_date}
                          </td>
                          <td className="py-2.5 px-3 font-bold text-primary">
                            {item.predicted_score}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">{item.study_hours}h</td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {item.sleep_hours !== undefined && item.sleep_hours !== null ? `${item.sleep_hours}h` : '—'}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {item.phone_usage_hours !== undefined && item.phone_usage_hours !== null ? `${item.phone_usage_hours}h` : '—'}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {item.stress_level !== undefined && item.stress_level !== null ? `${item.stress_level}/10` : '—'}
                          </td>
                          <td className="py-2.5 px-3 font-mono text-[10px] text-text-muted">
                            {item.model_version.split('-')[2] || 'v1.0.0'}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              ) : (
                <div className="py-12 text-center text-xs text-text-muted">
                  No predictions recorded yet. Click &quot;Run Productivity Prediction&quot; to compute your first estimate.
                </div>
              )}
            </div>

            <div className="pt-3 border-t border-slate-100 text-[11px] text-text-muted flex items-center justify-between">
              <span>Features: Study (h), Sleep (h), Phone (h), Stress (1-10)</span>
              <span>L2 Regularized Ridge (α=1.0)</span>
            </div>
          </div>
        </div>

        {/* Bottom Section: Trend Breakdown (Daily / Weekly / Monthly) */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 border-b border-border pb-4">
            <div>
              <h2 className="text-lg font-bold text-primary-900 capitalize">
                {activeTab} Activity & Habit Breakdown
              </h2>
              <p className="text-xs text-text-muted">
                {activeTab === 'daily' && `Detailed breakdown for ${selectedDate}`}
                {activeTab === 'weekly' && `7-day trend ending on ${selectedDate}`}
                {activeTab === 'monthly' && `30-day overview ending on ${selectedDate}`}
              </p>
            </div>
            <div className="flex items-center gap-4 text-xs font-semibold">
              <span className="flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-[#123C32]" />
                Study Duration
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-[#138F78]" />
                Focus Time
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-400" />
                Breaks
              </span>
            </div>
          </div>

          {loading ? (
            <div className="py-16 text-center text-xs text-text-muted">Loading activity visualization...</div>
          ) : activeTab === 'daily' && dailyData ? (
            /* Daily View */
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 space-y-2">
                <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
                  Study Sessions
                </span>
                <div className="text-2xl font-black text-primary-900">
                  {dailyData.total_study_formatted || '0m'}
                </div>
                <span className="text-xs text-text-secondary block">
                  {dailyData.completed_study_sessions} completed academic session(s)
                </span>
              </div>

              <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 space-y-2">
                <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
                  Focus vs Break Time
                </span>
                <div className="text-2xl font-black text-primary">
                  {dailyData.total_focus_minutes}m{' '}
                  <span className="text-sm font-semibold text-text-muted">/ {dailyData.total_break_minutes}m break</span>
                </div>
                <span className="text-xs text-text-secondary block">
                  {dailyData.completed_focus_count} focus blocks · {dailyData.completed_breaks_count} break intervals
                </span>
              </div>

              <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 space-y-2">
                <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
                  Tasks Completed
                </span>
                <div className="text-2xl font-black text-primary-900">
                  {dailyData.completed_tasks_count}
                </div>
                <span className="text-xs text-text-secondary block">
                  Finished on {dailyData.date}
                </span>
              </div>
            </div>
          ) : (activeTab === 'weekly' && weeklyData) || (activeTab === 'monthly' && monthlyData) ? (
            /* Weekly / Monthly Trend Table & Bars */
            <div className="space-y-4">
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead>
                    <tr className="border-b border-slate-100 text-[11px] uppercase tracking-wider text-text-muted font-bold">
                      <th className="py-2.5 px-3">Date</th>
                      <th className="py-2.5 px-3">Study Time</th>
                      <th className="py-2.5 px-3">Focus Time</th>
                      <th className="py-2.5 px-3">Break Time</th>
                      <th className="py-2.5 px-3">Tasks Finished</th>
                      <th className="py-2.5 px-3">Sleep</th>
                      <th className="py-2.5 px-3">Phone</th>
                      <th className="py-2.5 px-3">Stress</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 font-medium">
                    {((activeTab === 'weekly' ? weeklyData?.daily_points : monthlyData?.daily_points) || []).map(
                      (point) => (
                        <tr key={point.date} className="hover:bg-slate-50/60 transition-colors">
                          <td className="py-2.5 px-3 font-semibold text-text-primary">{point.date}</td>
                          <td className="py-2.5 px-3">
                            <div className="flex items-center gap-2">
                              <span className="w-12 text-primary-900 font-bold">{point.study_duration_hours}h</span>
                              <div className="w-20 bg-slate-100 h-1.5 rounded-full overflow-hidden">
                                <div
                                  className="h-full bg-[#123C32] rounded-full"
                                  style={{ width: `${Math.min(100, (point.study_duration_hours / 8) * 100)}%` }}
                                />
                              </div>
                            </div>
                          </td>
                          <td className="py-2.5 px-3 text-primary font-bold">{point.focus_duration_hours}h</td>
                          <td className="py-2.5 px-3 text-emerald-600">{point.break_duration_hours}h</td>
                          <td className="py-2.5 px-3 text-text-primary font-bold">{point.completed_tasks_count}</td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {point.sleep_hours !== undefined && point.sleep_hours !== null ? `${point.sleep_hours}h` : '—'}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {point.phone_usage_hours !== undefined && point.phone_usage_hours !== null ? `${point.phone_usage_hours}h` : '—'}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {point.stress_level !== undefined && point.stress_level !== null ? `${point.stress_level}/10` : '—'}
                          </td>
                        </tr>
                      )
                    )}
                  </tbody>
                </table>
              </div>

              {/* Summary Stats Footer */}
              <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 flex flex-wrap items-center justify-between text-xs text-text-secondary gap-3">
                <div className="flex items-center gap-6">
                  <span>
                    Total Study:{' '}
                    <strong className="text-primary-900">
                      {activeTab === 'weekly' ? formatHours(weeklyData?.total_study_seconds || 0) : formatHours(monthlyData?.total_study_seconds || 0)}
                    </strong>
                  </span>
                  <span>
                    Total Focus:{' '}
                    <strong className="text-primary">
                      {activeTab === 'weekly' ? `${weeklyData?.total_focus_minutes || 0}m` : `${monthlyData?.total_focus_minutes || 0}m`}
                    </strong>
                  </span>
                  <span>
                    Total Tasks Completed:{' '}
                    <strong className="text-text-primary">
                      {activeTab === 'weekly' ? weeklyData?.total_completed_tasks : monthlyData?.total_completed_tasks}
                    </strong>
                  </span>
                </div>
                <div>
                  Study Consistency:{' '}
                  <strong className="text-[#138F78]">
                    {activeTab === 'weekly' ? `${weeklyData?.study_consistency_percentage}%` : `${monthlyData?.study_consistency_percentage}%`}
                  </strong>
                </div>
              </div>
            </div>
          ) : (
            <div className="py-12 text-center text-xs text-text-muted">
              No activity logged for this time range. Complete study sessions or focus intervals to see trend analytics.
            </div>
          )}
        </div>
      </div>
    </DashboardTemplate>
  );
}
