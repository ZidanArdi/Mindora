import type { ProductivityChartData, ActivityItem } from '../../types';
import { UsersIcon, ListIcon, FolderIcon, ShieldIcon, MoreHorizontalIcon } from '../atoms';
import { cn } from '../../utils';

export interface AdminActivityChartProps {
  chartData: ProductivityChartData[];
  recentActivities: ActivityItem[];
  className?: string;
}

export const AdminActivityChart = ({
  chartData,
  recentActivities,
  className,
}: AdminActivityChartProps) => {
  const activityIconMap = {
    member_joined: <UsersIcon className="w-4 h-4 text-emerald-600" />,
    tasks_completed: <ListIcon className="w-4 h-4 text-blue-600" />,
    workspace_created: <FolderIcon className="w-4 h-4 text-purple-600" />,
    permissions_updated: <ShieldIcon className="w-4 h-4 text-amber-600" />,
  };

  const activityBgMap = {
    member_joined: 'bg-emerald-50 border-emerald-100',
    tasks_completed: 'bg-blue-50 border-blue-100',
    workspace_created: 'bg-purple-50 border-purple-100',
    permissions_updated: 'bg-amber-50 border-amber-100',
  };

  return (
    <div className={cn('grid grid-cols-1 lg:grid-cols-12 gap-6', className)}>
      {/* Left: Task productivity chart */}
      <div className="lg:col-span-8 bg-white rounded-2xl border border-border p-6 shadow-soft-sm flex flex-col justify-between space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
          <div>
            <h4 className="text-base font-bold text-text-primary tracking-tight">
              Task productivity
            </h4>
            <p className="text-xs text-text-secondary font-medium">
              Tasks created and completed this week
            </p>
          </div>

          {/* Legend */}
          <div className="flex items-center gap-4 text-xs font-semibold text-text-secondary">
            <div className="flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-sm bg-primary" />
              <span>Completed</span>
            </div>
            <div className="flex items-center gap-1.5">
              <span className="w-2.5 h-2.5 rounded-sm bg-nature-sage" />
              <span>Created</span>
            </div>
          </div>
        </div>

        {/* Visual Bar Chart */}
        <div className="pt-4 pb-2">
          <div className="h-44 sm:h-52 flex items-end justify-between gap-2 sm:gap-6 border-b border-border-subtle pb-3">
            {chartData.map((item) => (
              <div key={item.day} className="flex-1 flex flex-col items-center gap-2 h-full justify-end group">
                <div className="w-full max-w-[36px] flex items-end justify-center gap-1.5 h-full">
                  {/* Completed bar */}
                  <div
                    className="w-1/2 bg-primary hover:bg-primary-hover rounded-t-sm transition-all duration-300 relative group-hover:opacity-90"
                    style={{ height: `${item.completed}%` }}
                    title={`Completed: ${item.completed}`}
                  />
                  {/* Created bar */}
                  <div
                    className="w-1/2 bg-nature-sage hover:bg-emerald-300 rounded-t-sm transition-all duration-300 relative group-hover:opacity-90"
                    style={{ height: `${item.created}%` }}
                    title={`Created: ${item.created}`}
                  />
                </div>
                <span className="text-xs font-semibold text-text-secondary group-hover:text-primary transition-colors">
                  {item.day}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Right: Recent activity feed */}
      <div className="lg:col-span-4 bg-white rounded-2xl border border-border p-6 shadow-soft-sm flex flex-col justify-between space-y-4">
        <div className="flex items-center justify-between pb-1 border-b border-border-subtle">
          <div>
            <h4 className="text-base font-bold text-text-primary tracking-tight">
              Recent activity
            </h4>
            <p className="text-xs text-text-secondary font-medium">
              Latest team updates
            </p>
          </div>
          <button type="button" className="text-text-secondary hover:text-text-primary p-1" aria-label="More options">
            <MoreHorizontalIcon className="w-4 h-4" />
          </button>
        </div>

        <div className="space-y-3.5 flex-1">
          {recentActivities.map((act) => (
            <div key={act.id} className="flex items-start gap-3">
              <div className={cn('w-8 h-8 rounded-xl border flex items-center justify-center shrink-0 mt-0.5', activityBgMap[act.type])}>
                {activityIconMap[act.type]}
              </div>
              <div className="space-y-0.5 min-w-0 flex-1">
                <p className="text-xs font-semibold text-text-primary truncate">
                  {act.title}
                </p>
                <p className="text-[11px] text-text-secondary truncate">
                  {act.meta}
                </p>
              </div>
            </div>
          ))}
        </div>

        <div className="pt-2 border-t border-border-subtle text-center">
          <button
            type="button"
            className="text-xs font-semibold text-primary hover:text-primary-hover transition-colors cursor-pointer"
          >
            View all activity
          </button>
        </div>
      </div>
    </div>
  );
};
