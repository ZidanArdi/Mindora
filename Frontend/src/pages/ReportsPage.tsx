import { useState, useEffect, useCallback, useId } from 'react';
import { Link } from 'react-router-dom';
import { DashboardTemplate } from '../components/templates';
import { Button } from '../components/atoms';
import {
  ExportIcon,
  CalendarIcon,
  ClockIcon,
  CheckIcon,
  ChartBarIcon,
  SparklesIcon,
} from '../components/atoms/Icons';
import { productivityService } from '../services/productivityService';
import type {
  WeeklyAnalytics,
  ProductivityPrediction,
  UserProfile,
} from '../types';
import { useAuth } from '../hooks';

type PeriodPreset = '7d' | '30d' | 'custom';

function getLocalDateString(d: Date = new Date()): string {
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function getPastDateString(daysAgo: number): string {
  const d = new Date();
  d.setDate(d.getDate() - daysAgo);
  return getLocalDateString(d);
}

export default function ReportsPage() {
  const { user } = useAuth();
  const rawStartInputId = useId();
  const rawEndInputId = useId();
  const startInputId = `report-start-${rawStartInputId.replace(/[^a-zA-Z0-9_-]/g, '')}`;
  const endInputId = `report-end-${rawEndInputId.replace(/[^a-zA-Z0-9_-]/g, '')}`;

  const [periodPreset, setPeriodPreset] = useState<PeriodPreset>('7d');
  const [startDate, setStartDate] = useState<string>(() => getPastDateString(6));
  const [endDate, setEndDate] = useState<string>(() => getLocalDateString());
  const [customStart, setCustomStart] = useState<string>(() => getPastDateString(6));
  const [customEnd, setCustomEnd] = useState<string>(() => getLocalDateString());
  const [validationError, setValidationError] = useState<string | null>(null);

  const [loading, setLoading] = useState<boolean>(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [reportData, setReportData] = useState<WeeklyAnalytics | null>(null);
  const [latestPrediction, setLatestPrediction] = useState<ProductivityPrediction | null>(null);
  const [printedTimestamp, setPrintedTimestamp] = useState<string>(
    () => `${new Date().toLocaleDateString()} ${new Date().toLocaleTimeString()}`
  );

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

  const loadReport = useCallback(
    async (start: string, end: string) => {
      setLoading(true);
      setErrorMessage(null);
      try {
        const userTz = Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC';
        const [analytics, predHist] = await Promise.all([
          productivityService.getRangeAnalytics(start, end, userTz),
          productivityService.listPredictions({ page: 1, limit: 1 }).catch(() => null),
        ]);

        setReportData(analytics);
        if (predHist && predHist.items.length > 0) {
          setLatestPrediction(predHist.items[0]);
        }
      } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : 'Failed to load report data';
        setErrorMessage(msg);
      } finally {
        setLoading(false);
      }
    },
    []
  );

  useEffect(() => {
    loadReport(startDate, endDate);
  }, [loadReport, startDate, endDate]);

  const handlePresetSelect = (preset: PeriodPreset) => {
    setPeriodPreset(preset);
    setValidationError(null);

    if (preset === '7d') {
      const start = getPastDateString(6);
      const end = getLocalDateString();
      setStartDate(start);
      setEndDate(end);
      setCustomStart(start);
      setCustomEnd(end);
    } else if (preset === '30d') {
      const start = getPastDateString(29);
      const end = getLocalDateString();
      setStartDate(start);
      setEndDate(end);
      setCustomStart(start);
      setCustomEnd(end);
    }
  };

  const handleApplyCustomRange = (e: React.FormEvent) => {
    e.preventDefault();
    setValidationError(null);

    if (!customStart || !customEnd) {
      setValidationError('Please select both start and end dates.');
      return;
    }

    const startObj = new Date(customStart);
    const endObj = new Date(customEnd);

    if (isNaN(startObj.getTime()) || isNaN(endObj.getTime())) {
      setValidationError('Invalid date values provided.');
      return;
    }

    if (startObj > endObj) {
      setValidationError('Start date cannot be after end date.');
      return;
    }

    const diffDays = Math.round((endObj.getTime() - startObj.getTime()) / (1000 * 60 * 60 * 24)) + 1;
    if (diffDays > 90) {
      setValidationError('Custom date range cannot exceed 90 days.');
      return;
    }

    setStartDate(customStart);
    setEndDate(customEnd);
  };

  const handleExportCSV = () => {
    if (!reportData || !reportData.daily_points || reportData.daily_points.length === 0) {
      return;
    }

    const rows: string[][] = [];

    // Metadata comments
    rows.push(['# Mindora Productivity Report']);
    rows.push([`# User: ${displayUser.name} (${displayUser.email})`]);
    rows.push([`# Period: ${reportData.start_date} to ${reportData.end_date}`]);
    rows.push([`# Generated: ${new Date().toISOString()}`]);
    rows.push([
      `# Summary: Total Study: ${reportData.total_study_formatted}, Focus: ${reportData.total_focus_minutes}m, Break: ${reportData.total_break_minutes}m, Completed Tasks: ${reportData.total_completed_tasks}, Consistency: ${reportData.study_consistency_percentage}%`,
    ]);
    rows.push([]);

    // Column headers
    rows.push([
      'Date',
      'Study Duration (Hours)',
      'Focus Duration (Hours)',
      'Break Duration (Hours)',
      'Completed Tasks',
      'Sleep Duration (Hours)',
      'Phone Usage (Hours)',
      'Stress Level (1-10)',
    ]);

    // Data rows (strictly preserving NULL as empty rather than 0)
    reportData.daily_points.forEach((dp) => {
      rows.push([
        dp.date,
        dp.study_duration_hours.toFixed(2),
        dp.focus_duration_hours.toFixed(2),
        dp.break_duration_hours.toFixed(2),
        dp.completed_tasks_count.toString(),
        dp.sleep_hours !== undefined && dp.sleep_hours !== null ? dp.sleep_hours.toFixed(1) : '',
        dp.phone_usage_hours !== undefined && dp.phone_usage_hours !== null ? dp.phone_usage_hours.toFixed(1) : '',
        dp.stress_level !== undefined && dp.stress_level !== null ? dp.stress_level.toString() : '',
      ]);
    });

    const csvContent = rows
      .map((row) =>
        row
          .map((cell) => {
            const str = String(cell);
            if (str.includes(',') || str.includes('"') || str.includes('\n')) {
              return `"${str.replace(/"/g, '""')}"`;
            }
            return str;
          })
          .join(',')
      )
      .join('\r\n');

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', `mindora-productivity-report-${reportData.start_date}-to-${reportData.end_date}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  const handlePrint = () => {
    setPrintedTimestamp(`${new Date().toLocaleDateString()} ${new Date().toLocaleTimeString()}`);
    window.print();
  };

  const formatHours = (seconds: number) => {
    const hours = (seconds / 3600).toFixed(1);
    return `${hours}h`;
  };

  const totalDays = reportData?.daily_points.length || 0;
  const activeDays =
    reportData?.daily_points.filter((p) => p.study_duration_hours > 0 || p.focus_duration_hours > 0).length || 0;
  const hasAnyActivity =
    (reportData?.total_study_seconds || 0) > 0 ||
    (reportData?.total_focus_minutes || 0) > 0 ||
    (reportData?.total_completed_tasks || 0) > 0;

  return (
    <DashboardTemplate role={displayUser.role as 'user' | 'admin'} currentUser={displayUser}>
      {/* Print Specific CSS Overrides */}
      <style>{`
        @media print {
          nav, aside, header, .no-print, button, form {
            display: none !important;
          }
          body, main {
            background: #ffffff !important;
            padding: 0 !important;
            margin: 0 !important;
            max-width: 100% !important;
          }
          .print-header {
            display: block !important;
          }
          .shadow-soft-xs, .shadow-soft-sm, .shadow-soft-md {
            box-shadow: none !important;
          }
          .border {
            border-color: #cbd5e1 !important;
          }
          .page-break-inside-avoid {
            page-break-inside: avoid;
          }
        }
      `}</style>

      <div className="space-y-6 pb-12 animate-fadeIn max-w-7xl mx-auto">
        {/* Print Only Official Document Header */}
        <div className="hidden print-header border-b-2 border-slate-900 pb-4 mb-6">
          <div className="flex justify-between items-start">
            <div>
              <h1 className="text-2xl font-black text-slate-900 tracking-tight">
                MINDORA PRODUCTIVITY & HABIT REPORT
              </h1>
              <p className="text-xs text-slate-600 mt-1">
                Student: <strong>{displayUser.name}</strong> ({displayUser.email})
              </p>
            </div>
            <div className="text-right text-xs text-slate-500">
              <p>
                Period: <strong>{reportData?.start_date}</strong> to <strong>{reportData?.end_date}</strong>
              </p>
              <p>Printed: {printedTimestamp || 'Just now'}</p>
            </div>
          </div>
        </div>

        {/* Screen Header Section */}
        <div className="no-print flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-border pb-5">
          <div>
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-nature-mint text-primary">
                Phase 4.2 Reports
              </span>
              <span className="text-xs text-text-muted">Executive Productivity & Habit Audit</span>
            </div>
            <h1 className="text-2xl font-black text-primary-900 tracking-tight mt-1">
              Productivity Reports
            </h1>
            <p className="text-sm text-text-secondary">
              Comprehensive report generated strictly from verified study sessions, focus intervals, and logged daily habits.
            </p>
          </div>

          {/* Action Buttons: CSV Export & Print PDF */}
          <div className="flex flex-wrap items-center gap-2.5">
            <Button
              variant="outline"
              size="sm"
              onClick={handleExportCSV}
              disabled={loading || !reportData || reportData.daily_points.length === 0}
              className="gap-2 text-xs font-semibold"
              title="Download structured CSV file"
            >
              <ExportIcon className="w-4 h-4 text-text-muted" />
              Export CSV
            </Button>

            <Button
              variant="primary"
              size="sm"
              onClick={handlePrint}
              disabled={loading || !reportData}
              className="gap-2 text-xs font-semibold bg-primary hover:bg-primary-hover text-white"
              title="Print or save as PDF via browser"
            >
              <svg
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth={2}
                strokeLinecap="round"
                strokeLinejoin="round"
                className="w-4 h-4"
              >
                <polyline points="6 9 6 2 18 2 18 9" />
                <path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2" />
                <rect x="6" y="14" width="12" height="8" />
              </svg>
              Print / Save PDF
            </Button>
          </div>
        </div>

        {/* Controls: Period Filter Bar */}
        <div className="no-print p-4 rounded-2xl bg-white border border-border shadow-soft-xs space-y-3">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
            <div className="flex items-center gap-1.5 bg-slate-100 rounded-xl p-1">
              <button
                type="button"
                onClick={() => handlePresetSelect('7d')}
                className={`px-3 py-1.5 text-xs font-bold rounded-lg transition-all ${
                  periodPreset === '7d'
                    ? 'bg-white text-primary shadow-soft-xs'
                    : 'text-text-muted hover:text-text-primary'
                }`}
              >
                Last 7 Days
              </button>
              <button
                type="button"
                onClick={() => handlePresetSelect('30d')}
                className={`px-3 py-1.5 text-xs font-bold rounded-lg transition-all ${
                  periodPreset === '30d'
                    ? 'bg-white text-primary shadow-soft-xs'
                    : 'text-text-muted hover:text-text-primary'
                }`}
              >
                Last 30 Days
              </button>
              <button
                type="button"
                onClick={() => setPeriodPreset('custom')}
                className={`px-3 py-1.5 text-xs font-bold rounded-lg transition-all ${
                  periodPreset === 'custom'
                    ? 'bg-white text-primary shadow-soft-xs'
                    : 'text-text-muted hover:text-text-primary'
                }`}
              >
                Custom Range
              </button>
            </div>

            <div className="text-xs text-text-muted flex items-center gap-2">
              <CalendarIcon className="w-4 h-4 text-text-muted" />
              <span>
                Active Window: <strong className="text-text-primary">{startDate}</strong> to{' '}
                <strong className="text-text-primary">{endDate}</strong>
              </span>
            </div>
          </div>

          {/* Custom Date Form (Shown when custom preset active) */}
          {periodPreset === 'custom' && (
            <form
              onSubmit={handleApplyCustomRange}
              className="pt-3 border-t border-slate-100 flex flex-wrap items-center gap-3 text-xs"
            >
              <div className="flex items-center gap-2">
                <label htmlFor={startInputId} className="font-semibold text-text-secondary">
                  From:
                </label>
                <input
                  id={startInputId}
                  type="date"
                  value={customStart}
                  onChange={(e) => setCustomStart(e.target.value)}
                  className="px-2.5 py-1.5 rounded-lg border border-border bg-white text-text-primary font-medium focus:outline-none focus:ring-1 focus:ring-primary"
                />
              </div>

              <div className="flex items-center gap-2">
                <label htmlFor={endInputId} className="font-semibold text-text-secondary">
                  To:
                </label>
                <input
                  id={endInputId}
                  type="date"
                  value={customEnd}
                  onChange={(e) => setCustomEnd(e.target.value)}
                  className="px-2.5 py-1.5 rounded-lg border border-border bg-white text-text-primary font-medium focus:outline-none focus:ring-1 focus:ring-primary"
                />
              </div>

              <Button
                type="submit"
                variant="secondary"
                size="sm"
                className="font-bold text-xs px-3 py-1.5"
              >
                Apply Range
              </Button>
              <span className="text-[11px] text-text-muted">Maximum range: 90 days</span>
            </form>
          )}
        </div>

        {/* Validation Warning Alert */}
        {validationError && (
          <div className="no-print p-4 rounded-xl bg-amber-50 border border-amber-200 text-amber-800 text-xs flex items-center justify-between">
            <div className="flex items-center gap-2">
              <span className="font-bold">Invalid Date Range:</span>
              <span>{validationError}</span>
            </div>
            <button
              type="button"
              onClick={() => setValidationError(null)}
              className="text-amber-700 hover:text-amber-900 font-bold ml-2"
            >
              Dismiss
            </button>
          </div>
        )}

        {/* Backend Error Alert */}
        {errorMessage && (
          <div className="no-print p-4 rounded-xl bg-red-50 border border-red-200 text-red-700 text-xs flex items-center justify-between">
            <div>
              <span className="font-bold">Error loading report: </span>
              <span>{errorMessage}</span>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => loadReport(startDate, endDate)}
              className="text-xs font-bold text-red-700 border-red-300 hover:bg-red-100"
            >
              Retry
            </Button>
          </div>
        )}

        {/* Loading State Skeleton */}
        {loading && (
          <div className="py-20 text-center space-y-3">
            <div className="inline-block w-8 h-8 border-3 border-primary border-t-transparent rounded-full animate-spin" />
            <p className="text-xs font-semibold text-text-muted">Aggregating report metrics across date window...</p>
          </div>
        )}

        {/* Content Body (When Not Loading and Data Available) */}
        {!loading && reportData && (
          <>
            {/* Top Productivity Summary KPI Cards */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 page-break-inside-avoid">
              {/* Metric 1: Total Study Time */}
              <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
                <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
                  Total Study Duration
                </span>
                <div className="text-2xl font-black text-primary-900 mt-2">
                  {reportData.total_study_formatted || '0m'}
                </div>
                <div className="text-[11px] text-text-muted mt-1 flex items-center gap-1">
                  <ClockIcon className="w-3.5 h-3.5 text-text-muted" />
                  <span>
                    {formatHours(reportData.total_study_seconds)} recorded across {activeDays}/{totalDays} days
                  </span>
                </div>
              </div>

              {/* Metric 2: Deep Focus vs Breaks */}
              <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
                <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
                  Focus vs Breaks
                </span>
                <div className="text-2xl font-black text-primary mt-2">
                  {reportData.total_focus_minutes}m
                  <span className="text-xs font-semibold text-text-muted ml-1.5">
                    / {reportData.total_break_minutes}m break
                  </span>
                </div>
                <span className="text-[11px] text-text-muted mt-1 block">
                  Ratio: {reportData.total_focus_minutes > 0 ? (reportData.total_break_minutes / reportData.total_focus_minutes * 100).toFixed(0) : '0'}% recovery time
                </span>
              </div>

              {/* Metric 3: Completed Tasks */}
              <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
                <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
                  Completed Tasks
                </span>
                <div className="text-2xl font-black text-primary-900 mt-2">
                  {reportData.total_completed_tasks}
                </div>
                <div className="text-[11px] text-text-muted mt-1 flex items-center gap-1">
                  <CheckIcon className="w-3.5 h-3.5 text-emerald-600" />
                  <span>
                    {totalDays > 0 ? (reportData.total_completed_tasks / totalDays).toFixed(1) : 0} tasks/day avg
                  </span>
                </div>
              </div>

              {/* Metric 4: Study Consistency */}
              <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
                <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
                  Study Consistency
                </span>
                <div className="text-2xl font-black text-[#138F78] mt-2">
                  {reportData.study_consistency_percentage}%
                </div>
                <div className="w-full bg-slate-100 h-1.5 rounded-full overflow-hidden mt-2">
                  <div
                    className="h-full bg-[#138F78] rounded-full"
                    style={{ width: `${Math.min(100, Math.max(0, reportData.study_consistency_percentage))}%` }}
                  />
                </div>
              </div>
            </div>

            {/* Lifestyle Averages Banner (When Habits are recorded) */}
            <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 flex flex-wrap items-center justify-between gap-4 text-xs page-break-inside-avoid">
              <span className="font-bold text-text-primary uppercase tracking-wider text-[11px]">
                Habit Baseline Averages for Period:
              </span>
              <div className="flex flex-wrap items-center gap-6 text-text-secondary">
                <span>
                  Sleep:{' '}
                  <strong className="text-text-primary font-bold">
                    {reportData.average_sleep_hours !== undefined && reportData.average_sleep_hours !== null
                      ? `${reportData.average_sleep_hours.toFixed(1)} hrs/night`
                      : '— (Unrecorded)'}
                  </strong>
                </span>
                <span>
                  Screen Time:{' '}
                  <strong className="text-text-primary font-bold">
                    {reportData.average_phone_usage_hours !== undefined && reportData.average_phone_usage_hours !== null
                      ? `${reportData.average_phone_usage_hours.toFixed(1)} hrs/day`
                      : '— (Unrecorded)'}
                  </strong>
                </span>
                <span>
                  Stress Level:{' '}
                  <strong className="text-text-primary font-bold">
                    {reportData.average_stress_level !== undefined && reportData.average_stress_level !== null
                      ? `${reportData.average_stress_level.toFixed(1)} / 10`
                      : '— (Unrecorded)'}
                  </strong>
                </span>
              </div>
            </div>

            {/* ML Prediction Model Context Card (Clearly distinguished from observed activity) */}
            <div className="p-5 rounded-2xl bg-gradient-to-r from-slate-900 to-[#123C32] text-white shadow-soft-sm page-break-inside-avoid">
              <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                <div className="space-y-1.5">
                  <div className="flex items-center gap-2">
                    <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-emerald-400/20 text-emerald-300 border border-emerald-400/30">
                      <SparklesIcon className="w-3 h-3 text-emerald-400" />
                      Statistical Model Estimate
                    </span>
                    <span className="text-[10px] font-mono text-white/60">
                      {latestPrediction?.model_version || 'mindora-productivity-ridge-v1.0.0'}
                    </span>
                  </div>
                  <h3 className="text-base font-bold text-white">
                    Estimated Productivity Trajectory
                  </h3>
                  <p className="text-xs text-white/70 max-w-2xl leading-relaxed">
                    Statistical regression output evaluated against habit indicators. This metric represents a mathematical estimate, strictly distinguished from observed study duration and task logs.
                  </p>
                </div>

                <div className="flex items-center gap-4 bg-white/10 px-5 py-3 rounded-xl border border-white/10 shrink-0">
                  <div className="text-right">
                    <span className="text-[10px] uppercase font-bold text-white/70 block">
                      {latestPrediction?.is_baseline ? 'Baseline Estimate' : 'Calibrated Score'}
                    </span>
                    <span className="text-xs text-emerald-300">
                      {latestPrediction ? latestPrediction.prediction_date : 'No model run'}
                    </span>
                  </div>
                  <div className="text-3xl font-black text-emerald-400 font-mono">
                    {latestPrediction ? Math.round(latestPrediction.predicted_score) : '--'}
                    <span className="text-sm font-normal text-white/50">/100</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Daily Breakdown Activity Table */}
            <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4 page-break-inside-avoid">
              <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 border-b border-border pb-4">
                <div>
                  <h2 className="text-lg font-bold text-primary-900">
                    Daily Activity Breakdown
                  </h2>
                  <p className="text-xs text-text-muted">
                    Day-by-day logs across verified sessions and habit check-ins ({startDate} to {endDate})
                  </p>
                </div>
                <div className="no-print flex items-center gap-4 text-xs font-semibold">
                  <span className="flex items-center gap-1.5">
                    <span className="w-2.5 h-2.5 rounded-full bg-[#123C32]" />
                    Study Duration
                  </span>
                  <span className="flex items-center gap-1.5">
                    <span className="w-2.5 h-2.5 rounded-full bg-[#138F78]" />
                    Focus Intervals
                  </span>
                </div>
              </div>

              {!hasAnyActivity ? (
                <div className="py-14 text-center space-y-3">
                  <div className="w-12 h-12 rounded-full bg-slate-100 text-text-muted flex items-center justify-center mx-auto">
                    <ChartBarIcon className="w-6 h-6" />
                  </div>
                  <h3 className="text-sm font-bold text-text-primary">No Activity Recorded in this Period</h3>
                  <p className="text-xs text-text-muted max-w-md mx-auto">
                    No study sessions, focus intervals, or completed tasks were found between {startDate} and {endDate}.
                  </p>
                  <div className="no-print pt-2 flex items-center justify-center gap-3">
                    <Link to="/study">
                      <Button variant="secondary" size="sm" className="text-xs font-bold">
                        Record Study Session
                      </Button>
                    </Link>
                    <Link to="/focus">
                      <Button variant="outline" size="sm" className="text-xs font-bold">
                        Start Focus Timer
                      </Button>
                    </Link>
                  </div>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left text-xs">
                    <thead>
                      <tr className="border-b border-slate-100 text-[11px] uppercase tracking-wider text-text-muted font-bold">
                        <th className="py-2.5 px-3">Date</th>
                        <th className="py-2.5 px-3">Study Duration</th>
                        <th className="py-2.5 px-3">Focus Time</th>
                        <th className="py-2.5 px-3">Break Time</th>
                        <th className="py-2.5 px-3">Completed Tasks</th>
                        <th className="py-2.5 px-3">Sleep</th>
                        <th className="py-2.5 px-3">Phone</th>
                        <th className="py-2.5 px-3">Stress</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 font-medium">
                      {reportData.daily_points.map((point) => (
                        <tr key={point.date} className="hover:bg-slate-50/60 transition-colors">
                          <td className="py-2.5 px-3 font-semibold text-text-primary whitespace-nowrap">
                            {point.date}
                          </td>
                          <td className="py-2.5 px-3">
                            <div className="flex items-center gap-2">
                              <span className="w-12 text-primary-900 font-bold">{point.study_duration_hours}h</span>
                              <div className="w-16 bg-slate-100 h-1.5 rounded-full overflow-hidden">
                                <div
                                  className="h-full bg-[#123C32] rounded-full"
                                  style={{ width: `${Math.min(100, (point.study_duration_hours / 6) * 100)}%` }}
                                />
                              </div>
                            </div>
                          </td>
                          <td className="py-2.5 px-3 text-primary font-bold">
                            {point.focus_duration_hours}h
                          </td>
                          <td className="py-2.5 px-3 text-emerald-600">
                            {point.break_duration_hours}h
                          </td>
                          <td className="py-2.5 px-3 text-text-primary font-bold">
                            {point.completed_tasks_count}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {point.sleep_hours !== undefined && point.sleep_hours !== null
                              ? `${point.sleep_hours}h`
                              : '—'}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {point.phone_usage_hours !== undefined && point.phone_usage_hours !== null
                              ? `${point.phone_usage_hours}h`
                              : '—'}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {point.stress_level !== undefined && point.stress_level !== null
                              ? `${point.stress_level}/10`
                              : '—'}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}

              {/* Table Footer Summary Row */}
              <div className="pt-3 border-t border-slate-100 flex flex-wrap items-center justify-between text-xs text-text-secondary gap-3">
                <div className="flex flex-wrap items-center gap-6">
                  <span>
                    Total Study:{' '}
                    <strong className="text-primary-900 font-bold">
                      {reportData.total_study_formatted}
                    </strong>
                  </span>
                  <span>
                    Total Focus:{' '}
                    <strong className="text-primary font-bold">
                      {reportData.total_focus_minutes}m
                    </strong>
                  </span>
                  <span>
                    Total Breaks:{' '}
                    <strong className="text-emerald-700 font-bold">
                      {reportData.total_break_minutes}m
                    </strong>
                  </span>
                  <span>
                    Total Tasks Finished:{' '}
                    <strong className="text-text-primary font-bold">
                      {reportData.total_completed_tasks}
                    </strong>
                  </span>
                </div>
                <div>
                  Study Consistency:{' '}
                  <strong className="text-[#138F78] font-bold">
                    {reportData.study_consistency_percentage}%
                  </strong>
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </DashboardTemplate>
  );
}
