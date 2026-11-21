import { Link, useNavigate } from 'react-router-dom';
import { Button, ChevronRightIcon, ShieldIcon, UserIcon } from '../components/atoms';
import { useAuth } from '../hooks';

export default function HomePage() {
  const navigate = useNavigate();
  const { user, isAuthenticated } = useAuth();

  return (
    <div className="min-h-screen bg-surface-muted flex flex-col justify-between selection:bg-primary-100 selection:text-primary-900">
      {/* Top Header Navigation */}
      <header className="w-full bg-white/80 backdrop-blur-md border-b border-border sticky top-0 z-30">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
          <Link to="/" className="flex items-center gap-3">
            <img src="/logo.png" alt="Mindora" className="w-8 h-8 object-contain" />
            <span className="text-xl font-bold tracking-tight text-primary-900">Mindora</span>
          </Link>

          <div className="flex items-center gap-3">
            {isAuthenticated && user ? (
              <div className="flex items-center gap-3">
                <span className="text-xs text-text-secondary hidden sm:inline">
                  Masuk sebagai <strong className="text-text-primary">{user.name}</strong> ({user.role === 'admin' ? 'Administrator' : 'Pengguna'})
                </span>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={() => navigate(user.role === 'admin' ? '/admin' : '/dashboard')}
                  className="text-xs font-bold gap-1.5"
                >
                  <span>Buka {user.role === 'admin' ? 'Konsol Admin' : 'Ruang Kerja'}</span>
                  <ChevronRightIcon className="w-3.5 h-3.5" />
                </Button>
              </div>
            ) : (
              <div className="flex items-center gap-2 sm:gap-3">
                <Link
                  to="/login"
                  className="px-3 py-2 text-xs sm:text-sm font-semibold text-text-secondary hover:text-text-primary transition-colors"
                >
                  Masuk
                </Link>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={() => navigate('/register')}
                  className="text-xs sm:text-sm font-bold gap-1.5 shadow-soft-sm"
                >
                  <span>Mulai Sekarang</span>
                  <ChevronRightIcon className="w-3.5 h-3.5" />
                </Button>
              </div>
            )}
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <main className="flex-1 flex items-center justify-center p-6">
        <div className="max-w-3xl w-full text-center space-y-8 animate-fadeIn">
          {/* Eyebrow Pill */}
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full bg-primary-100 text-primary-900 text-xs font-bold border border-primary/20 shadow-soft-xs">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
            <span>Sistem Produktivitas & Fokus Akademik Mahasiswa</span>
          </div>

          {/* Headline */}
          <div className="space-y-4">
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-black text-text-primary tracking-tight leading-tight">
              Kuasai studi Anda dengan <span className="text-primary">kejernihan & tujuan nyata</span>.
            </h1>
            <p className="text-base sm:text-lg text-text-secondary max-w-xl mx-auto font-normal leading-relaxed">
              Mindora menghadirkan fokus bertema alam, manajemen tugas, dan penjadwalan akademik terstruktur ke dalam satu ruang kerja produktivitas yang aman.
            </p>
          </div>

          {/* Action CTAs */}
          <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-2">
            {isAuthenticated && user ? (
              <Button
                variant="primary"
                size="lg"
                onClick={() => navigate(user.role === 'admin' ? '/admin' : '/dashboard')}
                className="w-full sm:w-auto px-8 gap-2 text-sm font-bold shadow-soft-md"
              >
                <span>Lanjutkan ke {user.role === 'admin' ? 'Konsol Admin' : 'Ruang Kerja'}</span>
                <ChevronRightIcon className="w-4 h-4" />
              </Button>
            ) : (
              <>
                <Button
                  variant="primary"
                  size="lg"
                  onClick={() => navigate('/register')}
                  className="w-full sm:w-auto px-8 gap-2 text-sm font-bold shadow-soft-md"
                >
                  <span>Buat Akun Mahasiswa</span>
                  <ChevronRightIcon className="w-4 h-4" />
                </Button>
                <Button
                  variant="outline"
                  size="lg"
                  onClick={() => navigate('/login')}
                  className="w-full sm:w-auto px-8 gap-2 text-sm font-semibold border-border hover:bg-white"
                >
                  <span>Masuk</span>
                </Button>
              </>
            )}
          </div>

          {/* Feature Highlights Grid */}
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pt-10 text-left">
            <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs space-y-2">
              <div className="w-9 h-9 rounded-xl bg-nature-mint text-primary flex items-center justify-center font-bold">
                <UserIcon className="w-5 h-5" />
              </div>
              <h2 className="text-sm font-bold text-text-primary">Ruang Kerja Privat</h2>
              <p className="text-xs text-text-secondary leading-relaxed">
                Daftar tugas pribadi dan sesi timer fokus diisolasi secara ketat untuk setiap mahasiswa.
              </p>
            </div>

            <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs space-y-2">
              <div className="w-9 h-9 rounded-xl bg-emerald-100 text-emerald-800 flex items-center justify-center font-bold">
                <ShieldIcon className="w-5 h-5" />
              </div>
              <h2 className="text-sm font-bold text-text-primary">RBAC Institusional</h2>
              <p className="text-xs text-text-secondary leading-relaxed">
                Kontrol akses berbasis peran yang melindungi tata kelola administratif dan privasi mahasiswa.
              </p>
            </div>

            <div className="p-5 rounded-2xl bg-white border border-border shadow-soft-xs space-y-2">
              <div className="w-9 h-9 rounded-xl bg-sky-100 text-sky-800 flex items-center justify-center font-bold">
                <span className="text-xs font-mono font-bold">JWT</span>
              </div>
              <h2 className="text-sm font-bold text-text-primary">Autentikasi Cookie HttpOnly</h2>
              <p className="text-xs text-text-secondary leading-relaxed">
                Rotasi refresh token dan manajemen sesi aman yang terlindung dari pencurian token.
              </p>
            </div>
          </div>
        </div>
      </main>

      {/* Footer */}
      <footer className="w-full bg-white border-t border-border py-6 text-center text-xs text-text-muted">
        <p>© 2026 Mindora App. Hak cipta dilindungi undang-undang. Diamankan dengan JWT dan PostgreSQL RBAC.</p>
      </footer>
    </div>
  );
}
