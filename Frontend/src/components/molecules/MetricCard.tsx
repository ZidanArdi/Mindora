import type { KPICardData } from '../../types';
import { UsersIcon, ListIcon, ChartBarIcon, FolderIcon, TrendingUpIcon } from '../atoms';
import { cn } from '../../utils';

export interface MetricCardProps {
  data: KPICardData;
  className?: string;
}

export const MetricCard = ({ data, className }: MetricCardProps) => {
  const iconMap = {
    users: <UsersIcon className="w-5 h-5 text-emerald-600" />,
    tasks: <ListIcon className="w-5 h-5 text-blue-600" />,
    rate: <ChartBarIcon className="w-5 h-5 text-amber-600" />,
    workspaces: <FolderIcon className="w-5 h-5 text-purple-600" />,
  };

  const bgMap = {
    users: 'bg-emerald-50 border-emerald-100',
    tasks: 'bg-blue-50 border-blue-100',
    rate: 'bg-amber-50 border-amber-100',
    workspaces: 'bg-purple-50 border-purple-100',
  };

  return (
    <div
      className={cn(
        'bg-white rounded-2xl border border-border p-5 shadow-soft-sm hover:shadow-soft-md transition-shadow flex flex-col justify-between space-y-4',
        className
      )}
    >
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold text-text-secondary uppercase tracking-wider">
          {data.title}
        </span>
        <div className={cn('w-9 h-9 rounded-xl border flex items-center justify-center', bgMap[data.iconType])}>
          {iconMap[data.iconType]}
        </div>
      </div>

      <div className="flex items-baseline justify-between">
        <h3 className="text-2xl font-bold tracking-tight text-text-primary">
          {data.value}
        </h3>
        <div className="inline-flex items-center gap-1 text-xs font-semibold text-emerald-700 bg-emerald-50 px-2 py-0.5 rounded-full border border-emerald-200">
          <TrendingUpIcon className="w-3 h-3" />
          <span>{data.change}</span>
        </div>
      </div>
    </div>
  );
};
