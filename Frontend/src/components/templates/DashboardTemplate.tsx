import { useState, type ReactNode } from 'react';
import { Sidebar, TopNavbar } from '../organisms';
import type { UserProfile, NavCategory, WorkspaceItem } from '../../types';
import { cn } from '../../utils';

export interface DashboardTemplateProps {
  role: 'user' | 'admin';
  currentUser: UserProfile;
  myLists?: NavCategory[];
  workspaces?: WorkspaceItem[];
  children: ReactNode;
  className?: string;
}

export const DashboardTemplate = ({
  role,
  currentUser,
  myLists,
  workspaces,
  children,
  className,
}: DashboardTemplateProps) => {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <div className={cn('min-h-screen bg-surface-muted flex', className)}>
      {/* Sidebar */}
      <Sidebar
        role={role}
        currentUser={currentUser}
        myLists={myLists}
        workspaces={workspaces}
        isOpen={mobileMenuOpen}
        onClose={() => setMobileMenuOpen(false)}
      />

      {/* Main Content Area */}
      <div className="flex-1 flex flex-col min-w-0 min-h-screen">
        <TopNavbar
          role={role}
          currentUser={currentUser}
          onOpenMobileMenu={() => setMobileMenuOpen(true)}
        />

        <main className="flex-1 p-4 sm:p-6 lg:p-8 max-w-7xl w-full mx-auto space-y-6">
          {children}
        </main>
      </div>
    </div>
  );
};
