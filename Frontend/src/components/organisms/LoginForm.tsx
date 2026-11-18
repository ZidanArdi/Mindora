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
} from '../atoms';
import { FormField } from '../molecules';
import { cn } from '../../utils';
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
  const [turnstileToken, setTurnstileToken] = useState('1x0000000000000000000000000000000AA');
  
  const [loading, setLoading] = useState(false);
  const [errorMessage, setErrorMessage] = useState<string | null>(() => {
    return searchParams.get('session_expired') === 'true'
      ? 'Your session has expired. Please log in again to continue.'
      : null;
  });
  const [infoMessage, setInfoMessage] = useState<string | null>(null);
  const [showForgotModal, setShowForgotModal] = useState(false);

  // If already authenticated, redirect to appropriate workspace
  useEffect(() => {
    if (isAuthenticated && user) {
      if (user.role === 'admin' && role === 'admin') {
        navigate('/admin');
      } else {
        navigate('/dashboard');
      }
    }
  }, [isAuthenticated, user, navigate, role]);

  const handleRoleChange = (newRole: 'user' | 'admin') => {
    setRole(newRole);
    setErrorMessage(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);
    setInfoMessage(null);

    if (!email.trim()) {
      setErrorMessage('Please enter your email address');
      return;
    }
    if (!password) {
      setErrorMessage('Please enter your password');
      return;
    }

    setLoading(true);

    try {
      const result = await login({
        email: email.trim(),
        password,
        turnstile_token: turnstileToken,
      });

      // Role check and routing
      if (role === 'admin') {
        if (result.user.role === 'admin') {
          navigate('/admin');
        } else {
          // User tried to log into admin console with user credentials
          setInfoMessage('Your account has student permissions. Routing to user workspace...');
          setTimeout(() => navigate('/dashboard'), 800);
        }
      } else {
        navigate('/dashboard');
      }
    } catch (err: unknown) {
      if (err instanceof Error) {
        if (err.message.includes('Failed to fetch') || err.message.includes('NetworkError')) {
          setErrorMessage('Backend server is unreachable at http://localhost:8080. Please ensure the backend is running.');
        } else {
          setErrorMessage(err.message);
        }
      } else {
        setErrorMessage('Failed to sign in. Please verify your credentials and try again.');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleGoogleClick = () => {
    if (!FEATURES.ENABLE_GOOGLE_OAUTH) {
      setInfoMessage('Google OAuth integration is temporarily disabled during development.');
      return;
    }
    loginWithGoogle(turnstileToken);
  };

  return (
    <div className={cn('w-full max-w-md space-y-6', className)}>
      {/* Form Header */}
      <div className="space-y-2">
        <span className="text-[11px] font-bold text-primary tracking-widest uppercase block">
          Welcome Back
        </span>
        <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-primary">
          Sign in to Mindora
        </h1>
        <p className="text-sm text-text-secondary font-medium">
          Choose your workspace and continue where you left off.
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
          <span>User workspace</span>
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
          <span>Admin console</span>
        </button>
      </div>

      {/* Error / Alert Banner */}
      {errorMessage && (
        <div className="p-3.5 rounded-xl bg-red-50/90 border border-red-200 text-red-700 text-xs font-medium flex items-start gap-2.5 animate-fadeIn">
          <span className="font-bold text-red-800">Error:</span>
          <span>{errorMessage}</span>
        </div>
      )}

      {/* Info Notice Banner */}
      {infoMessage && (
        <div className="p-3.5 rounded-xl bg-nature-mint border border-primary/20 text-primary-900 text-xs font-medium flex items-start gap-2.5">
          <span className="font-bold">Notice:</span>
          <span>{infoMessage}</span>
        </div>
      )}

      {/* Google OAuth Initiation (Explicit Feature Flag Controlled) */}
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
              <span>Continue with Google</span>
            </Button>
          </div>

          {/* Divider */}
          <div className="relative flex items-center justify-center">
            <div className="w-full border-t border-border"></div>
            <span className="absolute bg-surface-card px-3 text-[11px] font-semibold text-text-muted uppercase tracking-wider">
              Or continue with email
            </span>
          </div>
        </>
      )}

      {/* Form Body */}
      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Email Field */}
        <FormField
          label="Email address"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder={role === 'admin' ? 'admin@mindora.app' : 'student@mindora.app'}
          required
          leftIcon={<MailIcon className="w-4 h-4 text-text-secondary" />}
        />

        {/* Password Field */}
        <FormField
          label="Password"
          type={showPassword ? 'text' : 'password'}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Enter your password"
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

        {/* Remember me & Forgot Password */}
        <div className="flex items-center justify-between pt-1">
          <Checkbox
            label="Remember me"
            checked={rememberMe}
            onChange={(e) => setRememberMe(e.target.checked)}
          />

          <button
            type="button"
            onClick={() => setShowForgotModal(true)}
            className="text-xs font-semibold text-primary hover:text-primary-hover transition-colors cursor-pointer"
          >
            Forgot password?
          </button>
        </div>

        {/* Turnstile Security Widget */}
        <TurnstileWidget
          action="login"
          onVerify={(token) => setTurnstileToken(token)}
        />

        {/* Primary Submit Button */}
        <div className="pt-2">
          <Button
            type="submit"
            variant="primary"
            size="lg"
            disabled={loading}
            className="w-full gap-2 text-sm font-bold justify-center shadow-soft-md"
          >
            <span>{loading ? 'Authenticating...' : `Sign in to ${role === 'admin' ? 'console' : 'workspace'}`}</span>
            <ChevronRightIcon className="w-4 h-4" />
          </Button>
        </div>
      </form>

      {/* Footer Navigation */}
      <div className="text-center space-y-4 pt-2">
        <p className="text-xs text-text-secondary font-medium">
          New to Mindora?{' '}
          <Link
            to="/register"
            className="font-bold text-primary hover:text-primary-hover transition-colors"
          >
            Create an account
          </Link>
        </p>

        <p className="text-[11px] text-text-muted">
          Secured with JWT authentication, refresh token rotation & Cloudflare Turnstile.
        </p>
      </div>

      {/* Forgot Password Modal (Pending Backend Implementation) */}
      {showForgotModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/40 backdrop-blur-sm animate-fadeIn">
          <div className="bg-surface-card border border-border rounded-2xl p-6 max-w-sm w-full space-y-4 shadow-soft-xl">
            <div className="w-10 h-10 rounded-xl bg-nature-mint text-primary flex items-center justify-center mx-auto">
              <LockIcon className="w-5 h-5" />
            </div>
            <div className="text-center space-y-2">
              <h3 className="text-base font-bold text-text-primary">Password Reset Notice</h3>
              <p className="text-xs text-text-secondary leading-relaxed">
                Self-service password recovery is currently pending administrative rollout in accordance with institutional security policies.
              </p>
              <p className="text-xs text-text-muted">
                Please contact your campus administrator or write to <span className="font-semibold text-primary">support@mindora.app</span> to request credential verification.
              </p>
            </div>
            <Button
              type="button"
              variant="primary"
              size="md"
              onClick={() => setShowForgotModal(false)}
              className="w-full text-xs font-bold"
            >
              Understood
            </Button>
          </div>
        </div>
      )}
    </div>
  );
};
