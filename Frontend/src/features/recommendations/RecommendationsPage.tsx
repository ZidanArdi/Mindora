import { useState, useEffect, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { DashboardTemplate } from '../../components/templates';
import {
  SparklesIcon,
  CheckIcon,
  CloseIcon,
  ChevronDownIcon,
  ShieldIcon,
  Button,
} from '../../components/atoms';
import { productivityService } from '../../services/productivityService';
import type {
  RecommendationItem,
  RecommendationPriority,
  UserProfile,
} from '../../types';
import { useAuth } from '../../hooks';
import { formatDate, formatDateTime, translatePriority, translateCategory } from '../../utils/i18n';

export default function RecommendationsPage() {
  const { user } = useAuth();
  const [recommendations, setRecommendations] = useState<RecommendationItem[]>([]);
  const [dataTier, setDataTier] = useState<string>('cold_start');
  const [loading, setLoading] = useState<boolean>(true);
  const [evaluating, setEvaluating] = useState<boolean>(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [actionSuccessMessage, setActionSuccessMessage] = useState<string | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [expandedEvidence, setExpandedEvidence] = useState<Record<string, boolean>>({});

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

  const fetchRecommendations = useCallback(async (shouldEvaluate: boolean = false) => {
    try {
      if (shouldEvaluate) {
        setEvaluating(true);
      } else {
        setLoading(true);
      }
      setErrorMessage(null);

      const res = await productivityService.getRecommendations({ evaluate: shouldEvaluate });
      setRecommendations(res.items || []);
      setDataTier(res.user_data_tier || 'cold_start');
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal memuat rekomendasi';
      setErrorMessage(msg);
    } finally {
      setLoading(false);
      setEvaluating(false);
    }
  }, []);

  useEffect(() => {
    let isMounted = true;
    productivityService
      .getRecommendations({ evaluate: true })
      .then((res) => {
        if (isMounted) {
          setRecommendations(res.items || []);
          setDataTier(res.user_data_tier || 'cold_start');
        }
      })
      .catch((err: unknown) => {
        if (isMounted) {
          const msg = err instanceof Error ? err.message : 'Gagal memuat rekomendasi';
          setErrorMessage(msg);
        }
      })
      .finally(() => {
        if (isMounted) {
          setLoading(false);
          setEvaluating(false);
        }
      });

    return () => {
      isMounted = false;
    };
  }, []);

  const handleStatusUpdate = async (id: string, status: 'dismissed' | 'completed') => {
    try {
      await productivityService.updateRecommendationStatus(id, status);
      // Remove updated recommendation from local state
      setRecommendations((prev) => prev.filter((r) => r.id !== id));
      setActionSuccessMessage(
        status === 'completed'
          ? 'Rekomendasi ditandai selesai.'
          : 'Rekomendasi diabaikan selama 48 jam.'
      );
      setTimeout(() => setActionSuccessMessage(null), 4000);
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : 'Gagal memperbarui status';
      setErrorMessage(msg);
    }
  };

  const toggleEvidence = (id: string) => {
    setExpandedEvidence((prev) => ({
      ...prev,
      [id]: !prev[id],
    }));
  };

  // Categories list derived from current recommendations + defaults
  const categories = [
    { id: 'all', label: 'Semua Rekomendasi' },
    { id: 'Task Prioritization', label: 'Tugas & Prioritas' },
    { id: 'Study Consistency', label: 'Konsistensi Belajar' },
    { id: 'Sleep Hygiene & Rest', label: 'Tidur & Istirahat' },
    { id: 'Break Management & Recovery', label: 'Jeda & Pemulihan' },
    { id: 'Lifestyle Balance', label: 'Keseimbangan Gaya Hidup' },
  ];

  const filteredRecommendations = recommendations.filter((r) => {
    if (selectedCategory === 'all') return true;
    return r.category.toLowerCase().includes(selectedCategory.toLowerCase());
  });

  const getPriorityBadgeClass = (priority: RecommendationPriority) => {
    switch (priority) {
      case 'critical':
        return 'bg-rose-100 text-rose-800 border-rose-300';
      case 'high':
        return 'bg-amber-100 text-amber-900 border-amber-300';
      case 'medium':
        return 'bg-emerald-100 text-[#123C32] border-emerald-300';
      case 'low':
        return 'bg-slate-100 text-slate-700 border-slate-300';
      default:
        return 'bg-slate-100 text-slate-700 border-slate-300';
    }
  };

  const getTierDetails = (tier: string) => {
    switch (tier) {
      case 'longitudinal':
        return {
          title: 'Tingkat 4: Trajektori Longitudinal',
          badge: 'Personalisasi Penuh',
          description:
            'Profil Anda telah memiliki 14+ hari data pelacakan belajar, kebiasaan, dan machine learning yang terverifikasi. Model historis lengkap telah aktif.',
          bg: 'bg-emerald-900 text-white',
        };
      case 'calibrated':
        return {
          title: 'Tingkat 3: Kebiasaan Terkalibrasi',
          badge: 'Kepercayaan Tinggi',
          description:
            'Data dasar terkalibrasi telah terbentuk. Mesin mengevaluasi rata-rata bergulir multi-hari untuk ritme belajar, defisit tidur, dan keseimbangan interval.',
          bg: 'bg-[#123C32] text-white',
        };
      case 'nascent':
        return {
          title: 'Tingkat 2: Aktif Harian',
          badge: 'Pola Mulai Terlihat',
          description:
            'Aktivitas hari ini telah tercatat. Penyelarasan fokus-istirahat dan peringatan tugas aktif telah diaktifkan. Lengkapi kebiasaan harian untuk mencapai status terkalibrasi.',
          bg: 'bg-emerald-800 text-white',
        };
      case 'cold_start':
        return {
          title: 'Tingkat 1A: Pengenalan (Cold Start)',
          badge: 'Panduan Dasar',
          description:
            'Selamat datang di Mindora! Rekomendasi awal difokuskan pada kalibrasi kebiasaan dan tenggat waktu tanpa asumsi perilaku prematur.',
          bg: 'bg-slate-800 text-white',
        };
      case 'dormant':
      default:
        return {
          title: 'Tingkat 1B: Mode Reaktivasi',
          badge: 'Aktivitas Dormant',
          description:
            'Belum ada aktivitas belajar atau pencatatan kebiasaan terkini. Mulai sesi belajar atau jadwalkan tugas kuliah untuk memperbarui rekomendasi.',
          bg: 'bg-slate-800 text-white',
        };
    }
  };

  const tierInfo = getTierDetails(dataTier);

  return (
    <DashboardTemplate role="user" currentUser={displayUser}>
      <div className="space-y-6 max-w-6xl mx-auto">
        {/* Page Header */}
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 bg-white p-6 sm:p-8 rounded-3xl border border-border shadow-soft-xs">
          <div className="space-y-1.5">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-100 text-[#123C32] text-xs font-bold border border-emerald-200">
              <SparklesIcon className="w-3.5 h-3.5 text-[#138F78]" />
              <span>MESIN REKOMENDASI DETERMINISTIK</span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-black text-[#123C32] tracking-tight">
              Rekomendasi Dipersonalisasi
            </h1>
            <p className="text-xs sm:text-sm text-[#526B60] max-w-2xl leading-relaxed">
              Intervensi terarah dan transparan berdasarkan sesi belajar terverifikasi, interval Pomodoro, kebiasaan harian, dan tenggat waktu tugas.
            </p>
          </div>

          <div className="flex items-center gap-3 shrink-0">
            <Button
              variant="outline"
              size="sm"
              onClick={() => fetchRecommendations(true)}
              disabled={evaluating || loading}
              className="text-xs font-bold border-emerald-300 text-[#123C32] hover:bg-emerald-50 cursor-pointer"
            >
              <SparklesIcon className={`w-3.5 h-3.5 mr-1.5 ${evaluating ? 'animate-spin' : ''}`} />
              {evaluating ? 'Mengevaluasi Aturan...' : 'Evaluasi Ulang Rekomendasi'}
            </Button>
          </div>
        </div>

        {/* User Data Tier Indicator Banner */}
        <div className={`rounded-3xl p-6 shadow-soft-xs flex flex-col md:flex-row items-start md:items-center justify-between gap-4 ${tierInfo.bg}`}>
          <div className="space-y-1.5">
            <div className="flex items-center gap-2">
              <span className="text-[10px] font-bold uppercase tracking-wider px-2.5 py-0.5 rounded-full bg-white/20 text-emerald-200">
                {tierInfo.badge}
              </span>
              <span className="text-xs font-semibold text-white/70">Tingkat Personalisasi</span>
            </div>
            <h2 className="text-lg font-bold text-white">{tierInfo.title}</h2>
            <p className="text-xs text-white/80 max-w-3xl leading-relaxed">
              {tierInfo.description}
            </p>
          </div>

          <div className="shrink-0 flex items-center gap-2 text-xs text-white/70 bg-white/10 px-3.5 py-2 rounded-2xl border border-white/10">
            <ShieldIcon className="w-4 h-4 text-emerald-300" />
            <span>Integritas Nol vs NULL Terlindungi</span>
          </div>
        </div>

        {/* Success or Error Feedback Banners */}
        {actionSuccessMessage && (
          <div className="p-4 rounded-2xl bg-emerald-50 border border-emerald-200 text-xs font-semibold text-[#123C32] flex items-center gap-2 animate-fadeIn">
            <CheckIcon className="w-4 h-4 text-[#138F78]" />
            <span>{actionSuccessMessage}</span>
          </div>
        )}

        {errorMessage && (
          <div className="p-4 rounded-2xl bg-rose-50 border border-rose-200 text-xs font-semibold text-rose-800 flex items-center justify-between gap-2 animate-fadeIn">
            <span>{errorMessage}</span>
            <Button
              variant="outline"
              size="sm"
              onClick={() => fetchRecommendations(true)}
              className="text-[11px] font-bold text-rose-800 border-rose-300 hover:bg-rose-100"
            >
              Coba Lagi
            </Button>
          </div>
        )}

        {/* Category Filters */}
        <div className="flex items-center gap-2 overflow-x-auto pb-1 scrollbar-none">
          {categories.map((cat) => (
            <button
              key={cat.id}
              type="button"
              onClick={() => setSelectedCategory(cat.id)}
              className={`px-4 py-2 rounded-xl text-xs font-bold transition-all cursor-pointer whitespace-nowrap ${
                selectedCategory === cat.id
                  ? 'bg-[#123C32] text-white shadow-soft-xs'
                  : 'bg-white text-[#526B60] hover:bg-slate-50 border border-border'
              }`}
            >
              {cat.label}
            </button>
          ))}
        </div>

        {/* Main Content Area */}
        {loading ? (
          <div className="space-y-4">
            {[1, 2].map((i) => (
              <div
                key={i}
                className="bg-white rounded-3xl border border-border p-6 shadow-soft-xs space-y-4 animate-pulse"
              >
                <div className="flex items-center justify-between">
                  <div className="h-5 w-32 bg-slate-200 rounded-full" />
                  <div className="h-5 w-24 bg-slate-200 rounded-full" />
                </div>
                <div className="h-6 w-3/4 bg-slate-200 rounded-lg" />
                <div className="h-14 w-full bg-slate-100 rounded-xl" />
              </div>
            ))}
          </div>
        ) : filteredRecommendations.length === 0 ? (
          <div className="bg-white rounded-3xl border border-dashed border-border p-12 text-center space-y-4 shadow-soft-xs">
            <div className="w-14 h-14 rounded-2xl bg-emerald-50 text-[#138F78] flex items-center justify-center mx-auto text-2xl font-bold">
              ✓
            </div>
            <div className="space-y-1">
              <h3 className="text-base font-bold text-[#123C32]">
                Tidak Ada Rekomendasi Tertunda
              </h3>
              <p className="text-xs text-[#526B60] max-w-md mx-auto leading-relaxed">
                {selectedCategory === 'all'
                  ? 'Semua target Anda berjalan optimal! Tidak ditemukan tugas mendesak yang terlewat, defisit tidur kronis, atau ketidakseimbangan ritme belajar.'
                  : `Tidak ada rekomendasi aktif dalam kategori "${selectedCategory}".`}
              </p>
            </div>
            <div className="pt-2">
              <Button
                variant="outline"
                size="sm"
                onClick={() => fetchRecommendations(true)}
                className="text-xs font-bold border-emerald-300 text-[#123C32] hover:bg-emerald-50"
              >
                Evaluasi Ulang Aturan
              </Button>
            </div>
          </div>
        ) : (
          <div className="space-y-4">
            {filteredRecommendations.map((rec) => {
              const isExpanded = !!expandedEvidence[rec.id];
              const evidenceKeys = Object.keys(rec.evidence || {});

              return (
                <div
                  key={rec.id}
                  className="bg-white rounded-3xl border border-border hover:border-emerald-300/60 p-6 sm:p-7 shadow-soft-xs transition-all space-y-5"
                >
                  {/* Top Badges */}
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="flex flex-wrap items-center gap-2">
                      <span
                        className={`px-3 py-1 rounded-full text-[10px] font-black uppercase tracking-wider border ${getPriorityBadgeClass(
                          rec.priority
                        )}`}
                      >
                        Prioritas {translatePriority(rec.priority)}
                      </span>
                      <span className="px-2.5 py-0.5 rounded-lg bg-slate-100 text-slate-700 text-[11px] font-semibold border border-slate-200">
                        {translateCategory(rec.category)}
                      </span>
                      <span className="text-[10px] font-mono text-slate-400">
                        {rec.rule_id}
                      </span>
                    </div>

                    <span className="text-[11px] text-[#526B60]">
                      Berlaku hingga:{' '}
                      {formatDateTime(rec.valid_until)}
                    </span>
                  </div>

                  {/* Title and Message */}
                  <div className="space-y-1.5">
                    <h3 className="text-lg font-bold text-[#123C32] tracking-tight">
                      {rec.title}
                    </h3>
                    <p className="text-sm text-[#123C32] font-medium leading-relaxed">
                      {rec.message}
                    </p>
                  </div>

                  {/* Explanation Callout */}
                  <div className="p-4 rounded-2xl bg-[#E8F5F1]/70 border border-emerald-100 text-xs text-[#123C32] leading-relaxed">
                    <span className="font-bold block text-[11px] uppercase tracking-wider text-[#138F78] mb-1">
                      Mengapa rekomendasi ini penting
                    </span>
                    {rec.explanation}
                  </div>

                  {/* Evidence Disclosure Accordion */}
                  <div className="border border-slate-200 rounded-2xl overflow-hidden bg-slate-50/50">
                    <button
                      type="button"
                      onClick={() => toggleEvidence(rec.id)}
                      className="w-full px-4 py-3 flex items-center justify-between text-left text-xs font-bold text-[#123C32] hover:bg-slate-100/70 transition-colors cursor-pointer"
                    >
                      <div className="flex items-center gap-2">
                        <ShieldIcon className="w-4 h-4 text-[#138F78]" />
                        <span>Bukti & Parameter Terverifikasi</span>
                        <span className="text-[10px] font-normal text-slate-500">
                          ({evidenceKeys.length} parameter terverifikasi)
                        </span>
                      </div>
                      <ChevronDownIcon
                        className={`w-4 h-4 text-slate-400 transition-transform duration-200 ${
                          isExpanded ? 'rotate-180 text-[#138F78]' : ''
                        }`}
                      />
                    </button>

                    {isExpanded && (
                      <div className="p-4 border-t border-slate-200 bg-white text-xs space-y-3">
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                          {evidenceKeys.map((key) => {
                            const val = rec.evidence[key];
                            const formattedKey = key
                              .replace(/_/g, ' ')
                              .replace(/\b\w/g, (c) => c.toUpperCase());
                            const formattedVal =
                              typeof val === 'object' && val !== null
                                ? JSON.stringify(val)
                                : String(val);

                            return (
                              <div
                                key={key}
                                className="p-2.5 rounded-xl bg-slate-50 border border-slate-100 flex flex-col justify-between"
                              >
                                <span className="text-[10px] font-semibold text-slate-500">
                                  {formattedKey}
                                </span>
                                <span className="text-xs font-mono font-bold text-[#123C32] mt-0.5 break-all">
                                  {formattedVal}
                                </span>
                              </div>
                            );
                          })}
                        </div>
                        <p className="text-[10px] text-slate-400 leading-relaxed border-t border-slate-100 pt-2">
                          Periode evaluasi:{' '}
                          {formatDate(rec.evaluation_period_start)} –{' '}
                          {formatDate(rec.evaluation_period_end)} (Kunci:{' '}
                          <code className="font-mono text-[9px] bg-slate-100 px-1 py-0.5 rounded">
                            {rec.evaluation_key}
                          </code>
                          )
                        </p>
                      </div>
                    )}
                  </div>

                  {/* Actions Footer */}
                  <div className="flex flex-wrap items-center justify-between gap-3 pt-2 border-t border-slate-100">
                    <div className="flex items-center gap-2">
                      {rec.action_route && (
                        <Link
                          to={rec.action_route}
                          className="inline-flex items-center px-4 py-2 rounded-xl text-xs font-bold text-white bg-[#138F78] hover:bg-[#123C32] transition-colors shadow-soft-xs"
                        >
                          {rec.action_label || 'Ambil Tindakan'}
                        </Link>
                      )}
                    </div>

                    <div className="flex items-center gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleStatusUpdate(rec.id, 'completed')}
                        className="text-xs font-bold text-emerald-800 border-emerald-300 hover:bg-emerald-50 cursor-pointer"
                      >
                        <CheckIcon className="w-3.5 h-3.5 mr-1 text-[#138F78]" />
                        Tandai Selesai
                      </Button>

                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleStatusUpdate(rec.id, 'dismissed')}
                        className="text-xs font-semibold text-slate-500 hover:text-rose-700 hover:bg-rose-50 cursor-pointer"
                      >
                        <CloseIcon className="w-3.5 h-3.5 mr-1" />
                        Abaikan (48 Jam)
                      </Button>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </DashboardTemplate>
  );
}

