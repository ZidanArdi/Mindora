import { useState } from 'react';
import { Button, BellIcon, ClockIcon, LogoutIcon, ExportIcon, PlusIcon, MenuIcon, Avatar } from '../atoms';
import { SearchInput } from '../molecules';
import type { UserProfile } from '../../types';
import { cn } from '../../utils';
import { Link } from 'react-router-dom';

export interface TopNavbarProps {
  role: 'user' | 'admin';
  currentUser: UserProfile;
  onOpenMobileMenu?: () => void;
  className?: string;
}

export const TopNavbar = ({
  role,
  currentUser,
  onOpenMobileMenu,
  className,
}: TopNavbarProps) => {
  const [hasUnread] = useState(true);

  return (
    <header
      className={cn(
        'sticky top-0 z-30 bg-surface/90 backdrop-blur-md border-b border-border px-4 lg:px-8 py-3.5 flex items-center justify-between gap-4 transition-all',
        className
      )}
    >
      {/* Left: Mobile hamburger & search */}
      <div className="flex items-center gap-3 flex-1 max-w-lg">
        {onOpenMobileMenu && (
          <button
            type="button"
            onClick={onOpenMobileMenu}
            className="p-2 rounded-xl text-text-secondary hover:text-text-primary hover:bg-surface-muted lg:hidden"
            aria-label="Open sidebar navigation"
          >
            <MenuIcon className="w-5 h-5" />
          </button>
        )}

        <SearchInput
          placeholder={
            role === 'user' ? 'Search task...' : 'Search members, tasks, workspace...'
          }
          className="w-full"
        />
      </div>

      {/* Right: Actions, Notifications, User */}
      <div className="flex items-center gap-3">
        {role === 'user' ? (
          <div className="hidden sm:flex items-center gap-2">
            <Button variant="outline" size="sm" className="gap-2 text-xs">
              <ExportIcon className="w-3.5 h-3.5 text-text-secondary" />
              <span>Export summary</span>
            </Button>
          </div>
        ) : (
          <Button variant="primary" size="sm" className="gap-2 text-xs">
            <PlusIcon className="w-3.5 h-3.5" />
            <span>Invite member</span>
          </Button>
        )}

        <div className="flex items-center gap-1.5 border-l border-border pl-3">
          {role === 'user' && (
            <button
              type="button"
              className="p-2 rounded-xl text-text-secondary hover:text-text-primary hover:bg-surface-muted transition-colors relative hidden sm:flex"
              title="Recent time tracker"
              aria-label="Time tracker"
            >
              <ClockIcon className="w-4 h-4" />
            </button>
          )}

          <button
            type="button"
            className="p-2 rounded-xl text-text-secondary hover:text-text-primary hover:bg-surface-muted transition-colors relative"
            title="Notifications"
            aria-label="Notifications"
          >
            <BellIcon className="w-4 h-4" />
            {hasUnread && (
              <span className="absolute top-1.5 right-1.5 w-2 h-2 rounded-full bg-emerald-500 ring-2 ring-white" />
            )}
          </button>

          <Link
            to="/login"
            className="p-2 rounded-xl text-text-secondary hover:text-red-600 hover:bg-surface-muted transition-colors hidden sm:flex"
            title="Sign out"
            aria-label="Sign out"
          >
            <LogoutIcon className="w-4 h-4" />
          </Link>

          <Avatar initials={currentUser.avatarText} size="md" className="ml-1 cursor-pointer" />
        </div>
      </div>
    </header>
  );
};
