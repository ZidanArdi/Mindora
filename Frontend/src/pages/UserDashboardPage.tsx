import { useNavigate } from 'react-router-dom';
import { DashboardTemplate } from '../components/templates';
import { UserHeroBanner, AIAssistantWidget, TaskSection, FocusTimerWidget } from '../components/organisms';
import { CalendarWidget } from '../components/molecules';
import { ShieldIcon, Button } from '../components/atoms';
import {
  mockCurrentUser,
  mockUserTasks,
  mockUpcomingSchedules,
  mockMyLists,
  mockWorkspaces,
} from '../mocks';
import { useAuth } from '../hooks';
import type { UserProfile } from '../types';

export default function UserDashboardPage() {
  const navigate = useNavigate();
  const { user } = useAuth();

  const displayUser: UserProfile = user
    ? {
        id: user.id,
        name: user.name,
        email: user.email,
        role: user.role,
        avatarText: user.name
          .split(' ')
          .map((n) => n[0])
          .join('')
          .slice(0, 2)
          .toUpperCase() || 'ST',
        workspaceName: 'Academic Workspace',
      }
    : mockCurrentUser;

  const isAdminViewingUserWorkspace = user?.role === 'admin';

  return (
    <DashboardTemplate
      role="user"
      currentUser={displayUser}
      myLists={mockMyLists}
      workspaces={mockWorkspaces}
    >
      {isAdminViewingUserWorkspace ? (
        <div className="p-8 rounded-2xl bg-surface-card border border-border space-y-4 text-center max-w-2xl mx-auto shadow-soft-sm my-12 animate-fadeIn">
          <div className="w-12 h-12 rounded-xl bg-nature-mint text-primary flex items-center justify-center mx-auto">
            <ShieldIcon className="w-6 h-6" />
          </div>
          <h2 className="text-xl font-bold text-text-primary">
            Student Privacy Policy Enforcement
          </h2>
          <p className="text-sm text-text-secondary leading-relaxed">
            In accordance with Mindora RBAC and student data privacy rules, personal productivity records, private study tasks, and focus timer sessions are strictly isolated and cannot be accessed from an administrative account.
          </p>
          <div className="pt-2 flex justify-center">
            <Button
              variant="primary"
              size="md"
              onClick={() => navigate('/admin')}
              className="text-xs font-bold"
            >
              Go to Admin Console
            </Button>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          {/* Main Feed Column (8 cols on desktop) */}
          <div className="lg:col-span-8 space-y-6">
            {/* Greeting Hero Banner with Up Next Schedule */}
            <UserHeroBanner
              userName={displayUser.name}
              taskCount={mockUserTasks.length}
              upcomingSchedules={mockUpcomingSchedules}
            />

            {/* AI Planning Assistant Card */}
            <AIAssistantWidget />

            {/* Task Management Section (Add task + Today's list) */}
            <TaskSection initialTasks={mockUserTasks} />
          </div>

          {/* Right Rail Column (4 cols on desktop) */}
          <div className="lg:col-span-4 space-y-6">
            {/* Calendar Month Widget */}
            <CalendarWidget />

            {/* Focus Timer Widget with Pine Tree Silhouette */}
            <FocusTimerWidget />
          </div>
        </div>
      )}
    </DashboardTemplate>
  );
}
