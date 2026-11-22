import { useState, useEffect, useCallback, useId } from 'react';
import { DashboardTemplate } from '../../components/templates';
import { Button } from '../../components/atoms';
import { productivityService } from '../../services/productivityService';
import type {
  DailyAnalytics,
  WeeklyAnalytics,
  MonthlyAnalytics,
  AnalyticsSummary,
  ProductivityPrediction,
  PredictionHistoryResponse,
  UserProfile,
} from '../../types';
import { useAuth } from '../../hooks';

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
    return `${hours} jam`;
  };

  return (
    <DashboardTemplate role="user" currentUser={displayUser}>
      <div className="space-y-6 pb-12 animate-fadeIn max-w-7xl mx-auto">
        {/* Header Section */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-border pb-5">
          <div>
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-nature-mint text-primary">
                Fase 3 Aktif
              </span>
              <span className="text-xs text-text-muted">Analitik & Mesin ML Terintegrasi</span>
            </div>
            <h1 className="text-2xl font-black text-primary-900 tracking-tight mt-1">
              Analitik Produktivitas & Prediksi ML
            </h1>
            <p className="text-sm text-text-secondary">
              Analitik objektif yang dihitung murni dari sesi belajar terverifikasi, interval fokus, dan catatan kebiasaan harian.
            </p>
          </div>

          {/* Controls: Date Picker & Period Tabs */}
          <div className="flex flex-wrap items-center gap-2">
            <div className="flex items-center bg-white border border-border rounded-xl p-1 shadow-soft-xs">
              <label htmlFor={dateInputId} className="sr-only">Pilih tanggal target analitik</label>
              <input
                id={dateInputId}
                type="date"
                value={selectedDate}
                onChange={(e) => setSelectedDate(e.target.value)}
                className="text-xs font-semibold text-text-primary px-2.5 py-1.5 bg-transparent border-none focus:outline-none"
              />
            </div>

            <div className="flex items-center bg-slate-100 rounded-xl p-1">
              {[
                { tab: 'daily' as const, label: 'Harian' },
                { tab: 'weekly' as const, label: 'Mingguan' },
                { tab: 'monthly' as const, label: 'Bulanan' },
              ].map(({ tab, label }) => (
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
                  {label}
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
              Tutup
            </button>
          </div>
        )}

        {/* Top KPI Cards Row */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
            <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
              Waktu Belajar (7 Hari)
            </span>
            <div className="text-2xl font-black text-primary-900 mt-2">
              {summaryData ? summaryData.week_study_formatted : '0 jam 0 mnt'}
            </div>
            <span className="text-[11px] text-text-muted mt-1 block">
              {summaryData ? `${summaryData.today_study_formatted} tercatat hari ini` : 'Belum ada belajar hari ini'}
            </span>
          </div>

          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
            <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
              Fokus Mendalam (7 Hari)
            </span>
            <div className="text-2xl font-black text-primary mt-2">
              {summaryData ? `${summaryData.week_focus_minutes} mnt` : '0 mnt'}
            </div>
            <span className="text-[11px] text-text-muted mt-1 block">
              {summaryData ? `${summaryData.today_focus_minutes} mnt fokus hari ini` : 'Belum ada interval'}
            </span>
          </div>

          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
            <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
              Tugas Selesai (7 Hari)
            </span>
            <div className="text-2xl font-black text-primary-900 mt-2">
              {summaryData ? summaryData.week_completed_tasks : 0}
            </div>
            <span className="text-[11px] text-text-muted mt-1 block">
              {summaryData ? `${summaryData.today_completed_tasks} tugas selesai hari ini` : '0 hari ini'}
            </span>
          </div>

          <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs">
            <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
              Konsistensi Belajar
            </span>
            <div className="text-2xl font-black text-[#138F78] mt-2">
              {summaryData ? `${summaryData.study_consistency_percentage}%` : '0%'}
            </div>
            <span className="text-[11px] text-text-muted mt-1 block">
              Hari belajar aktif / periode 7 hari
            </span>
          </div>
        </div>

        {/* Middle Section: ML Prediction Engine Showcase & History */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Prediction Action Card */}
          <div className="p-6 rounded-2xl bg-gradient-to-br from-[#123C32] to-[#0A261F] text-white shadow-soft-md flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between">
                <span
                  className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider ${
                    latestPrediction?.is_baseline || latestPrediction?.prediction_state === 'baseline'
                      ? 'bg-amber-400/20 text-amber-300 border border-amber-400/30'
                      : 'bg-emerald-400/20 text-emerald-300 border border-emerald-400/30'
                  }`}
                >
                  {latestPrediction?.is_baseline || latestPrediction?.prediction_state === 'baseline'
                    ? 'Estimasi Awal'
                    : (latestPrediction?.state_label || 'Estimasi Dipersonalisasi')}
                </span>
                <span className="text-[10px] font-mono text-white/60">
                  {latestPrediction?.model_version || 'mindora-productivity-ridge-v1.0.0'}
                </span>
              </div>

              <h2 className="text-lg font-bold text-white mt-3">
                Estimasi Skor Produktivitas
              </h2>
              <p className="text-xs text-white/70 mt-1 leading-relaxed">
                {latestPrediction?.state_description ||
                  (latestPrediction?.is_baseline
                    ? 'Estimasi ini menggunakan nilai referensi untuk satu atau lebih kebiasaan yang belum dicatat. Catat kebiasaan harian Anda untuk memperoleh estimasi yang lebih personal.'
                    : 'Estimasi ini dihitung menggunakan durasi belajar, tidur, penggunaan ponsel, dan tingkat stres yang Anda catat.')}
              </p>

              {/* Cold-Start & Unrecorded Habits Transparency Box */}
              {latestPrediction?.is_baseline &&
                latestPrediction?.unrecorded_features &&
                latestPrediction.unrecorded_features.length > 0 && (
                  <div className="mt-3 p-3 rounded-xl bg-amber-500/10 border border-amber-500/20 text-[11px] text-amber-200/90 leading-snug">
                    <span className="font-bold text-amber-300">Nilai Referensi Diterapkan:</span>{' '}
                    Kebiasaan yang belum dicatat (
                    {latestPrediction.unrecorded_features
                      .map((f) => f.replace(/_/g, ' '))
                      .join(', ')}
                    ) diimputasi menggunakan nilai referensi populasi. Lengkapi catatan Anda di Pelacakan Belajar dan Gaya Hidup untuk mendapatkan estimasi yang dipersonalisasi.
                  </div>
                )}

              {/* Score Display */}
              <div className="my-6 text-center py-4 rounded-xl bg-white/5 border border-white/10">
                <div className="text-5xl font-black tracking-tight text-emerald-400">
                  {latestPrediction ? Math.round(latestPrediction.predicted_score) : '--'}
                  <span className="text-xl font-normal text-white/50">/100</span>
                </div>
                <span className="text-[11px] font-medium text-white/70 mt-1 block">
                  {latestPrediction
                    ? `Diestimasi untuk tanggal: ${latestPrediction.prediction_date}`
                    : 'Belum ada prediksi dihitung'}
                </span>
              </div>

              {/* Disclaimer */}
              <div className="p-3 rounded-lg bg-white/5 text-[11px] text-white/80 border border-white/10 leading-snug">
                <span className="font-bold text-emerald-300">Estimasi Statistik:</span> Skor ini merupakan estimasi regresi probabilistik berdasarkan pola kebiasaan akademik. Bukan merupakan nilai ujian objektif atau nilai akademik resmi.
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
                {predicting ? 'Menghitung Regresi...' : 'Jalankan Prediksi Produktivitas'}
              </Button>
            </div>
          </div>

          {/* Prediction History Table */}
          <div className="lg:col-span-2 p-6 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between pb-3 border-b border-border">
                <div>
                  <h2 className="text-base font-bold text-primary-900">Riwayat Prediksi & Cuplikan Fitur</h2>
                  <p className="text-xs text-text-muted">Log audit dari estimasi machine learning sebelumnya</p>
                </div>
                <span className="text-xs font-mono font-bold text-primary bg-nature-mint px-2 py-0.5 rounded-md">
                  {predictionHistory?.total || 0} total catatan
                </span>
              </div>

              {loading ? (
                <div className="py-12 text-center text-xs text-text-muted">Memuat riwayat prediksi...</div>
              ) : predictionHistory && predictionHistory.items.length > 0 ? (
                <div className="overflow-x-auto mt-3">
                  <table className="w-full text-left text-xs">
                    <thead>
                      <tr className="border-b border-slate-100 text-[11px] uppercase tracking-wider text-text-muted font-bold">
                        <th className="py-2.5 px-3">Tanggal</th>
                        <th className="py-2.5 px-3">Jenis</th>
                        <th className="py-2.5 px-3">Skor</th>
                        <th className="py-2.5 px-3">Belajar</th>
                        <th className="py-2.5 px-3">Tidur</th>
                        <th className="py-2.5 px-3">Ponsel</th>
                        <th className="py-2.5 px-3">Stres</th>
                        <th className="py-2.5 px-3">Model</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100 font-medium">
                      {predictionHistory.items.map((item) => (
                        <tr key={item.id} className="hover:bg-slate-50/60 transition-colors">
                          <td className="py-2.5 px-3 font-semibold text-text-primary">
                            {item.prediction_date}
                          </td>
                          <td className="py-2.5 px-3">
                            <span
                              className={`inline-flex items-center px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider ${
                                item.is_baseline || item.prediction_state === 'baseline'
                                  ? 'bg-amber-100 text-amber-800'
                                  : 'bg-emerald-100 text-emerald-800'
                              }`}
                            >
                              {item.is_baseline || item.prediction_state === 'baseline'
                                ? 'Awal'
                                : 'Personal'}
                            </span>
                          </td>
                          <td className="py-2.5 px-3 font-bold text-primary">
                            {item.predicted_score}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">{item.study_hours} jam</td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {item.sleep_hours !== undefined && item.sleep_hours !== null ? `${item.sleep_hours} jam` : 'â€”'}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {item.phone_usage_hours !== undefined && item.phone_usage_hours !== null ? `${item.phone_usage_hours} jam` : 'â€”'}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {item.stress_level !== undefined && item.stress_level !== null ? `${item.stress_level}/10` : 'â€”'}
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
                  Belum ada prediksi tercatat. Klik &quot;Jalankan Prediksi Produktivitas&quot; untuk menghitung estimasi pertama Anda.
                </div>
              )}
            </div>

            <div className="pt-3 border-t border-slate-100 text-[11px] text-text-muted flex items-center justify-between">
              <span>Fitur: Belajar (jam), Tidur (jam), Ponsel (jam), Stres (1-10)</span>
              <span>Ridge Terregularisasi L2 (Î±=1.0)</span>
            </div>
          </div>
        </div>

        {/* Bottom Section: Trend Breakdown (Daily / Weekly / Monthly) */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 border-b border-border pb-4">
            <div>
              <h2 className="text-lg font-bold text-primary-900 capitalize">
                Rincian Aktivitas & Kebiasaan {activeTab === 'daily' ? 'Harian' : activeTab === 'weekly' ? 'Mingguan' : 'Bulanan'}
              </h2>
              <p className="text-xs text-text-muted">
                {activeTab === 'daily' && `Rincian detail untuk tanggal ${selectedDate}`}
                {activeTab === 'weekly' && `Tren 7 hari berakhir pada ${selectedDate}`}
                {activeTab === 'monthly' && `Ringkasan 30 hari berakhir pada ${selectedDate}`}
              </p>
            </div>
            <div className="flex items-center gap-4 text-xs font-semibold">
              <span className="flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-[#123C32]" />
                Durasi Belajar
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-[#138F78]" />
                Waktu Fokus
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-2.5 h-2.5 rounded-full bg-emerald-400" />
                Istirahat
              </span>
            </div>
          </div>

          {loading ? (
            <div className="py-16 text-center text-xs text-text-muted">Memuat visualisasi aktivitas...</div>
          ) : activeTab === 'daily' && dailyData ? (
            /* Daily View */
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 space-y-2">
                <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
                  Sesi Belajar
                </span>
                <div className="text-2xl font-black text-primary-900">
                  {dailyData.total_study_formatted || '0 mnt'}
                </div>
                <span className="text-xs text-text-secondary block">
                  {dailyData.completed_study_sessions} sesi akademik selesai
                </span>
              </div>

              <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 space-y-2">
                <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
                  Waktu Fokus vs Istirahat
                </span>
                <div className="text-2xl font-black text-primary">
                  {dailyData.total_focus_minutes} mnt{' '}
                  <span className="text-sm font-semibold text-text-muted">/ {dailyData.total_break_minutes} mnt istirahat</span>
                </div>
                <span className="text-xs text-text-secondary block">
                  {dailyData.completed_focus_count} blok fokus Â· {dailyData.completed_breaks_count} interval istirahat
                </span>
              </div>

              <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 space-y-2">
                <span className="text-xs font-bold text-text-muted uppercase tracking-wider block">
                  Tugas Selesai
                </span>
                <div className="text-2xl font-black text-primary-900">
                  {dailyData.completed_tasks_count}
                </div>
                <span className="text-xs text-text-secondary block">
                  Selesai pada {dailyData.date}
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
                      <th className="py-2.5 px-3">Tanggal</th>
                      <th className="py-2.5 px-3">Waktu Belajar</th>
                      <th className="py-2.5 px-3">Waktu Fokus</th>
                      <th className="py-2.5 px-3">Waktu Istirahat</th>
                      <th className="py-2.5 px-3">Tugas Selesai</th>
                      <th className="py-2.5 px-3">Tidur</th>
                      <th className="py-2.5 px-3">Ponsel</th>
                      <th className="py-2.5 px-3">Stres</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100 font-medium">
                    {((activeTab === 'weekly' ? weeklyData?.daily_points : monthlyData?.daily_points) || []).map(
                      (point) => (
                        <tr key={point.date} className="hover:bg-slate-50/60 transition-colors">
                          <td className="py-2.5 px-3 font-semibold text-text-primary">{point.date}</td>
                          <td className="py-2.5 px-3">
                            <div className="flex items-center gap-2">
                              <span className="w-12 text-primary-900 font-bold">{point.study_duration_hours} jam</span>
                              <div className="w-20 bg-slate-100 h-1.5 rounded-full overflow-hidden">
                                <div
                                  className="h-full bg-[#123C32] rounded-full"
                                  style={{ width: `${Math.min(100, (point.study_duration_hours / 8) * 100)}%` }}
                                />
                              </div>
                            </div>
                          </td>
                          <td className="py-2.5 px-3 text-primary font-bold">{point.focus_duration_hours} jam</td>
                          <td className="py-2.5 px-3 text-emerald-600">{point.break_duration_hours} jam</td>
                          <td className="py-2.5 px-3 text-text-primary font-bold">{point.completed_tasks_count}</td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {point.sleep_hours !== undefined && point.sleep_hours !== null ? `${point.sleep_hours} jam` : 'â€”'}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {point.phone_usage_hours !== undefined && point.phone_usage_hours !== null ? `${point.phone_usage_hours} jam` : 'â€”'}
                          </td>
                          <td className="py-2.5 px-3 text-text-secondary">
                            {point.stress_level !== undefined && point.stress_level !== null ? `${point.stress_level}/10` : 'â€”'}
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
                    Total Belajar:{' '}
                    <strong className="text-primary-900">
                      {activeTab === 'weekly' ? formatHours(weeklyData?.total_study_seconds || 0) : formatHours(monthlyData?.total_study_seconds || 0)}
                    </strong>
                  </span>
                  <span>
                    Total Fokus:{' '}
                    <strong className="text-primary">
                      {activeTab === 'weekly' ? `${weeklyData?.total_focus_minutes || 0} mnt` : `${monthlyData?.total_focus_minutes || 0} mnt`}
                    </strong>
                  </span>
                  <span>
                    Total Tugas Selesai:{' '}
                    <strong className="text-text-primary">
                      {activeTab === 'weekly' ? weeklyData?.total_completed_tasks : monthlyData?.total_completed_tasks}
                    </strong>
                  </span>
                </div>
                <div>
                  Konsistensi Belajar:{' '}
                  <strong className="text-[#138F78]">
                    {activeTab === 'weekly' ? `${weeklyData?.study_consistency_percentage}%` : `${monthlyData?.study_consistency_percentage}%`}
                  </strong>
                </div>
              </div>
            </div>
          ) : (
            <div className="py-12 text-center text-xs text-text-muted">
              Tidak ada aktivitas tercatat untuk rentang waktu ini. Selesaikan sesi belajar atau interval fokus untuk melihat analitik tren.
            </div>
          )}
        </div>
      </div>
    </DashboardTemplate>
  );
}

