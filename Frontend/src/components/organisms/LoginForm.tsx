import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { UserIcon, ShieldIcon, MailIcon, LockIcon, EyeIcon, EyeOffIcon, ChevronRightIcon, Button } from '../atoms';
import { FormField } from '../molecules';
import { Checkbox } from '../atoms';
import { cn } from '../../utils';

export interface LoginFormProps {
  className?: string;
}

export const LoginForm = ({ className }: LoginFormProps) => {
  const navigate = useNavigate();
  const [role, setRole] = useState<'user' | 'admin'>('user');
  const [email, setEmail] = useState('john@mindora.app');
  const [password, setPassword] = useState('••••••••••••');
  const [showPassword, setShowPassword] = useState(false);
  const [rememberMe, setRememberMe] = useState(true);
  const [loading, setLoading] = useState(false);

  const handleRoleChange = (newRole: 'user' | 'admin') => {
    setRole(newRole);
    if (newRole === 'admin') {
      setEmail('john.admin@mindora.app');
    } else {
      setEmail('john@mindora.app');
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);

    setTimeout(() => {
      setLoading(false);
      if (role === 'admin') {
        navigate('/admin');
      } else {
        navigate('/');
      }
    }, 400);
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

      {/* Form Body */}
      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Email Field */}
        <FormField
          label="Email address"
          type="email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          placeholder="your.email@mindora.app"
          required
          leftIcon={<MailIcon className="w-4 h-4 text-text-secondary" />}
        />

        {/* Password Field */}
        <FormField
          label="Password"
          type={showPassword ? 'text' : 'password'}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Enter password"
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

          <a
            href="#forgot-password"
            className="text-xs font-semibold text-primary hover:text-primary-hover transition-colors"
          >
            Forgot password?
          </a>
        </div>

        {/* Primary Submit Button */}
        <div className="pt-2">
          <Button
            type="submit"
            variant="primary"
            size="lg"
            disabled={loading}
            className="w-full gap-2 text-sm font-bold justify-center shadow-soft-md"
          >
            <span>{loading ? 'Signing in...' : `Sign in to ${role === 'admin' ? 'console' : 'workspace'}`}</span>
            <ChevronRightIcon className="w-4 h-4" />
          </Button>
        </div>
      </form>

      {/* Footer Navigation */}
      <div className="text-center space-y-4 pt-2">
        <p className="text-xs text-text-secondary font-medium">
          New to Mindora?{' '}
          <a
            href="#create-account"
            className="font-bold text-primary hover:text-primary-hover transition-colors"
          >
            Create an account
          </a>
        </p>

        {/* Non-intrusive footnote */}
        <p className="text-[11px] text-text-muted">
          Frontend preview · Authentication will be enabled after the backend is connected.
        </p>
      </div>
    </div>
  );
};
