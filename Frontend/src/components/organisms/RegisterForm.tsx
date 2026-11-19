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
import { cn } from '../../utils';
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
      setErrorMessage('Full name is required');
      return;
    }
    if (!email.trim()) {
      setErrorMessage('Email address is required');
      return;
    }
    if (!isPasswordValid) {
      setErrorMessage('Please ensure your password satisfies all security requirements');
      return;
    }
    if (password !== confirmPassword) {
      setErrorMessage('Passwords do not match');
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
      if (err instanceof Error) {
        if (err.message.includes('Failed to fetch') || err.message.includes('NetworkError')) {
          setErrorMessage('Backend server is unreachable at http://localhost:8080. Please ensure the backend is running.');
        } else {
          setErrorMessage(err.message);
        }
      } else {
        setErrorMessage('Failed to create account. Please check your information and try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleClick = () => {
    if (!FEATURES.ENABLE_GOOGLE_OAUTH) {
      setErrorMessage('Google OAuth registration is temporarily disabled during development.');
      return;
    }
    loginWithGoogle(turnstileToken);
  };

  return (
    <div className={cn('w-full max-w-md space-y-5', className)}>
      {/* Form Header */}
      <div className="space-y-1.5">
        <span className="text-[11px] font-bold text-primary tracking-widest uppercase block">
          Get Started
        </span>
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-primary">
          Create your account
        </h1>
        <p className="text-sm text-text-secondary font-medium">
          Start managing your academic productivity with Mindora.
        </p>
      </div>

      {/* Error Banner */}
      {errorMessage && (
        <div className="p-3.5 rounded-xl bg-red-50/90 border border-red-200 text-red-700 text-xs font-medium flex items-start gap-2.5 animate-fadeIn">
          <span className="font-bold text-red-800">Registration Error:</span>
          <span>{errorMessage}</span>
        </div>
      )}

      {/* Google Sign-Up Button (Explicit Feature Flag Controlled) */}
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
              <span>Sign up with Google</span>
            </Button>
          </div>

          {/* Divider */}
          <div className="relative flex items-center justify-center">
            <div className="w-full border-t border-border"></div>
            <span className="absolute bg-surface-card px-3 text-[11px] font-semibold text-text-muted uppercase tracking-wider">
              Or register with email
            </span>
          </div>
        </>
      )}

      {/* Form Body */}
      <form onSubmit={handleSubmit} className="space-y-3.5">
        {/* Full Name */}
        <FormField
          label="Full Name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g. John Doe"
          required
          leftIcon={<UserIcon className="w-4 h-4 text-text-secondary" />}
        />

        {/* Email Address */}
        <FormField
          label="Email address"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="student@mindora.app"
          required
          leftIcon={<MailIcon className="w-4 h-4 text-text-secondary" />}
        />

        {/* Password */}
        <FormField
          label="Password"
          type={showPassword ? 'text' : 'password'}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Create a strong password"
          required
          leftIcon={<LockIcon className="w-4 h-4 text-text-secondary" />}
          rightIcon={
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="p-1 text-text-secondary hover:text-text-primary transition-colors cursor-pointer"
              aria-label={showPassword ? 'Hide password' : 'Show password'}
            >
              {showPassword ? <EyeOffIcon className="w-4 h-4" /> : <EyeIcon className="w-4 h-4" />}
            </button>
          }
        />

        {/* Confirm Password */}
        <FormField
          label="Confirm Password"
          type={showConfirmPassword ? 'text' : 'password'}
          value={confirmPassword}
          onChange={(e) => setConfirmPassword(e.target.value)}
          placeholder="Confirm your password"
          required
          leftIcon={<LockIcon className="w-4 h-4 text-text-secondary" />}
          rightIcon={
            <button
              type="button"
              onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              className="p-1 text-text-secondary hover:text-text-primary transition-colors cursor-pointer"
              aria-label={showConfirmPassword ? 'Hide password' : 'Show password'}
            >
              {showConfirmPassword ? <EyeOffIcon className="w-4 h-4" /> : <EyeIcon className="w-4 h-4" />}
            </button>
          }
        />

        {/* Real-time Password Strength Checklist */}
        {password.length > 0 && (
          <div className="p-3 rounded-xl bg-surface-muted border border-border text-[11px] space-y-1.5 transition-all">
            <p className="font-semibold text-text-secondary">Password Requirements:</p>
            <div className="grid grid-cols-2 gap-x-2 gap-y-1 text-text-muted">
              <span className={cn('flex items-center gap-1.5', hasMinLen ? 'text-primary font-medium' : '')}>
                <CheckIcon className={cn('w-3 h-3', hasMinLen ? 'text-primary' : 'text-text-muted/50')} />
                8+ characters
              </span>
              <span className={cn('flex items-center gap-1.5', hasUpper ? 'text-primary font-medium' : '')}>
                <CheckIcon className={cn('w-3 h-3', hasUpper ? 'text-primary' : 'text-text-muted/50')} />
                Uppercase letter
              </span>
              <span className={cn('flex items-center gap-1.5', hasLower ? 'text-primary font-medium' : '')}>
                <CheckIcon className={cn('w-3 h-3', hasLower ? 'text-primary' : 'text-text-muted/50')} />
                Lowercase letter
              </span>
              <span className={cn('flex items-center gap-1.5', hasNumberOrSymbol ? 'text-primary font-medium' : '')}>
                <CheckIcon className={cn('w-3 h-3', hasNumberOrSymbol ? 'text-primary' : 'text-text-muted/50')} />
                Number or symbol
              </span>
            </div>
            {confirmPassword.length > 0 && (
              <div className="pt-1 border-t border-border/60">
                <span className={cn('flex items-center gap-1.5', isMatch ? 'text-primary font-medium' : 'text-red-500 font-medium')}>
                  <CheckIcon className={cn('w-3 h-3', isMatch ? 'text-primary' : 'text-red-400')} />
                  {isMatch ? 'Passwords match' : 'Passwords do not match'}
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
            <span>{loading ? 'Creating Account...' : 'Create Account'}</span>
            <ChevronRightIcon className="w-4 h-4" />
          </Button>
        </div>
      </form>

      {/* Footer Navigation */}
      <div className="text-center space-y-2 pt-1">
        <p className="text-xs text-text-secondary font-medium">
          Already have an account?{' '}
          <Link
            to="/login"
            className="font-bold text-primary hover:text-primary-hover transition-colors"
          >
            Sign in
          </Link>
        </p>

        <p className="text-[11px] text-text-muted">
          By signing up, you agree to Mindora's Terms of Service and Privacy Policy.
        </p>
      </div>
    </div>
  );
};
