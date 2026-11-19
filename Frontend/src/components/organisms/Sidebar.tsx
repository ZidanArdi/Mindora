import { Link, useNavigate, useLocation } from 'react-router-dom';
import type { NavCategory, WorkspaceItem, UserProfile } from '../../types';
import { useAuth } from '../../hooks';
import {
  GridIcon,
  ListIcon,
  ClockIcon,
  SparklesIcon,
  ChartBarIcon,
  UserIcon,
  SettingsIcon,
  ExportIcon,
  UsersIcon,
  FolderIcon,
  ShieldIcon,
  LogoutIcon,
  ChevronLeftIcon,
  CloseIcon,
  Avatar,
} from '../atoms';
import { cn } from '../../utils';

export interface SidebarProps {
  role: 'user' | 'admin';
  currentUser: UserProfile;
  myLists?: NavCategory[];
  workspaces?: WorkspaceItem[];
  isOpen?: boolean;
  onClose?: () => void;
  className?: string;
}

export const Sidebar = ({
  role,
  currentUser,
  myLists: _myLists = [],
  workspaces: _workspaces = [],
  isOpen,
  onClose,
  className,
}: SidebarProps) => {
  const navigate = useNavigate();
  const location = useLocation();
  const { logout } = useAuth();

  const handleLogout = async () => {
    try {
      await logout();
    } finally {
      navigate('/login');
    }
  };

  const isItemActive = (href: string) => {
    if (href === '/dashboard') {
      return location.pathname === '/dashboard';
    }
    return location.pathname.startsWith(href);
  };

  const userCoreNavItems = [
    { label: 'Dashboard', icon: <GridIcon className="w-4 h-4" />, href: '/dashboard' },
    { label: 'Tasks', icon: <ListIcon className="w-4 h-4" />, href: '/tasks' },
    { label: 'Study Tracking', icon: <ClockIcon className="w-4 h-4" />, href: '/study' },
    {
      label: 'Focus',
      icon: (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-4 h-4">
          <circle cx="12" cy="12" r="10" />
          <circle cx="12" cy="12" r="6" />
          <circle cx="12" cy="12" r="2" />
        </svg>
      ),
      href: '/focus',
    },
    {
      label: 'Lifestyle Tracking',
      icon: (
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round" strokeLinejoin="round" className="w-4 h-4">
          <path d="M12 2a10 10 0 1 0 10 10c0-4.42-3-8-7-9.5" />
          <path d="M12 2c2 3 2 7 0 10" />
        </svg>
      ),
      href: '/lifestyle',
    },
  ];

  const userInsightsNavItems = [
    { label: 'Analytics', icon: <ChartBarIcon className="w-4 h-4" />, href: '/analytics' },
    { label: 'Reports', icon: <ExportIcon className="w-4 h-4" />, href: '/reports', tag: 'Phase 4' },
    { label: 'Recommendations', icon: <SparklesIcon className="w-4 h-4" />, href: '/recommendations', tag: 'Phase 4' },
  ];

  const userSettingsNavItems = [
    { label: 'Profile', icon: <UserIcon className="w-4 h-4" />, href: '/profile' },
    { label: 'Settings', icon: <SettingsIcon className="w-4 h-4" />, href: '/settings' },
  ];

  const adminNavItemsWorkspace = [
    { label: 'Overview', icon: <GridIcon className="w-4 h-4" />, href: '/admin' },
    { label: 'Members', icon: <UsersIcon className="w-4 h-4" />, href: '/admin#members' },
    { label: 'Workspaces', icon: <FolderIcon className="w-4 h-4" />, href: '/admin#workspaces' },
    { label: 'Reports', icon: <ChartBarIcon className="w-4 h-4" />, href: '/admin#reports' },
  ];

  const adminNavItemsManagement = [
    { label: 'Roles & access', icon: <ShieldIcon className="w-4 h-4" />, href: '/admin#roles' },
    { label: 'Settings', icon: <SettingsIcon className="w-4 h-4" />, href: '/admin#settings' },
  ];

  return (
    <>
      {/* Mobile backdrop */}
      {isOpen && (
        <div
          className="fixed inset-0 bg-slate-900/40 z-40 lg:hidden backdrop-blur-xs transition-opacity"
          onClick={onClose}
        />
      )}

      <aside
        className={cn(
          'fixed lg:sticky top-0 left-0 z-50 h-screen w-64 bg-white border-r border-border flex flex-col justify-between p-5 transition-transform duration-200 select-none overflow-y-auto no-scrollbar',
          isOpen ? 'translate-x-0 shadow-soft-lg' : '-translate-x-full lg:translate-x-0',
          className
        )}
      >
        <div className="space-y-6">
          {/* Logo Header */}
          <div className="flex items-center justify-between">
            <Link to={role === 'user' ? '/dashboard' : '/admin'} className="flex items-center gap-2.5">
              <img src="/logo.png" alt="Mindora Logo" className="w-8 h-8 object-contain" />
              <div>
                <span className="text-lg font-bold tracking-tight text-primary-900 block leading-tight">
                  Mindora
                </span>
                {role === 'admin' && (
                  <span className="text-[10px] font-bold text-primary tracking-widest uppercase block">
                    Admin Console
                  </span>
                )}
              </div>
            </Link>

            {onClose && (
              <button
                type="button"
                onClick={onClose}
                className="p-1 rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-muted lg:hidden"
                aria-label="Close menu"
              >
                <CloseIcon className="w-5 h-5" />
              </button>
            )}
          </div>

          {/* User Mode Navigation */}
          {role === 'user' ? (
            <div className="space-y-5">
              {/* Core Productivity Modules */}
              <div className="space-y-1">
                <span className="px-3 text-[11px] font-bold text-text-secondary uppercase tracking-wider block">
                  Productivity
                </span>
                <nav className="space-y-0.5 pt-1">
                  {userCoreNavItems.map((item) => {
                    const active = isItemActive(item.href);
                    return (
                      <Link
                        key={item.label}
                        to={item.href}
                        onClick={onClose}
                        className={cn(
                          'flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all',
                          active
                            ? 'bg-emerald-50 text-[#123C32] font-bold shadow-soft-xs border border-emerald-200/60'
                            : 'text-text-secondary hover:text-[#123C32] hover:bg-surface-muted'
                        )}
                      >
                        <span className={cn(active ? 'text-[#138F78]' : 'text-text-secondary')}>
                          {item.icon}
                        </span>
                        <span>{item.label}</span>
                      </Link>
                    );
                  })}
                </nav>
              </div>

              {/* Insights & Future Modules */}
              <div className="space-y-1">
                <span className="px-3 text-[11px] font-bold text-text-secondary uppercase tracking-wider block">
                  Insights & AI
                </span>
                <nav className="space-y-0.5 pt-1">
                  {userInsightsNavItems.map((item) => {
                    const active = isItemActive(item.href);
                    return (
                      <Link
                        key={item.label}
                        to={item.href}
                        onClick={onClose}
                        className={cn(
                          'flex items-center justify-between px-3 py-2 rounded-xl text-sm font-medium transition-all',
                          active
                            ? 'bg-emerald-50 text-[#123C32] font-bold shadow-soft-xs border border-emerald-200/60'
                            : 'text-text-secondary hover:text-[#123C32] hover:bg-surface-muted'
                        )}
                      >
                        <div className="flex items-center gap-3">
                          <span className={cn(active ? 'text-[#138F78]' : 'text-text-secondary')}>
                            {item.icon}
                          </span>
                          <span>{item.label}</span>
                        </div>
                        {item.tag && (
                          <span className="text-[10px] font-bold text-teal-800 bg-teal-50 px-1.5 py-0.5 rounded border border-teal-200">
                            {item.tag}
                          </span>
                        )}
                      </Link>
                    );
                  })}
                </nav>
              </div>

              {/* Preferences & Settings */}
              <div className="space-y-1">
                <span className="px-3 text-[11px] font-bold text-text-secondary uppercase tracking-wider block">
                  Settings
                </span>
                <nav className="space-y-0.5 pt-1">
                  {userSettingsNavItems.map((item) => {
                    const active = isItemActive(item.href);
                    return (
                      <Link
                        key={item.label}
                        to={item.href}
                        onClick={onClose}
                        className={cn(
                          'flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all',
                          active
                            ? 'bg-emerald-50 text-[#123C32] font-bold shadow-soft-xs border border-emerald-200/60'
                            : 'text-text-secondary hover:text-[#123C32] hover:bg-surface-muted'
                        )}
                      >
                        <span className={cn(active ? 'text-[#138F78]' : 'text-text-secondary')}>
                          {item.icon}
                        </span>
                        <span>{item.label}</span>
                      </Link>
                    );
                  })}
                </nav>
              </div>

              {/* Quick switch to Admin demo */}
              <div className="pt-2 border-t border-border-subtle">
                <Link
                  to="/admin"
                  className="flex items-center gap-2.5 px-3 py-2 rounded-xl text-xs font-semibold text-emerald-800 bg-emerald-50 hover:bg-emerald-100 transition-colors border border-emerald-200"
                >
                  <ShieldIcon className="w-4 h-4 text-emerald-600" />
                  <span>Switch to Admin Console</span>
                </Link>
              </div>
            </div>
          ) : (
            /* Admin Mode Navigation */
            <div className="space-y-6">
              {/* WORKSPACE group */}
              <div className="space-y-2">
                <span className="px-3 text-xs font-semibold text-text-secondary uppercase tracking-wider block">
                  Workspace
                </span>
                <nav className="space-y-1">
                  {adminNavItemsWorkspace.map((item) => {
                    const active = isItemActive(item.href);
                    return (
                      <Link
                        key={item.label}
                        to={item.href}
                        onClick={onClose}
                        className={cn(
                          'flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all',
                          active
                            ? 'bg-primary-100 text-primary-900 font-semibold shadow-soft-sm'
                            : 'text-text-secondary hover:text-text-primary hover:bg-surface-muted'
                        )}
                      >
                        <span className={cn(active ? 'text-primary' : 'text-text-secondary')}>
                          {item.icon}
                        </span>
                        <span>{item.label}</span>
                      </Link>
                    );
                  })}
                </nav>
              </div>

              {/* MANAGEMENT group */}
              <div className="space-y-2">
                <span className="px-3 text-xs font-semibold text-text-secondary uppercase tracking-wider block">
                  Management
                </span>
                <nav className="space-y-1">
                  {adminNavItemsManagement.map((item) => {
                    const active = isItemActive(item.href);
                    return (
                      <Link
                        key={item.label}
                        to={item.href}
                        onClick={onClose}
                        className={cn(
                          'flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all',
                          active
                            ? 'bg-primary-100 text-primary-900 font-semibold shadow-soft-sm'
                            : 'text-text-secondary hover:text-text-primary hover:bg-surface-muted'
                        )}
                      >
                        <span className={cn(active ? 'text-primary' : 'text-text-secondary')}>
                          {item.icon}
                        </span>
                        <span>{item.label}</span>
                      </Link>
                    );
                  })}
                </nav>
              </div>

              {/* Back to main app button */}
              <div className="pt-2 border-t border-border-subtle">
                <Link
                  to="/dashboard"
                  className="flex items-center gap-2 px-3 py-2 rounded-xl text-xs font-semibold text-text-secondary hover:text-primary hover:bg-surface-muted transition-colors border border-border"
                >
                  <ChevronLeftIcon className="w-4 h-4" />
                  <span>Back to workspace</span>
                </Link>
              </div>
            </div>
          )}
        </div>

        {/* User Footer Profile */}
        <div className="pt-4 border-t border-border mt-4">
          <div className="flex items-center justify-between p-2 rounded-xl hover:bg-surface-muted transition-colors">
            <div className="flex items-center gap-2.5 min-w-0">
              <Avatar initials={currentUser.avatarText} size="sm" />
              <div className="min-w-0">
                <p className="text-xs font-semibold text-text-primary truncate leading-tight">
                  {currentUser.name}
                </p>
                <p className="text-[10px] text-text-secondary truncate">
                  {role === 'admin' ? 'Super admin' : currentUser.email}
                </p>
              </div>
            </div>
            <button
              type="button"
              onClick={handleLogout}
              title="Logout"
              className="p-1 text-text-secondary hover:text-red-600 transition-colors cursor-pointer"
              aria-label="Logout"
            >
              <LogoutIcon className="w-4 h-4" />
            </button>
          </div>
        </div>
      </aside>
    </>
  );
};
