import { DashboardTemplate } from '../components/templates';
import { UserHeroBanner, AIAssistantWidget, TaskSection, FocusTimerWidget } from '../components/organisms';
import { CalendarWidget } from '../components/molecules';
import {
  mockCurrentUser,
  mockUserTasks,
  mockUpcomingSchedules,
  mockMyLists,
  mockWorkspaces,
} from '../mocks';

export default function UserDashboardPage() {
  return (
    <DashboardTemplate
      role="user"
      currentUser={mockCurrentUser}
      myLists={mockMyLists}
      workspaces={mockWorkspaces}
    >
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        {/* Main Feed Column (8 cols on desktop) */}
        <div className="lg:col-span-8 space-y-6">
          {/* Greeting Hero Banner with Up Next Schedule */}
          <UserHeroBanner
            userName={mockCurrentUser.name}
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
    </DashboardTemplate>
  );
}
