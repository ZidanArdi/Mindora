import type { KPICardData } from '../../types';
import { MetricCard } from '../../components/molecules';
import { ChevronDownIcon, CalendarIcon } from '../../components/atoms';
import { cn, getGreeting } from '../../utils';

export interface AdminOverviewRowProps {
  userName?: string;
  kpis: KPICardData[];
  className?: string;
}

export const AdminOverviewRow = ({
  userName = 'Admin',
  kpis,
  className,
}: AdminOverviewRowProps) => {
  return (
    <div className={cn('space-y-6', className)}>
      {/* Header section with date selector */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
        <div className="space-y-1">
          <span className="text-[11px] font-bold text-primary tracking-wider uppercase block">
            Ringkasan Administrator
          </span>
          <h2 className="text-2xl sm:text-3xl font-bold tracking-tight text-text-primary">
            {getGreeting(userName)}
          </h2>
          <p className="text-sm text-text-secondary font-medium">
            Berikut aktivitas terkini di seluruh tim dan instansi Anda.
          </p>
        </div>

        {/* Date Filter Dropdown */}
        <div className="inline-flex items-center gap-2 px-3.5 py-2 rounded-xl border border-border bg-white text-xs font-semibold text-text-primary shadow-soft-sm cursor-pointer hover:bg-surface-muted self-start sm:self-auto transition-colors">
          <CalendarIcon className="w-4 h-4 text-primary" />
          <span>Minggu Ini</span>
          <ChevronDownIcon className="w-3.5 h-3.5 text-text-secondary" />
        </div>
      </div>

      {/* KPI Cards Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-5">
        {kpis.map((kpi) => (
          <MetricCard key={kpi.id} data={kpi} />
        ))}
      </div>
    </div>
  );
};

