import { useLocation, Link } from 'react-router-dom';
import { DashboardTemplate } from '../../components/templates';
import type { UserProfile } from '../../types';
import { useAuth } from '../../hooks';

interface ModuleConfig {
  title: string;
  subtitle: string;
  phase: string;
  badge: string;
  description: string;
  expectedFeatures: string[];
  icon: string;
}

const moduleConfigs: Record<string, ModuleConfig> = {
  '/analytics': {
    title: 'Analisis Kognitif & Kebiasaan',
    subtitle: 'Korelasi antara durasi belajar, variabel gaya hidup, dan densitas distraksi.',
    phase: 'Fase 3',
    badge: 'Hadir di Fase 3',
    description:
      'Modul Analisis menyajikan visualisasi statistik mendalam dan korelasi data antara pola tidur, waktu layar, kafein, serta konsistensi belajar.',
    expectedFeatures: [
      'Kurva energi sirkadian vs performa belajar',
      'Korelasi konsumsi kafein vs durasi fokus',
      'Tren frekuensi distraksi berdasarkan waktu belajar',
      'Diagram radar konsistensi kebiasaan mingguan',
    ],
    icon: 'Chart',
  },
  '/reports': {
    title: 'Laporan Produktivitas & Ekspor',
    subtitle: 'Ringkasan PDF komprehensif dan portabilitas data CSV.',
    phase: 'Fase 4',
    badge: 'Hadir di Fase 4',
    description:
      'Hasilkan laporan capaian akademik, verifikasi jam belajar untuk keperluan beasiswa, dan ekspor data pribadi Anda secara aman.',
    expectedFeatures: [
      'Dokumen PDF laporan produktivitas bulanan',
      'Ekspor data mentah CSV untuk analisis penelitian pribadi',
      'Ringkasan mingguan otomatis via email',
      'Sertifikasi bukti waktu belajar terverifikasi',
    ],
    icon: 'Report',
  },
  '/messages': {
    title: 'Grup Belajar & Pesan',
    subtitle: 'Kolaborasi rekan sejawat, diskusi kelompok belajar, dan pesan akademik.',
    phase: 'Fase 4',
    badge: 'Hadir di Fase 4',
    description:
      'Terhubung dengan teman sekelas, bagikan undangan ruang belajar, dan kolaborasi mengerjakan tugas kelompok dengan komunikasi waktu-nyata.',
    expectedFeatures: [
      'Saluran kelompok belajar berbasis mata kuliah',
      'Pesan langsung antarmahasiswa (P2P)',
      'Undangan sesi belajar dan ruang fokus bersama',
      'Lampiran dokumen tugas dan ulasan rekan sebaya',
    ],
    icon: 'Chat',
  },
  '/recommendations': {
    title: 'Rekomendasi Produktivitas AI',
    subtitle: 'Saran pembelajaran mesin berdasarkan kalibrasi kebiasaan pribadi Anda.',
    phase: 'Fase 3',
    badge: 'Hadir di Fase 3',
    description:
      'Setelah data kebiasaan yang memadai terkumpul, mesin rekomendasi Machine Learning akan memberikan saran jam belajar optimal dan penyesuaian waktu tidur.',
    expectedFeatures: [
      'Rekomendasi durasi sesi belajar yang dipersonalisasi',
      'Notifikasi target jam tidur berdasarkan jadwal belajar esok hari',
      'Peringatan jeda waktu layar sebelum blok fokus mendalam',
      'Saran istirahat belajar yang adaptif terhadap konteks',
    ],
    icon: 'AI',
  },
  '/profile': {
    title: 'Profil Mahasiswa & Konteks Akademik',
    subtitle: 'Kelola program studi, target belajar, dan jadwal semester akademik.',
    phase: 'Fase 4',
    badge: 'Hadir di Fase 4',
    description:
      'Kustomisasi rincian mahasiswa, kurikulum akademik, periode ujian, dan preferensi notifikasi Anda.',
    expectedFeatures: [
      'Pengelola jadwal semester akademik & periode ujian',
      'Pemetaan silabus mata kuliah dan bobot SKS',
      'Kustomisasi avatar profil dan identitas mahasiswa',
      'Preferensi notifikasi sesi dan pengingat',
    ],
    icon: 'Education',
  },
  '/settings': {
    title: 'Pengaturan & Keamanan',
    subtitle: 'Otentikasi akun, manajemen sesi, dan preferensi privasi.',
    phase: 'Fase 4',
    badge: 'Hadir di Fase 4',
    description:
      'Konfigurasi pergantian kata sandi, keamanan 2FA, tinjauan sesi perangkat aktif, dan kebijakan privasi data.',
    expectedFeatures: [
      'Otentikasi multi-faktor (MFA / 2FA)',
      'Inspeksi sesi aktif dan keluar jarak jauh',
      'Kustomisasi tema tampilan (Botanical Sage / Forest Night)',
      'Ekspor akun dan kontrol privasi data',
    ],
    icon: 'Settings',
  },
};

