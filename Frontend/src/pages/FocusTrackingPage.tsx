import { useState, useEffect, useRef } from 'react';
import { DashboardTemplate } from '../components/templates';
import { Button } from '../components/atoms';
import { productivityService } from '../services/productivityService';
import type { FocusSession, FocusDailyStats, UserProfile } from '../types';
import { useAuth } from '../hooks';
import { formatDateTime, formatTime, translateStatus } from '../utils/i18n';

export default function FocusTrackingPage() {
  const { user } = useAuth();
  const [activeInterval, setActiveInterval] = useState<FocusSession | null>(null);
  const [history, setHistory] = useState<FocusSession[]>([]);
  const [dailyStats, setDailyStats] = useState<FocusDailyStats>({
    completed_focus_minutes: 0,
    completed_focus_count: 0,
    completed_breaks_count: 0,
    total_interruptions: 0,
  });
  const [loading, setLoading] = useState(true);

  // Timer configuration
  const [sessionType, setSessionType] = useState<'focus' | 'break'>('focus');
  const [targetMinutes, setTargetMinutes] = useState(25);
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const [interruptions, setInterruptions] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const timerRef = useRef<number | null>(null);

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

  const focusPresets = [15, 25, 45, 60];
  const breakPresets = [5, 10, 15, 30];

  const loadData = async () => {
    try {
      setLoading(true);
      const [active, hist, stats] = await Promise.all([
        productivityService.getActiveFocusInterval(),
        productivityService.listFocusIntervals({ limit: 15 }),
        productivityService.getDailyFocusStats(),
      ]);
      setActiveInterval(active);
      setHistory(hist.intervals);
      setDailyStats(stats);

      if (active) {
        setSessionType(active.session_type);
        setInterruptions(active.interruption_count);
        const startedAtMs = new Date(active.started_at).getTime();
        const initialElapsed = Math.max(
          active.duration_minutes * 60,
          Math.floor((Date.now() - startedAtMs) / 1000)
        );
        setElapsedSeconds(initialElapsed);
      } else {
        setElapsedSeconds(0);
        setInterruptions(0);
      }
    } catch (err: unknown) {
      console.error('Failed to load focus intervals:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  // Timer tick for active interval
  useEffect(() => {
    if (timerRef.current) clearInterval(timerRef.current);

    if (activeInterval && activeInterval.status === 'active') {
      const startedAtMs = new Date(activeInterval.started_at).getTime();
      timerRef.current = window.setInterval(() => {
        const currentElapsed = Math.floor((Date.now() - startedAtMs) / 1000);
        setElapsedSeconds(Math.max(activeInterval.duration_minutes * 60, currentElapsed));
      }, 1000);
    }

    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [activeInterval]);

  const handleStartInterval = async () => {
    try {
      setSubmitting(true);
      const newInterval = await productivityService.startFocusInterval({
        session_type: sessionType,
        target_minutes: targetMinutes,
      });
      setActiveInterval(newInterval);
      setInterruptions(0);
      setElapsedSeconds(0);
      await loadData();
    } catch (err: unknown) {
      console.error('Failed to start focus interval:', err);
      alert('Gagal memulai sesi. Pastikan tidak ada interval lain yang sedang aktif.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleIncrementInterruption = async () => {
    if (!activeInterval) return;
    try {
      const updated = await productivityService.incrementInterruption(activeInterval.id);
      setActiveInterval(updated);
      setInterruptions(updated.interruption_count);
    } catch (err: unknown) {
      console.error('Failed to increment interruption:', err);
    }
  };

  const handleFinishInterval = async () => {
    if (!activeInterval) return;
    try {
      setSubmitting(true);
      await productivityService.finishFocusInterval(activeInterval.id);
      setActiveInterval(null);
      setElapsedSeconds(0);
      setInterruptions(0);
      await loadData();
    } catch (err: unknown) {
      console.error('Failed to finish interval:', err);
      alert('Gagal menyelesaikan interval.');
    } finally {
      setSubmitting(false);
    }
  };

  const handleAbandonInterval = async () => {
    if (!activeInterval) return;
    if (!window.confirm('Apakah Anda yakin ingin membatalkan interval ini?')) return;
    try {
      setSubmitting(true);
      await productivityService.abandonFocusInterval(activeInterval.id);
      setActiveInterval(null);
      setElapsedSeconds(0);
      setInterruptions(0);
      await loadData();
    } catch (err: unknown) {
      console.error('Failed to abandon interval:', err);
    } finally {
      setSubmitting(false);
    }
  };

  const formatTimerDigits = (totalSeconds: number) => {
    const mins = Math.floor(totalSeconds / 60);
    const secs = totalSeconds % 60;
    return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
  };

  const formatMinutesDisplay = (minutes: number) => {
    if (minutes >= 60) {
      const h = Math.floor(minutes / 60);
      const m = minutes % 60;
      return `${h} jam ${m} mnt`;
    }
    return `${minutes} mnt`;
  };

  const targetSecs = targetMinutes * 60;
  const progressPercent = Math.min(
    100,
    targetSecs > 0 ? (elapsedSeconds / targetSecs) * 100 : 0
  );

  return (
    <DashboardTemplate role="user" currentUser={displayUser}>
      <div className="space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-2 border-b border-border/60">
          <div>
            <h1 className="text-2xl sm:text-3xl font-extrabold text-[#123C32] tracking-tight">
              Interval Fokus & Pomodoro
            </h1>
            <p className="text-sm text-[#526B60] mt-1">
              Sesi fokus berkala dan istirahat pemulihan dengan pelacakan distraksi yang terverifikasi.
            </p>
          </div>
          {activeInterval && (
            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-emerald-100/80 border border-emerald-300 text-emerald-800 text-xs font-bold animate-pulse">
              <span className="w-2 h-2 rounded-full bg-emerald-600" />
              <span>Interval Sedang Berjalan ({activeInterval.session_type === 'focus' ? 'FOKUS' : 'ISTIRAHAT'})</span>
            </div>
          )}
        </div>

        {/* Daily Stats Summary */}
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="p-4 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
            <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
              Waktu Fokus Hari Ini
            </span>
            <div className="mt-2 text-2xl font-black text-[#123C32]">
              {formatMinutesDisplay(dailyStats.completed_focus_minutes)}
            </div>
            <span className="text-[11px] text-[#526B60] mt-1">
              {dailyStats.completed_focus_count} blok fokus selesai
            </span>
          </div>

          <div className="p-4 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
            <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
              Blok Istirahat
            </span>
            <div className="mt-2 text-2xl font-black text-teal-700">
              {dailyStats.completed_breaks_count}
            </div>
            <span className="text-[11px] text-[#526B60] mt-1">
              Interval istirahat pemulihan
            </span>
          </div>

          <div className="p-4 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
            <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
              Interval Selesai
            </span>
            <div className="mt-2 text-2xl font-black text-[#138F78]">
              {dailyStats.completed_focus_count + dailyStats.completed_breaks_count}
            </div>
            <span className="text-[11px] text-[#526B60] mt-1">Di semua jenis sesi</span>
          </div>

          <div className="p-4 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between">
            <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
              Total Distraksi
            </span>
            <div className="mt-2 text-2xl font-black text-amber-600">
              {dailyStats.total_interruptions}
            </div>
            <span className="text-[11px] text-[#526B60] mt-1">Poin distraksi tercatat</span>
          </div>
        </div>

        {/* Main Focus Console Grid */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          {/* Active Interval Timer Card (7 cols) */}
          <div className="lg:col-span-7 bg-white rounded-3xl border border-emerald-100 p-6 sm:p-8 shadow-soft-sm relative overflow-hidden">
            {/* Soft botanical accent silhouette */}
            <div className="absolute right-0 top-0 pointer-events-none opacity-10">
              <svg width="220" height="220" viewBox="0 0 200 200" fill="none">
                <circle cx="100" cy="100" r="80" stroke="#138F78" strokeWidth="8" strokeDasharray="12 12" />
                <path d="M100 20 L100 180 M20 100 L180 100" stroke="#138F78" strokeWidth="2" />
              </svg>
            </div>

            <div className="relative z-10 flex flex-col items-center text-center space-y-6">
              {/* Session Type Selector (disabled if running) */}
              {!activeInterval ? (
                <div className="inline-flex p-1 rounded-2xl bg-surface-muted border border-border">
                  <button
                    type="button"
                    onClick={() => {
                      setSessionType('focus');
                      setTargetMinutes(25);
                    }}
                    className={`px-5 py-2 rounded-xl text-xs font-bold transition-all ${
                      sessionType === 'focus'
                        ? 'bg-[#123C32] text-white shadow-soft-xs'
                        : 'text-[#526B60] hover:text-[#123C32]'
                    }`}
                  >
                    Fokus Mendalam
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setSessionType('break');
                      setTargetMinutes(5);
                    }}
                    className={`px-5 py-2 rounded-xl text-xs font-bold transition-all ${
                      sessionType === 'break'
                        ? 'bg-[#138F78] text-white shadow-soft-xs'
                        : 'text-[#526B60] hover:text-[#123C32]'
                    }`}
                  >
                    Istirahat Pemulihan
                  </button>
                </div>
              ) : (
                <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-emerald-50 border border-emerald-200">
                  <span
                    className={`w-2.5 h-2.5 rounded-full ${
                      activeInterval.session_type === 'focus' ? 'bg-[#138F78]' : 'bg-teal-500'
                    }`}
                  />
                  <span className="text-xs font-bold text-[#123C32] uppercase tracking-wider">
                    {activeInterval.session_type === 'focus' ? 'Sesi Fokus Mendalam' : 'Istirahat Pemulihan'}
                  </span>
                </div>
              )}

              {/* Large Timer Visualizer */}
              <div className="relative w-64 h-64 flex flex-col items-center justify-center rounded-full bg-linear-to-b from-emerald-50/70 to-teal-50/40 border-8 border-emerald-100/80 shadow-inner">
                {/* SVG Circular Progress */}
                <svg className="absolute inset-0 w-full h-full -rotate-90">
                  <circle
                    cx="120"
                    cy="120"
                    r="110"
                    className="text-emerald-100/50"
                    strokeWidth="8"
                    stroke="currentColor"
                    fill="transparent"
                  />
                  <circle
                    cx="120"
                    cy="120"
                    r="110"
                    className="text-[#138F78] transition-all duration-1000 ease-linear"
                    strokeWidth="8"
                    strokeDasharray={2 * Math.PI * 110}
                    strokeDashoffset={
                      2 * Math.PI * 110 * (1 - Math.min(1, progressPercent / 100))
                    }
                    strokeLinecap="round"
                    stroke="currentColor"
                    fill="transparent"
                  />
                </svg>

                <div className="relative z-10 flex flex-col items-center">
                  <span className="text-4xl sm:text-5xl font-black tracking-tight text-[#123C32] font-mono">
                    {formatTimerDigits(elapsedSeconds)}
                  </span>
                  <span className="text-xs font-medium text-[#526B60] mt-1">
                    Target: {targetMinutes} mnt
                  </span>
                </div>
              </div>

              {/* Preset Selectors (when not running) */}
              {!activeInterval && (
                <div className="w-full space-y-3">
                  <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
                    Pilih Target Durasi
                  </span>
                  <div className="flex flex-wrap justify-center gap-2">
                    {(sessionType === 'focus' ? focusPresets : breakPresets).map((m) => (
                      <button
                        key={m}
                        type="button"
                        onClick={() => setTargetMinutes(m)}
                        className={`px-4 py-2 rounded-xl text-xs font-bold transition-all border ${
                          targetMinutes === m
                            ? 'bg-[#123C32] text-white border-[#123C32]'
                            : 'bg-white text-[#526B60] border-border hover:border-emerald-300'
                        }`}
                      >
                        {m} mnt
                      </button>
                    ))}
                  </div>
                </div>
              )}

              {/* Interruption Logger (during active interval) */}
              {activeInterval && (
                <div className="w-full max-w-sm p-4 rounded-2xl bg-amber-50/70 border border-amber-200/80 flex items-center justify-between">
                  <div className="text-left">
                    <span className="text-xs font-bold text-amber-900 block">
                      Distraksi Tercatat
                    </span>
                    <span className="text-xs text-amber-700">
                      Panggilan telepon, ketukan pintu, atau dorongan membuka medsos
                    </span>
                  </div>
                  <div className="flex items-center gap-3">
                    <span className="text-xl font-black text-amber-900">{interruptions}</span>
                    <button
                      type="button"
                      onClick={handleIncrementInterruption}
                      className="px-3 py-1.5 rounded-xl bg-amber-200 hover:bg-amber-300 text-amber-900 text-xs font-bold transition-colors cursor-pointer active:scale-95"
                    >
                      +1 Catat
                    </button>
                  </div>
                </div>
              )}

              {/* Controls */}
              <div className="flex items-center gap-3 pt-2">
                {!activeInterval ? (
                  <Button
                    variant="primary"
                    size="lg"
                    onClick={handleStartInterval}
                    disabled={submitting}
                    className="px-8 font-bold text-sm bg-[#138F78] hover:bg-[#123C32] text-white rounded-2xl shadow-soft-sm"
                  >
                    Mulai {sessionType === 'focus' ? 'Sesi Fokus' : 'Istirahat'}
                  </Button>
                ) : (
                  <>
                    <Button
                      variant="primary"
                      size="md"
                      onClick={handleFinishInterval}
                      disabled={submitting}
                      className="px-6 font-bold text-xs bg-[#138F78] hover:bg-[#123C32] text-white rounded-xl shadow-soft-sm"
                    >
                      Selesaikan & Simpan Blok
                    </Button>
                    <button
                      type="button"
                      onClick={handleAbandonInterval}
                      disabled={submitting}
                      className="px-4 py-2 rounded-xl text-xs font-bold text-rose-600 hover:bg-rose-50 border border-rose-200 transition-colors"
                    >
                      Batalkan
                    </button>
                  </>
                )}
              </div>
            </div>
          </div>

          {/* Quick Context & Guide (5 cols) */}
          <div className="lg:col-span-5 space-y-6">
            <div className="bg-white rounded-3xl border border-border p-6 shadow-soft-xs space-y-4">
              <h3 className="text-base font-bold text-[#123C32]">Prinsip Fokus Mindora</h3>
              <p className="text-xs leading-relaxed text-[#526B60]">
                Berbeda dengan timer konvensional, Mindora memperlakukan sesi fokus dan istirahat pemulihan sebagai catatan terpisah yang unik.
              </p>
              <div className="space-y-3 pt-2">
                <div className="flex items-start gap-3 p-3 rounded-xl bg-emerald-50/60 border border-emerald-100">
                  <span className="w-5 h-5 rounded-full bg-emerald-600 text-white text-[11px] font-bold flex items-center justify-center shrink-0 mt-0.5">
                    1
                  </span>
                  <div>
                    <span className="text-xs font-bold text-emerald-950 block">Catatan Berkala (Atomic)</span>
                    <span className="text-[11px] text-emerald-800">
                      Blok belajar 25 menit membuat barisnya sendiri. Mengambil istirahat 5 menit membuat catatan istirahat terpisah.
                    </span>
                  </div>
                </div>

                <div className="flex items-start gap-3 p-3 rounded-xl bg-teal-50/60 border border-teal-100">
                  <span className="w-5 h-5 rounded-full bg-teal-600 text-white text-[11px] font-bold flex items-center justify-center shrink-0 mt-0.5">
                    2
                  </span>
                  <div>
                    <span className="text-xs font-bold text-teal-950 block">Kalibrasi Distraksi</span>
                    <span className="text-[11px] text-teal-800">
                      Mencatat distraksi memungkinkan evaluasi kualitas fokus mendalam dibandingkan kepadatan distraksi.
                    </span>
                  </div>
                </div>

                <div className="flex items-start gap-3 p-3 rounded-xl bg-slate-50 border border-slate-200/60">
                  <span className="w-5 h-5 rounded-full bg-slate-600 text-white text-[11px] font-bold flex items-center justify-center shrink-0 mt-0.5">
                    3
                  </span>
                  <div>
                    <span className="text-xs font-bold text-slate-900 block">Waktu Berjalan Terverifikasi Server</span>
                    <span className="text-[11px] text-slate-600">
                      Menutup peramban atau berpindah tab tidak mereset sesi Anda. Waktu berjalan dihitung ulang secara akurat di server.
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Interval History Table */}
        <div className="bg-white rounded-3xl border border-border p-6 shadow-soft-xs space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-bold text-[#123C32]">Riwayat Interval Fokus & Istirahat</h2>
            <span className="text-xs font-medium text-text-secondary">
              Menampilkan {history.length} catatan terakhir
            </span>
          </div>

          {loading ? (
            <div className="py-12 text-center text-sm text-text-muted">Memuat interval...</div>
          ) : history.length === 0 ? (
            <div className="py-12 text-center space-y-2">
              <span className="text-3xl block">⏳</span>
              <p className="text-sm font-semibold text-text-primary">Belum ada interval fokus tercatat</p>
              <p className="text-xs text-text-secondary">
                Pilih target durasi fokus di atas dan mulai interval pertama Anda.
              </p>
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-border/80 text-text-secondary uppercase tracking-wider font-semibold">
                    <th className="py-3 px-3">Jenis</th>
                    <th className="py-3 px-3">Status</th>
                    <th className="py-3 px-3">Durasi (Aktual)</th>
                    <th className="py-3 px-3">Distraksi</th>
                    <th className="py-3 px-3">Mulai</th>
                    <th className="py-3 px-3">Selesai</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/40">
                  {history.map((row) => (
                    <tr key={row.id} className="hover:bg-surface-muted/40 transition-colors">
                      <td className="py-3.5 px-3">
                        <span
                          className={`inline-flex items-center px-2.5 py-1 rounded-full text-[11px] font-bold ${
                            row.session_type === 'focus'
                              ? 'bg-emerald-100 text-[#123C32]'
                              : 'bg-teal-100 text-teal-800'
                          }`}
                        >
                          {row.session_type === 'focus' ? '🎯 Fokus' : '☕ Istirahat'}
                        </span>
                      </td>
                      <td className="py-3.5 px-3">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-bold uppercase ${
                            row.status === 'completed'
                              ? 'bg-emerald-50 text-emerald-700 border border-emerald-200'
                              : row.status === 'active'
                              ? 'bg-amber-50 text-amber-700 border border-amber-200'
                              : 'bg-rose-50 text-rose-700 border border-rose-200'
                          }`}
                        >
                          {translateStatus(row.status)}
                        </span>
                      </td>
                      <td className="py-3.5 px-3 font-medium text-text-primary">
                        {row.duration_minutes > 0 ? `${row.duration_minutes} mnt` : '< 1 mnt'}
                      </td>
                      <td className="py-3.5 px-3 font-semibold text-amber-700">
                        {row.interruption_count}
                      </td>
                      <td className="py-3.5 px-3 text-text-secondary">
                        {formatDateTime(row.started_at)}
                      </td>
                      <td className="py-3.5 px-3 text-text-secondary">
                        {row.ended_at ? formatTime(row.ended_at) : '—'}
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
