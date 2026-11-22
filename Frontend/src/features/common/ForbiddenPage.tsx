import { useNavigate } from 'react-router-dom';
import { ShieldIcon, Button, ChevronLeftIcon } from '../../components/atoms';
import { useAuth } from '../../hooks';

export default function ForbiddenPage() {
  const navigate = useNavigate();
  const { user, logout } = useAuth();

  const handleReturn = () => {
    if (user?.role === 'admin') {
      navigate('/admin');
    } else {
      navigate('/dashboard');
    }
  };

  const handleSwitchAccount = async () => {
    await logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-surface-muted flex flex-col items-center justify-center p-6">
      <div className="w-full max-w-md bg-surface-card border border-border rounded-2xl p-8 shadow-soft-lg text-center space-y-6 animate-fadeIn">
        {/* Shield Restriction Badge */}
        <div className="w-16 h-16 rounded-2xl bg-amber-50 text-amber-600 border border-amber-200 flex items-center justify-center mx-auto shadow-soft-sm">
          <ShieldIcon className="w-8 h-8" />
        </div>

        {/* Content */}
        <div className="space-y-2">
          <span className="text-xs font-bold text-amber-600 uppercase tracking-widest block">
            403 â€” Akses Dibatasi
          </span>
          <h1 className="text-2xl font-bold tracking-tight text-text-primary">
            Izin Tidak Mencukupi
          </h1>
          <p className="text-sm text-text-secondary leading-relaxed font-normal">
            Akun Anda saat ini{' '}
            {user?.email && (
              <span className="font-semibold text-text-primary">({user.email})</span>
            )}{' '}
            dengan peran <span className="font-bold text-primary">{user?.role === 'admin' ? 'Administrator' : 'Mahasiswa'}</span> tidak
            memiliki izin administratif yang diperlukan untuk mengakses halaman ini.
          </p>
        </div>

        {/* Action Controls */}
        <div className="pt-2 space-y-3">
          <Button
            type="button"
            variant="primary"
            size="md"
            onClick={handleReturn}
            className="w-full justify-center text-xs font-bold gap-2 shadow-soft-sm"
          >
            <ChevronLeftIcon className="w-4 h-4" />
            <span>Kembali ke {user?.role === 'admin' ? 'Konsol Admin' : 'Ruang Kerja Saya'}</span>
          </Button>

          <Button
            type="button"
            variant="outline"
            size="md"
            onClick={handleSwitchAccount}
            className="w-full justify-center text-xs font-bold text-text-secondary hover:text-text-primary hover:bg-surface-muted"
          >
            Masuk dengan akun lain
          </Button>
        </div>

        {/* Institutional Policy Note */}
        <p className="text-[11px] text-text-muted pt-2 border-t border-border/60">
          Mindora menerapkan kontrol akses berbasis peran (RBAC). Setiap percobaan akses dicatat sesuai dengan kepatuhan kebijakan audit akademik.
        </p>
      </div>
    </div>
  );
}

