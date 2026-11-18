import { useNavigate } from 'react-router-dom';
import { ShieldIcon, Button, ChevronLeftIcon } from '../components/atoms';
import { useAuth } from '../hooks';

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
            403 — Access Restricted
          </span>
          <h1 className="text-2xl font-bold tracking-tight text-text-primary">
            Insufficient Permissions
          </h1>
          <p className="text-sm text-text-secondary leading-relaxed font-normal">
            Your current account{' '}
            {user?.email && (
              <span className="font-semibold text-text-primary">({user.email})</span>
            )}{' '}
            with role <span className="font-bold text-primary capitalize">{user?.role || 'user'}</span> does
            not possess the required administrative clearance to view this resource.
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
            <span>Return to {user?.role === 'admin' ? 'Admin Console' : 'My Workspace'}</span>
          </Button>

          <Button
            type="button"
            variant="outline"
            size="md"
            onClick={handleSwitchAccount}
            className="w-full justify-center text-xs font-bold text-text-secondary hover:text-text-primary hover:bg-surface-muted"
          >
            Sign in with a different account
          </Button>
        </div>

        {/* Institutional Policy Note */}
        <p className="text-[11px] text-text-muted pt-2 border-t border-border/60">
          Mindora enforces institutional role-based access control (RBAC). Access attempts are logged in compliance with audit policy.
        </p>
      </div>
    </div>
  );
}
