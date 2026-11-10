import type { UpcomingSchedule } from '../../types';
import { ClockIcon } from '../atoms';
import { cn } from '../../utils';

export interface UpcomingMeetingCardProps {
  schedule: UpcomingSchedule;
  className?: string;
}

export const UpcomingMeetingCard = ({ schedule, className }: UpcomingMeetingCardProps) => {
  return (
    <div
      className={cn(
        'p-3.5 rounded-xl border border-border bg-white shadow-soft-sm hover:border-primary/40 transition-colors space-y-1.5',
        className
      )}
    >
      <h5 className="text-xs font-semibold text-text-primary leading-tight line-clamp-2">
        {schedule.title}
      </h5>
      <div className="flex items-center gap-1.5 text-[11px] text-text-secondary font-medium">
        <ClockIcon className="w-3.5 h-3.5 text-primary" />
        <span>{schedule.timeRange}</span>
      </div>
    </div>
  );
};