export default function PlaceholderModulePage() {
  const { user } = useAuth();
  const location = useLocation();

  const currentPath = location.pathname;
  const config = moduleConfigs[currentPath] || {
    title: 'Modul Produktivitas',
    subtitle: 'Fitur ini sedang dalam tahap pengembangan aktif.',
    phase: 'Fase 3 / 4',
    badge: 'Dalam Pengembangan',
    description: 'Modul ini dijadwalkan untuk dirilis pada pembaruan Mindora berikutnya.',
    expectedFeatures: ['Integrasi fitur mendalam', 'Antarmuka responsif', 'Kemampuan ekspor data'],
    icon: 'Habit',
  };

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

  return (
    <DashboardTemplate role="user" currentUser={displayUser}>
      <div className="max-w-4xl mx-auto py-8 space-y-6">
        {/* Module Header Card */}
        <div className="relative overflow-hidden rounded-3xl bg-linear-to-br from-emerald-50/80 via-white to-teal-50/70 border border-emerald-200/70 p-8 shadow-soft-sm">
          {/* Subtle background decoration */}
          <div className="absolute right-0 bottom-0 pointer-events-none opacity-15">
            <svg width="240" height="200" viewBox="0 0 240 200" fill="none">
              <circle cx="180" cy="120" r="80" fill="#138F78" />
              <path d="M100 20 L150 150 L50 150 Z" fill="#123C32" />
            </svg>
          </div>

          <div className="relative z-10 space-y-4">
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-emerald-100 text-[#123C32] text-xs font-bold border border-emerald-300">
              <span className="w-2 h-2 rounded-full bg-[#138F78] animate-ping" />
              <span>{config.badge}</span>
            </div>

            <div className="flex items-start gap-4">
              <div className="w-14 h-14 rounded-2xl bg-white border border-emerald-200/80 shadow-soft-xs flex items-center justify-center text-3xl shrink-0">
                {config.icon}
              </div>
              <div>
                <h1 className="text-2xl sm:text-3xl font-extrabold text-[#123C32] tracking-tight">
                  {config.title}
                </h1>
                <p className="text-sm text-[#526B60] mt-1">{config.subtitle}</p>
              </div>
            </div>

            <p className="text-sm text-text-secondary leading-relaxed pt-2 max-w-2xl">
              {config.description}
            </p>
          </div>
        </div>

        {/* Planned Capabilities Section */}
        <div className="bg-white rounded-3xl border border-border p-6 sm:p-8 shadow-soft-xs space-y-4">
          <h2 className="text-base font-bold text-[#123C32]">Kemampuan Terencana di {config.phase}</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 pt-1">
            {config.expectedFeatures.map((feat, idx) => (
              <div
                key={idx}
                className="flex items-center gap-3 p-3.5 rounded-2xl bg-slate-50/80 border border-slate-200/60"
              >
                <span className="w-6 h-6 rounded-full bg-emerald-100 text-[#123C32] text-xs font-bold flex items-center justify-center shrink-0">
                  ✓
                </span>
                <span className="text-xs font-semibold text-text-primary">{feat}</span>
              </div>
            ))}
          </div>

          <div className="pt-4 border-t border-border/60 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <span className="text-xs text-text-secondary">
              Pelacakan produktivitas inti (Tugas, Belajar, Fokus, Gaya Hidup) aktif sepenuhnya.
            </span>
            <div className="flex items-center gap-3">
              <Link
                to="/dashboard"
                className="px-4 py-2 rounded-xl text-xs font-bold text-[#123C32] bg-emerald-50 hover:bg-emerald-100 border border-emerald-200 transition-colors"
              >
                ← Kembali ke Dasbor
              </Link>
              <Link
                to="/tasks"
                className="px-4 py-2 rounded-xl text-xs font-bold text-white bg-[#138F78] hover:bg-[#123C32] transition-colors shadow-soft-xs"
              >
                Buka Pengelola Tugas →
              </Link>
            </div>
          </div>
        </div>
      </div>
    </DashboardTemplate>
  );
}

