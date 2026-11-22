import { Link } from 'react-router-dom';
import { DashboardTemplate } from '../../components/templates';
import { Button } from '../../components/atoms';
import {
  MessageSquareIcon,
  SparklesIcon,
  ClockIcon,
  ShieldIcon,
  CheckIcon,
} from '../../components/atoms/Icons';
import type { UserProfile } from '../../types';
import { useAuth } from '../../hooks';

export default function MessagesPage() {
  const { user } = useAuth();

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

  return (
    <DashboardTemplate role={displayUser.role as 'user' | 'admin'} currentUser={displayUser}>
      <div className="space-y-6 pb-12 animate-fadeIn max-w-7xl mx-auto">
        {/* Header Section */}
        <div className="flex flex-col md:flex-row md:items-center md:justify-between gap-4 border-b border-border pb-5">
          <div>
            <div className="flex items-center gap-2">
              <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-amber-100 text-amber-800 border border-amber-200">
                Fase 4.3 Â· Layanan LLM Ditangguhkan
              </span>
              <span className="text-xs text-text-muted">Komunikasi Akademik & Panduan AI</span>
            </div>
            <h1 className="text-2xl font-black text-primary-900 tracking-tight mt-1">
              Pesan & Percakapan
            </h1>
            <p className="text-sm text-text-secondary">
              Panduan produktivitas, saluran komunikasi, dan status bantuan percakapan.
            </p>
          </div>

          <div className="flex items-center gap-2">
            <Link to="/recommendations">
              <Button
                variant="primary"
                size="sm"
                className="gap-2 text-xs font-semibold bg-primary hover:bg-primary-hover text-white"
              >
                <SparklesIcon className="w-4 h-4 text-emerald-300" />
                Lihat Rekomendasi Aktif
              </Button>
            </Link>
          </div>
        </div>

        {/* Technical Status & Architecture Transparency Card */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
          <div className="flex items-start gap-4">
            <div className="w-12 h-12 rounded-2xl bg-amber-50 border border-amber-200 flex items-center justify-center shrink-0 text-amber-700">
              <ShieldIcon className="w-6 h-6" />
            </div>
            <div className="space-y-1.5 flex-1">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <h2 className="text-base font-bold text-primary-900">
                  Mesin Perencanaan AI Percakapan â€” Status Ditangguhkan
                </h2>
                <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[11px] font-bold bg-amber-100 text-amber-900">
                  <span className="w-2 h-2 rounded-full bg-amber-600 animate-pulse" />
                  Layanan Tidak Terhubung
                </span>
              </div>
              <p className="text-xs text-text-secondary leading-relaxed">
                Endpoint asisten LLM percakapan (semula direncanakan sebagai{' '}
                <code className="px-1.5 py-0.5 rounded bg-slate-100 font-mono text-[11px] text-text-primary">
                  POST /api/v1/ai/plan-day
                </code>
                ) sengaja ditangguhkan pada rilis ini. Mindora menganut kebijakan rekayasa yang mengutamakan privasi dan operasi luring (offline-first): tidak ada API LLM cloud eksternal (OpenAI, Gemini, Anthropic) atau chatbot generatif yang tidak terverifikasi aktif, memastikan data aktivitas mahasiswa tetap sepenuhnya rahasia dan terlindungi dari halusinasi kecerdasan buatan.
              </p>
            </div>
          </div>

          <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
            <div>
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Standar Privasi
              </span>
              <p className="text-text-secondary mt-0.5">
                Nol transmisi LLM pihak ketiga; log belajar pribadi tetap tersimpan secara lokal/on-premise.
              </p>
            </div>
            <div>
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Alternatif Aktif
              </span>
              <p className="text-text-secondary mt-0.5">
                Mesin aturan deterministik terverifikasi (REC-TASK, REC-SLEEP, REC-BREAK, REC-COLD, REC-STUDY).
              </p>
            </div>
            <div>
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Kebijakan AI Tanpa Manipulasi
              </span>
              <p className="text-text-secondary mt-0.5">
                Nol respon palsu atau chatbot tiruan; seluruh data produktivitas yang ditampilkan 100% terverifikasi.
              </p>
            </div>
          </div>
        </div>

        {/* Two-Column Grid: Active Productivity Hub vs Future Peer Collaboration Roadmap */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Column 1: Active Guidance Tools (Working Today) */}
          <div className="p-6 rounded-2xl bg-gradient-to-br from-[#123C32] to-[#0A261F] text-white shadow-soft-md flex flex-col justify-between space-y-6">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-emerald-400/20 text-emerald-300 border border-emerald-400/30">
                  <SparklesIcon className="w-3 h-3 text-emerald-400" />
                  Siap Produksi (Fase 4.1)
                </span>
                <span className="text-[11px] font-mono text-white/60">Mesin Deterministik</span>
              </div>

              <div>
                <h3 className="text-lg font-bold text-white">
                  Mesin Rekomendasi Deterministik
                </h3>
                <p className="text-xs text-white/70 mt-1 leading-relaxed">
                  Alih-alih chatbot probabilistik yang rentan halusinasi, Mindora mengevaluasi kebiasaan terverifikasi Anda terhadap aturan matematis yang ketat untuk menghasilkan rekomendasi terpersonalisasi:
                </p>
              </div>

              <div className="space-y-2.5 text-xs text-white/90">
                <div className="flex items-start gap-2.5 p-2.5 rounded-xl bg-white/5 border border-white/10">
                  <CheckIcon className="w-4 h-4 text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong className="text-emerald-300">Intervensi Tugas Terlewat:</strong> Peringatan waktu nyata untuk tenggat waktu prioritas tinggi.
                  </div>
                </div>

                <div className="flex items-start gap-2.5 p-2.5 rounded-xl bg-white/5 border border-white/10">
                  <CheckIcon className="w-4 h-4 text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong className="text-emerald-300">Pengingat Kalibrasi Kebiasaan:</strong> Panduan awal saat log tidur atau waktu layar harian belum tercatat.
                  </div>
                </div>

                <div className="flex items-start gap-2.5 p-2.5 rounded-xl bg-white/5 border border-white/10">
                  <CheckIcon className="w-4 h-4 text-emerald-400 mt-0.5 shrink-0" />
                  <div>
                    <strong className="text-emerald-300">Penyelarasan Fokus-Istirahat:</strong> Peringatan pemulihan kelelahan Pomodoro saat fokus melebihi 100 menit tanpa istirahat memadai.
                  </div>
                </div>
              </div>
            </div>

            <div className="pt-4 border-t border-white/10">
              <Link to="/recommendations" className="block">
                <Button
                  variant="primary"
                  size="md"
                  className="w-full justify-center bg-emerald-500 hover:bg-emerald-600 text-slate-900 font-bold border-none"
                >
                  Jelajahi Pusat Rekomendasi
                </Button>
              </Link>
            </div>
          </div>

          {/* Column 2: Peer Collaboration & Study Groups Roadmap */}
          <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs flex flex-col justify-between space-y-6">
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-slate-100 text-text-muted">
                  <MessageSquareIcon className="w-3 h-3 text-text-muted" />
                  Tahapan Mendatang
                </span>
                <span className="text-[11px] font-mono text-text-muted">Saluran Kolaborasi</span>
              </div>

              <div>
                <h3 className="text-lg font-bold text-primary-900">
                  Kelompok Belajar & Kolaborasi Rekan
                </h3>
                <p className="text-xs text-text-secondary mt-1 leading-relaxed">
                  Arsitektur perpesanan antar-mahasiswa disiapkan untuk rilis kolaboratif multipengguna. Kemampuan yang direncanakan meliputi:
                </p>
              </div>

              <div className="space-y-2.5 text-xs text-text-secondary">
                <div className="flex items-start gap-2.5 p-3 rounded-xl bg-slate-50 border border-slate-200/80">
                  <span className="w-5 h-5 rounded-full bg-primary-100 text-primary font-bold text-[10px] flex items-center justify-center shrink-0">
                    1
                  </span>
                  <div>
                    <strong className="text-text-primary block">Saluran Belajar Berbasis Mata Kuliah</strong>
                    Diskusi dan berbagi materi yang diorganisasi per mata kuliah akademik.
                  </div>
                </div>

                <div className="flex items-start gap-2.5 p-3 rounded-xl bg-slate-50 border border-slate-200/80">
                  <span className="w-5 h-5 rounded-full bg-primary-100 text-primary font-bold text-[10px] flex items-center justify-center shrink-0">
                    2
                  </span>
                  <div>
                    <strong className="text-text-primary block">Ruang Fokus Pomodoro Bersama</strong>
                    Interval fokus tersinkronisasi dan timer akuntabilitas bersama rekan belajar.
                  </div>
                </div>

                <div className="flex items-start gap-2.5 p-3 rounded-xl bg-slate-50 border border-slate-200/80">
                  <span className="w-5 h-5 rounded-full bg-primary-100 text-primary font-bold text-[10px] flex items-center justify-center shrink-0">
                    3
                  </span>
                  <div>
                    <strong className="text-text-primary block">Konsultasi Akademik Langsung</strong>
                    Pesan privat antara mahasiswa dan asisten dosen pengampu.
                  </div>
                </div>
              </div>
            </div>

            <div className="p-3.5 rounded-xl bg-slate-50 border border-slate-200 text-xs text-text-muted flex items-center justify-between">
              <span>Status: Direncanakan untuk Kerangka Kolaboratif</span>
              <span className="font-semibold text-text-secondary">Luring</span>
            </div>
          </div>
        </div>

        {/* Empty / Deferred Channel State Notice */}
        <div className="p-10 rounded-2xl bg-white border border-border shadow-soft-xs text-center space-y-3">
          <div className="w-12 h-12 rounded-2xl bg-slate-100 text-text-muted flex items-center justify-center mx-auto">
            <MessageSquareIcon className="w-6 h-6" />
          </div>
          <h3 className="text-sm font-bold text-text-primary">
            Tidak Ada Percakapan Pesan Aktif
          </h3>
          <p className="text-xs text-text-muted max-w-md mx-auto leading-relaxed">
            Saluran perpesanan dinonaktifkan selama layanan LLM percakapan eksternal ditangguhkan. Untuk merencanakan hari dan mengoptimalkan beban akademik Anda, gunakan modul produktivitas Mindora yang terverifikasi.
          </p>
          <div className="pt-2 flex flex-wrap items-center justify-center gap-3">
            <Link to="/study">
              <Button variant="secondary" size="sm" className="text-xs font-bold gap-1.5">
                <ClockIcon className="w-3.5 h-3.5" />
                Pelacakan Belajar
              </Button>
            </Link>
            <Link to="/focus">
              <Button variant="outline" size="sm" className="text-xs font-bold">
                Timer Fokus
              </Button>
            </Link>
            <Link to="/reports">
              <Button variant="outline" size="sm" className="text-xs font-bold">
                Laporan Produktivitas
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </DashboardTemplate>
  );
}

