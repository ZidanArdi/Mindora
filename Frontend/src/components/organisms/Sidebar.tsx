import { useState } from 'react';
import { Link } from 'react-router-dom';
import type { NavCategory, WorkspaceItem, UserProfile } from '../../types';
import {
  GridIcon,
  ListIcon,
  CalendarIcon,
  MessageSquareIcon,
  PaletteIcon,
  HashIcon,
  PlusIcon,
  UsersIcon,
  FolderIcon,
  ChartBarIcon,
  ShieldIcon,
  SettingsIcon,
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
  myLists = [],
  workspaces = [],
  isOpen,
  onClose,
  className,
}: SidebarProps) => {
  const [activeItem, setActiveItem] = useState(
    role === 'user' ? 'Dashboard' : 'Overview'
  );

  const userNavItems = [
    { label: 'Dashboard', icon: <GridIcon className="w-4 h-4" />, href: '/' },
    { label: 'List', icon: <ListIcon className="w-4 h-4" />, href: '/#list' },
    { label: 'Calendar', icon: <CalendarIcon className="w-4 h-4" />, href: '/#calendar' },
    { label: 'Messages', icon: <MessageSquareIcon className="w-4 h-4" />, href: '/#messages' },
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
            <Link to={role === 'user' ? '/' : '/admin'} className="flex items-center gap-2.5">
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
            <div className="space-y-6">
              {/* Main Nav */}
              <nav className="space-y-1">
                {userNavItems.map((item) => {
                  const isActive = activeItem === item.label;
                  return (
                    <Link
                      key={item.label}
                      to={item.href}
                      onClick={() => setActiveItem(item.label)}
                      className={cn(
                        'flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all',
                        isActive
                          ? 'bg-primary-100 text-primary-900 font-semibold shadow-soft-sm'
                          : 'text-text-secondary hover:text-text-primary hover:bg-surface-muted'
                      )}
                    >
                      <span className={cn(isActive ? 'text-primary' : 'text-text-secondary')}>
                        {item.icon}
                      </span>
                      <span>{item.label}</span>
                    </Link>
                  );
                })}
              </nav>

              {/* My List Section */}
              <div className="space-y-2">
                <div className="flex items-center justify-between px-3 text-xs font-semibold text-text-secondary uppercase tracking-wider">
                  <span>My list</span>
                  <button type="button" className="text-text-secondary hover:text-text-primary p-0.5" aria-label="Add list">
                    <PlusIcon className="w-3.5 h-3.5" />
                  </button>
                </div>
                <div className="space-y-1">
                  {myLists.map((cat) => (
                    <button
                      key={cat.id}
                      type="button"
                      className="w-full flex items-center justify-between px-3 py-1.5 rounded-xl text-sm text-text-secondary hover:text-text-primary hover:bg-surface-muted transition-colors cursor-pointer"
                    >
                      <div className="flex items-center gap-3">
                        <span className="w-2 h-2 rounded-full shrink-0" style={{ backgroundColor: cat.color }} />
                        <span className="font-medium text-sm">{cat.name}</span>
                      </div>
                      {cat.count !== undefined && (
                        <span className="text-[11px] font-semibold text-text-muted bg-surface-muted px-2 py-0.5 rounded-full">
                          {cat.count}
                        </span>
                      )}
                    </button>
                  ))}
                </div>
              </div>

              {/* Workspace Section */}
              <div className="space-y-2">
                <div className="flex items-center justify-between px-3 text-xs font-semibold text-text-secondary uppercase tracking-wider">
                  <span>Workspace</span>
                  <button type="button" className="text-text-secondary hover:text-text-primary p-0.5" aria-label="Add workspace">
                    <PlusIcon className="w-3.5 h-3.5" />
                  </button>
                </div>
                <div className="space-y-1">
                  {workspaces.map((ws) => (
                    <button
                      key={ws.id}
                      type="button"
                      className="w-full flex items-center gap-3 px-3 py-1.5 rounded-xl text-sm text-text-secondary hover:text-text-primary hover:bg-surface-muted transition-colors cursor-pointer"
                    >
                      <span className="text-text-secondary">
                        {ws.iconType === 'palette' ? (
                          <PaletteIcon className="w-4 h-4 text-emerald-600" />
                        ) : (
                          <HashIcon className="w-4 h-4 text-teal-600" />
                        )}
                      </span>
                      <span className="font-medium text-sm">{ws.name}</span>
                    </button>
                  ))}
                </div>
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
                    const isActive = activeItem === item.label;
                    return (
                      <Link
                        key={item.label}
                        to={item.href}
                        onClick={() => setActiveItem(item.label)}
                        className={cn(
                          'flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all',
                          isActive
                            ? 'bg-primary-100 text-primary-900 font-semibold shadow-soft-sm'
                            : 'text-text-secondary hover:text-text-primary hover:bg-surface-muted'
                        )}
                      >
                        <span className={cn(isActive ? 'text-primary' : 'text-text-secondary')}>
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
                    const isActive = activeItem === item.label;
                    return (
                      <Link
                        key={item.label}
                        to={item.href}
                        onClick={() => setActiveItem(item.label)}
                        className={cn(
                          'flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all',
                          isActive
                            ? 'bg-primary-100 text-primary-900 font-semibold shadow-soft-sm'
                            : 'text-text-secondary hover:text-text-primary hover:bg-surface-muted'
                        )}
                      >
                        <span className={cn(isActive ? 'text-primary' : 'text-text-secondary')}>
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
                  to="/"
                  className="flex items-center gap-2 px-3 py-2 rounded-xl text-xs font-semibold text-text-secondary hover:text-primary hover:bg-surface-muted transition-colors border border-border"
                >
                  <ChevronLeftIcon className="w-4 h-4" />
                  <span>Back to main app</span>
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
            <Link to="/login" title="Logout" className="p-1 text-text-secondary hover:text-red-600 transition-colors">
              <LogoutIcon className="w-4 h-4" />
            </Link>
          </div>
        </div>
      </aside>
    </>
  );
};
