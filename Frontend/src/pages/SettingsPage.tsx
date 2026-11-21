import { Link } from 'react-router-dom';
import { DashboardTemplate } from '../components/templates';
import { Button } from '../components/atoms';
import {
  SettingsIcon,
  ShieldIcon,
  ClockIcon,
  ExportIcon,
  LogoutIcon,
} from '../components/atoms/Icons';
import type { UserProfile } from '../types';
import { useAuth } from '../hooks';

export default function SettingsPage() {
  const { user, logout } = useAuth();

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
        workspaceName: user.role === 'admin' ? 'Administrasi Sistem' : 'Ruang Kerja Akademik',
      }
    : {
        id: 'usr_guest',
        name: 'Mahasiswa',
        email: 'student@mindora.app',
        role: 'user',
        avatarText: 'MH',
        workspaceName: 'Ruang Kerja Akademik',
      };

  const detectedTimezone = typeof Intl !== 'undefined'
    ? Intl.DateTimeFormat().resolvedOptions().timeZone
    : 'UTC';

  return (
    <DashboardTemplate role={displayUser.role as 'user' | 'admin'} currentUser={displayUser}>
      <div className="space-y-6 pb-12 animate-fadeIn max-w-4xl mx-auto">
        {/* Header Section */}
        <div className="border-b border-border pb-5">
          <div className="flex items-center gap-2">
            <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-bold uppercase tracking-wider bg-nature-mint text-primary">
              Konfigurasi Sistem
            </span>
            <span className="text-xs text-text-muted">Preferensi & Keamanan</span>
          </div>
          <h1 className="text-2xl font-black text-primary-900 tracking-tight mt-1">
            Pengaturan Aplikasi
          </h1>
          <p className="text-sm text-text-secondary">
            Kalibrasi zona waktu, keamanan sesi, dan kontrol portabilitas data.
          </p>
        </div>

        {/* Section 1: System Timezone Alignment */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
          <div className="flex items-center gap-2 border-b border-border pb-3">
            <ClockIcon className="w-5 h-5 text-primary" />
            <div>
              <h2 className="text-base font-bold text-primary-900">Kalibrasi Zona Waktu</h2>
              <p className="text-xs text-text-muted">
                Mengatur batas pergantian hari untuk Pelacakan Belajar, Interval Fokus Pomodoro, dan Laporan Produktivitas.
              </p>
            </div>
          </div>

          <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 text-xs">
            <div>
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Zona Waktu Peramban Terdeteksi (IANA)
              </span>
              <span className="font-mono text-sm font-bold text-primary mt-1 block">
                {detectedTimezone || 'UTC'}
              </span>
              <p className="text-text-secondary text-[11px] mt-0.5">
                Semua metrik harian dan perhitungan tren 7 hari diagregasi relatif terhadap tengah malam di lokasi ini.
              </p>
            </div>
            <span className="inline-flex items-center px-2.5 py-1 rounded-full text-[10px] font-bold bg-emerald-100 text-emerald-800 self-start sm:self-center">
              Tersinkronisasi
            </span>
          </div>
        </div>

        {/* Section 2: Data Portability & Reports */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
          <div className="flex items-center gap-2 border-b border-border pb-3">
            <ExportIcon className="w-5 h-5 text-primary" />
            <div>
              <h2 className="text-base font-bold text-primary-900">Portabilitas Data & Laporan</h2>
              <p className="text-xs text-text-muted">
                Ekspor sesi akademik terverifikasi, catatan fokus, dan catatan kebiasaan Anda.
              </p>
            </div>
          </div>

          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 rounded-xl bg-slate-50 border border-slate-200/80 text-xs">
            <div>
              <strong className="text-text-primary block">Dokumen Produktivitas Komprehensif</strong>
              <p className="text-text-secondary mt-0.5">
                Hasilkan ekspor file CSV terstruktur dan ringkasan PDF yang siap dicetak untuk audit kemajuan akademik.
              </p>
            </div>
            <Link to="/reports" className="shrink-0">
              <Button variant="outline" size="sm" className="text-xs font-bold gap-2">
                <ExportIcon className="w-3.5 h-3.5" />
                Buka Pusat Laporan
              </Button>
            </Link>
          </div>
        </div>

        {/* Section 3: Authentication & Session Security */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
          <div className="flex items-center gap-2 border-b border-border pb-3">
            <ShieldIcon className="w-5 h-5 text-primary" />
            <div>
              <h2 className="text-base font-bold text-primary-900">Keamanan & Akses Sesi</h2>
              <p className="text-xs text-text-muted">
                Status autentikasi kriptografis dan kontrol pengakhiran sesi.
              </p>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 text-xs">
            <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80">
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Arsitektur Token
              </span>
              <span className="text-text-secondary mt-1 block">
                Autentikasi JWT dual-token dengan access token 15 menit dan rotasi sesi terenkripsi melalui HttpOnly cookie.
              </span>
            </div>

            <div className="p-4 rounded-xl bg-slate-50 border border-slate-200/80">
              <span className="font-bold text-text-primary block text-[11px] uppercase tracking-wider">
                Status Sesi Aktif
              </span>
              <div className="flex items-center gap-2 mt-1">
                <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
                <span className="font-bold text-text-primary">Terautentikasi</span>
                <span className="text-[11px] text-text-muted">({displayUser.email})</span>
              </div>
            </div>
          </div>

          <div className="pt-2 flex justify-end">
            <Button
              variant="outline"
              size="sm"
              onClick={logout}
              className="text-xs font-bold text-rose-600 border-rose-200 hover:bg-rose-50 gap-2"
            >
              <LogoutIcon className="w-4 h-4 text-rose-500" />
              Keluar dari Sesi Saat Ini
            </Button>
          </div>
        </div>

        {/* Section 4: Security Roadmap (Transparently Deferred) */}
        <div className="p-6 rounded-2xl bg-white border border-border shadow-soft-xs space-y-4">
          <div className="flex items-center gap-2 border-b border-border pb-3">
            <SettingsIcon className="w-5 h-5 text-primary" />
            <div>
              <h2 className="text-base font-bold text-primary-900">Rencana Keamanan Mendatang</h2>
              <p className="text-xs text-text-muted">
                Peningkatan keamanan yang direncanakan untuk rilis mendatang.
              </p>
            </div>
          </div>

          <div className="space-y-3 text-xs">
            <div className="p-3.5 rounded-xl bg-slate-50 border border-slate-200/80 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
              <div>
                <strong className="text-text-primary block">Autentikasi Multi-Faktor (TOTP 2FA)</strong>
                <p className="text-text-secondary text-[11px] mt-0.5">
                  Verifikasi dua langkah berbasis kunci perangkat keras atau aplikasi autentikator.
                </p>
              </div>
              <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-slate-200 text-slate-700 self-start sm:self-center">
                Rencana Keamanan
              </span>
            </div>

            <div className="p-3.5 rounded-xl bg-slate-50 border border-slate-200/80 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2">
              <div>
                <strong className="text-text-primary block">Ringkasan Aktivitas Email Otomatis</strong>
                <p className="text-text-secondary text-[11px] mt-0.5">
                  Laporan produktivitas mingguan terjadwal yang dikirimkan via email.
                </p>
              </div>
              <span className="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold uppercase tracking-wider bg-slate-200 text-slate-700 self-start sm:self-center">
                Ditangguhkan
              </span>
            </div>
          </div>
        </div>
      </div>
    </DashboardTemplate>
  );
}
