import { useState, useEffect } from 'react';
import { useNavigate, useSearchParams, Link } from 'react-router-dom';
import {
  UserIcon,
  ShieldIcon,
  MailIcon,
  LockIcon,
  EyeIcon,
  EyeOffIcon,
  ChevronRightIcon,
  GoogleIcon,
  Button,
  Checkbox,
  TurnstileWidget,
} from '../../components/atoms';
import { FormField } from '../../components/molecules';
import { cn, translateErrorMessage } from '../../utils';
import { useAuth } from '../../hooks';
import { FEATURES } from '../../config/features';

export interface LoginFormProps {
  className?: string;
}

export const LoginForm = ({ className }: LoginFormProps) => {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const { login, loginWithGoogle, isAuthenticated, user } = useAuth();

  const [role, setRole] = useState<'user' | 'admin'>('user');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(true);
  const [turnstileToken, setTurnstileToken] = useState('');
  const [turnstileResetKey, setTurnstileResetKey] = useState(0);
  
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(() => {
    return searchParams.get('session_expired') === 'true'
      ? 'Sesi Anda telah berakhir. Silakan masuk kembali untuk melanjutkan.'
      : null;
  });
  const [infoMessage, setInfoMessage] = useState<string | null>(null);
  const [showForgotModal, setShowForgotModal] = useState(false);

  // If already authenticated, redirect to appropriate workspace
  useEffect(() => {
    if (isAuthenticated && user) {
      if (user.role === 'admin') {
        navigate('/admin');
      } else {
        navigate('/dashboard');
      }
    }
  }, [isAuthenticated, user, navigate]);

  const isEmailValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
  const isPasswordValid = password.length > 0;
  const isCredentialsReady = isEmailValid && isPasswordValid;

  const handleRoleChange = (newRole: 'user' | 'admin') => {
    setRole(newRole);
    setErrorMessage(null);
    setTurnstileToken('');
    setTurnstileResetKey((k) => k + 1);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);
    setInfoMessage(null);

    if (!isEmailValid) {
      setErrorMessage('Silakan masukkan alamat email yang valid');
      return;
    }
    if (!isPasswordValid) {
      setErrorMessage('Silakan masukkan kata sandi Anda');
      return;
    }
    if (!turnstileToken) {
      setErrorMessage('Silakan selesaikan verifikasi keamanan sebelum melanjutkan.');
      return;
    }

    setLoading(true);

    try {
      const result = await login({
        email: email.trim(),
        password,
        turnstile_token: turnstileToken,
        portal: role,
      });

      navigate(result.user.role === 'admin' ? '/admin' : '/dashboard');
    } catch (err: unknown) {
      setErrorMessage(translateErrorMessage(err, 'Gagal masuk. Silakan periksa kredensial Anda dan coba lagi.'));
      setTurnstileToken('');
      setTurnstileResetKey((k) => k + 1);
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleClick = () => {
    if (!FEATURES.ENABLE_GOOGLE_OAUTH) {
      setInfoMessage('Integrasi Google OAuth dinonaktifkan sementara selama tahap pengembangan.');
      return;
    }
    loginWithGoogle(turnstileToken);
  };

  return (
    <div className={cn('w-full max-w-md space-y-6', className)}>
      {/* Form Header */}
      <div className="space-y-2">
        <span className="text-[11px] font-bold text-primary tracking-widest uppercase block">
          Selamat Datang Kembali
        </span>
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-primary">
          Masuk ke Mindora
        </h1>
        <p className="text-sm text-text-secondary font-medium">
          Pilih ruang kerja Anda dan lanjutkan aktivitas Anda.
        </p>
      </div>

      {/* Role Selection Tabs */}
      <div className="grid grid-cols-2 gap-2 p-1 rounded-2xl bg-surface-muted border border-border">
        <button
          type="button"
          onClick={() => handleRoleChange('user')}
          className={cn(
            'flex items-center justify-center gap-2 py-2.5 px-3 rounded-xl text-xs font-bold transition-all cursor-pointer',
            role === 'user'
              ? 'bg-white text-primary-900 shadow-soft-sm border border-border'
              : 'text-text-secondary hover:text-text-primary'
          )}
        >
          <UserIcon className="w-3.5 h-3.5 text-primary" />
          <span>Ruang Kerja Pengguna</span>
        </button>

        <button
          type="button"
          onClick={() => handleRoleChange('admin')}
          className={cn(
            'flex items-center justify-center gap-2 py-2.5 px-3 rounded-xl text-xs font-bold transition-all cursor-pointer',
            role === 'admin'
              ? 'bg-white text-primary-900 shadow-soft-sm border border-border'
              : 'text-text-secondary hover:text-text-primary'
          )}
        >
          <ShieldIcon className="w-3.5 h-3.5 text-primary" />
          <span>Konsol Admin</span>
        </button>
      </div>

      {/* Error / Alert Banner */}
      {errorMessage && (
        <div className="p-3.5 rounded-xl bg-red-50/90 border border-red-200 text-red-700 text-xs font-medium flex items-start gap-2.5 animate-fadeIn">
          <span className="font-bold text-red-800">Kesalahan:</span>
          <span>{errorMessage}</span>
        </div>
      )}

      {/* Info Notice Banner */}
      {infoMessage && (
        <div className="p-3.5 rounded-xl bg-nature-mint border border-primary/20 text-primary-900 text-xs font-medium flex items-start gap-2.5">
          <span className="font-bold">Pemberitahuan:</span>
          <span>{infoMessage}</span>
        </div>
      )}

      {/* Google OAuth Initiation */}
      {FEATURES.ENABLE_GOOGLE_OAUTH && (
        <>
          <div>
            <Button
              type="button"
              variant="outline"
              size="lg"
              onClick={handleGoogleClick}
              disabled={loading}
              className="w-full gap-3 text-xs sm:text-sm font-semibold justify-center hover:bg-surface-muted border-border shadow-soft-sm"
            >
              <GoogleIcon className="w-4 h-4" />
              <span>Lanjutkan dengan Google</span>
            </Button>
          </div>

          {/* Divider */}
          <div className="relative flex items-center justify-center">
            <div className="w-full border-t border-border"></div>
            <span className="absolute bg-surface-card px-3 text-[11px] font-semibold text-text-muted uppercase tracking-wider">
              Atau lanjutkan dengan email
            </span>
          </div>
        </>
      )}

      {/* Form Body */}
      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Email Field */}
        <FormField
          label="Alamat Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder={role === 'admin' ? 'admin@mindora.app' : 'mahasiswa@mindora.app'}
          required
          leftIcon={<MailIcon className="w-4 h-4 text-text-secondary" />}
        />

        {/* Password Field */}
        <FormField
          label="Kata Sandi"
          type={showPassword ? 'text' : 'password'}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Masukkan kata sandi Anda"
          required
          leftIcon={<LockIcon className="w-4 h-4 text-text-secondary" />}
          rightIcon={
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="p-1 text-text-secondary hover:text-text-primary transition-colors cursor-pointer"
              aria-label={showPassword ? 'Sembunyikan kata sandi' : 'Tampilkan kata sandi'}
            >
              {showPassword ? <EyeOffIcon className="w-4 h-4" /> : <EyeIcon className="w-4 h-4" />}
            </button>
          }
        />

        {/* Remember me & Forgot Password */}
        <div className="flex items-center justify-between pt-1">
          <Checkbox
            label="Ingat saya"
            checked={rememberMe}
            onChange={(e) => setRememberMe(e.target.checked)}
          />

          <button
            type="button"
            onClick={() => setShowForgotModal(true)}
            className="text-xs font-semibold text-primary hover:text-primary-hover transition-colors cursor-pointer"
          >
            Lupa kata sandi?
          </button>
        </div>

        {/* Turnstile Security Widget */}
        <TurnstileWidget
          action="login"
          onVerify={(token) => setTurnstileToken(token)}
          resetKey={turnstileResetKey}
        />

        {/* Primary Submit Button */}
        <div className="pt-2">
          <Button
            type="submit"
            variant="primary"
            size="lg"
            disabled={loading || !isCredentialsReady || !turnstileToken}
            className="w-full gap-2 text-sm font-bold justify-center shadow-soft-md"
          >
            <span>{loading ? 'Mengautentikasi...' : `Masuk ke ${role === 'admin' ? 'Konsol Admin' : 'Ruang Kerja'}`}</span>
            <ChevronRightIcon className="w-4 h-4" />
          </Button>
        </div>
      </form>

      {/* Footer Navigation */}
      <div className="text-center space-y-4 pt-2">
        <p className="text-xs text-text-secondary font-medium">
          Belum punya akun Mindora?{' '}
          <Link
            to="/register"
            className="font-bold text-primary hover:text-primary-hover transition-colors"
          >
            Daftar akun baru
          </Link>
        </p>

        <p className="text-[11px] text-text-muted">
          Diamankan dengan autentikasi JWT, rotasi refresh token & Cloudflare Turnstile.
        </p>
      </div>

      {/* Forgot Password Modal */}
      {showForgotModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm animate-fadeIn">
          <div className="bg-surface-card border border-border rounded-2xl p-6 max-w-sm w-full space-y-4 shadow-soft-xl">
            <div className="w-10 h-10 rounded-xl bg-nature-mint text-primary flex items-center justify-center mx-auto">
              <LockIcon className="w-5 h-5" />
            </div>
            <div className="text-center space-y-2">
              <h3 className="text-base font-bold text-text-primary">Pemberitahuan Atur Ulang Kata Sandi</h3>
              <p className="text-xs text-text-secondary leading-relaxed">
                Pemulihan kata sandi mandiri saat ini sedang dalam proses penerapan administratif sesuai kebijakan keamanan institusi.
              </p>
              <p className="text-xs text-text-muted">
                Silakan hubungi administrator kampus Anda atau kirim email ke <span className="font-semibold text-primary">support@mindora.app</span> untuk verifikasi kredensial.
              </p>
            </div>
            <Button
              type="button"
              variant="primary"
              size="md"
              onClick={() => setShowForgotModal(false)}
              className="w-full text-xs font-bold"
            >
              Mengerti
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};

