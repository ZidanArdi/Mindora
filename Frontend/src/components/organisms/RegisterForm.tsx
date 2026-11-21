import { useState, useEffect } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import {
  UserIcon,
  MailIcon,
  LockIcon,
  EyeIcon,
  EyeOffIcon,
  ChevronRightIcon,
  GoogleIcon,
  CheckIcon,
  Button,
  TurnstileWidget,
} from '../atoms';
import { FormField } from '../molecules';
import { cn, translateErrorMessage } from '../../utils';
import { useAuth } from '../../hooks';
import { FEATURES } from '../../config/features';

export interface RegisterFormProps {
  className?: string;
}

export const RegisterForm = ({ className }: RegisterFormProps) => {
  const navigate = useNavigate();
  const { register, loginWithGoogle, isAuthenticated, user } = useAuth();

  // Redirect authenticated user to workspace
  useEffect(() => {
    if (isAuthenticated && user) {
      navigate(user.role === 'admin' ? '/admin' : '/dashboard', { replace: true });
    }
  }, [isAuthenticated, user, navigate]);

  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [turnstileToken, setTurnstileToken] = useState('1x0000000000000000000000000000000AA');

  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Password requirement tests (matches backend validator)
  const hasMinLen = password.length >= 8;
  const hasUpper = /[A-Z]/.test(password);
  const hasLower = /[a-z]/.test(password);
  const hasNumberOrSymbol = /[\d!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?`~]/.test(password);
  const isPasswordValid = hasMinLen && hasUpper && hasLower && hasNumberOrSymbol;
  const isMatch = password === confirmPassword && confirmPassword.length > 0;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);

    if (!name.trim()) {
      setErrorMessage('Nama lengkap wajib diisi');
      return;
    }
    if (!email.trim()) {
      setErrorMessage('Alamat email wajib diisi');
      return;
    }
    if (!isPasswordValid) {
      setErrorMessage('Pastikan kata sandi Anda memenuhi semua persyaratan keamanan');
      return;
    }
    if (password !== confirmPassword) {
      setErrorMessage('Kata sandi tidak cocok');
      return;
    }

    setLoading(true);

    try {
      await register({
        name: name.trim(),
        email: email.trim(),
        password,
        turnstile_token: turnstileToken,
      });

      // Public registration strictly routes to student user workspace
      navigate('/dashboard');
    } catch (err: unknown) {
      setErrorMessage(translateErrorMessage(err, 'Gagal membuat akun. Silakan periksa data Anda dan coba lagi.'));
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleClick = () => {
    if (!FEATURES.ENABLE_GOOGLE_OAUTH) {
      setErrorMessage('Pendaftaran Google OAuth dinonaktifkan sementara selama tahap pengembangan.');
      return;
    }
    loginWithGoogle(turnstileToken);
  };

  return (
    <div className={cn('w-full max-w-md space-y-5', className)}>
      {/* Form Header */}
      <div className="space-y-1.5">
        <span className="text-[11px] font-bold text-primary tracking-widest uppercase block">
          Mulai Sekarang
        </span>
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-primary">
          Buat Akun Anda
        </h1>
        <p className="text-sm text-text-secondary font-medium">
          Mulai kelola produktivitas akademik Anda bersama Mindora.
        </p>
      </div>

      {/* Error Banner */}
      {errorMessage && (
        <div className="p-3.5 rounded-xl bg-red-50/90 border border-red-200 text-red-700 text-xs font-medium flex items-start gap-2.5 animate-fadeIn">
          <span className="font-bold text-red-800">Kesalahan:</span>
          <span>{errorMessage}</span>
        </div>
      )}

      {/* Google Sign-Up Button */}
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
              <span>Daftar dengan Google</span>
            </Button>
          </div>

          {/* Divider */}
          <div className="relative flex items-center justify-center">
            <div className="w-full border-t border-border"></div>
            <span className="absolute bg-surface-card px-3 text-[11px] font-semibold text-text-muted uppercase tracking-wider">
              Atau daftar dengan email
            </span>
          </div>
        </>
      )}

      {/* Form Body */}
      <form onSubmit={handleSubmit} className="space-y-3.5">
        {/* Full Name */}
        <FormField
          label="Nama Lengkap"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="cth. Budi Santoso"
          required
          leftIcon={<UserIcon className="w-4 h-4 text-text-secondary" />}
        />

        {/* Email Address */}
        <FormField
          label="Alamat Email"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="mahasiswa@mindora.app"
          required
          leftIcon={<MailIcon className="w-4 h-4 text-text-secondary" />}
        />

        {/* Password */}
        <FormField
          label="Kata Sandi"
          type={showPassword ? 'text' : 'password'}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Buat kata sandi yang kuat"
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

        {/* Confirm Password */}
        <FormField
          label="Konfirmasi Kata Sandi"
          type={showConfirmPassword ? 'text' : 'password'}
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          placeholder="Konfirmasi kata sandi Anda"
          required
          leftIcon={<LockIcon className="w-4 h-4 text-text-secondary" />}
          rightIcon={
            <button
              type="button"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              className="p-1 text-text-secondary hover:text-text-primary transition-colors cursor-pointer"
              aria-label={showConfirmPassword ? 'Sembunyikan kata sandi' : 'Tampilkan kata sandi'}
            >
              {showConfirmPassword ? <EyeOffIcon className="w-4 h-4" /> : <EyeIcon className="w-4 h-4" />}
            </button>
          }
        />

        {/* Real-time Password Strength Checklist */}
        {password.length > 0 && (
          <div className="p-3 rounded-xl bg-surface-muted border border-border text-[11px] space-y-1.5 transition-all">
            <p className="font-semibold text-text-secondary">Persyaratan Kata Sandi:</p>
            <div className="grid grid-cols-2 gap-x-2 gap-y-1 text-text-muted">
              <span className={cn('flex items-center gap-1.5', hasMinLen ? 'text-primary font-medium' : '')}>
                <CheckIcon className={cn('w-3 h-3', hasMinLen ? 'text-primary' : 'text-text-muted/50')} />
                Minimal 8 karakter
              </span>
              <span className={cn('flex items-center gap-1.5', hasUpper ? 'text-primary font-medium' : '')}>
                <CheckIcon className={cn('w-3 h-3', hasUpper ? 'text-primary' : 'text-text-muted/50')} />
                Huruf besar (A-Z)
              </span>
              <span className={cn('flex items-center gap-1.5', hasLower ? 'text-primary font-medium' : '')}>
                <CheckIcon className={cn('w-3 h-3', hasLower ? 'text-primary' : 'text-text-muted/50')} />
                Huruf kecil (a-z)
              </span>
              <span className={cn('flex items-center gap-1.5', hasNumberOrSymbol ? 'text-primary font-medium' : '')}>
                <CheckIcon className={cn('w-3 h-3', hasNumberOrSymbol ? 'text-primary' : 'text-text-muted/50')} />
                Angka atau simbol
              </span>
            </div>
            {confirmPassword.length > 0 && (
              <div className="pt-1 border-t border-border/60">
                <span className={cn('flex items-center gap-1.5', isMatch ? 'text-primary font-medium' : 'text-red-500 font-medium')}>
                  <CheckIcon className={cn('w-3 h-3', isMatch ? 'text-primary' : 'text-red-400')} />
                  {isMatch ? 'Kata sandi cocok' : 'Kata sandi tidak cocok'}
                </span>
              </div>
            )}
          </div>
        )}

        {/* Turnstile Security Widget */}
        <TurnstileWidget
          action="register"
          onVerify={(token) => setTurnstileToken(token)}
        />

        {/* Submit Button */}
        <div className="pt-1">
          <Button
            type="submit"
            variant="primary"
            size="lg"
            disabled={loading || (password.length > 0 && (!isPasswordValid || !isMatch))}
            className="w-full gap-2 text-sm font-bold justify-center shadow-soft-md"
          >
            <span>{loading ? 'Membuat Akun...' : 'Buat Akun'}</span>
            <ChevronRightIcon className="w-4 h-4" />
          </Button>
        </div>
      </form>

      {/* Footer Navigation */}
      <div className="text-center space-y-2 pt-1">
        <p className="text-xs text-text-secondary font-medium">
          Sudah memiliki akun?{' '}
          <Link
            to="/login"
            className="font-bold text-primary hover:text-primary-hover transition-colors"
          >
            Masuk
          </Link>
        </p>

        <p className="text-[11px] text-text-muted">
          Dengan mendaftar, Anda menyetujui Ketentuan Layanan dan Kebijakan Privasi Mindora.
        </p>
      </div>
    </div>
  );
};
