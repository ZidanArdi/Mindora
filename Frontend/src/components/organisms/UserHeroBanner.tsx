import type { UpcomingSchedule } from '../../types';
import { UpcomingMeetingCard } from '../molecules';
import { ChevronDownIcon } from '../atoms';
import { cn } from '../../utils';

export interface UserHeroBannerProps {
  userName?: string;
  taskCount?: number;
  upcomingSchedules?: UpcomingSchedule[];
  className?: string;
}

export const UserHeroBanner = ({
  userName = 'John',
  taskCount = 6,
  upcomingSchedules = [],
  className,
}: UserHeroBannerProps) => {
  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-3xl bg-linear-to-r from-emerald-100/70 via-nature-mint/60 to-teal-50/80 border border-emerald-200/60 p-6 sm:p-7 shadow-soft-sm',
        className
      )}
    >
      {/* Botanical Organic Background Silhouette */}
      <div className="absolute right-0 bottom-0 pointer-events-none opacity-20 transform translate-x-8 translate-y-6">
        <svg width="240" height="160" viewBox="0 0 240 160" fill="none">
          <ellipse cx="120" cy="140" rx="100" ry="60" fill="#168F7A" />
          <ellipse cx="190" cy="130" rx="70" ry="50" fill="#20B8A5" />
          <path d="M120 40 L130 90 L110 90 Z" fill="#6F9F78" />
          <path d="M160 50 L170 95 L150 95 Z" fill="#A8C5A5" />
        </svg>
      </div>

      <div className="relative z-10 grid grid-cols-1 md:grid-cols-12 gap-6 items-center">
        {/* Left: Greeting and summary */}
        <div className="md:col-span-7 space-y-2">
          <div className="inline-flex items-center gap-2 cursor-pointer group">
            <h2 className="text-2xl sm:text-3xl font-bold tracking-tight text-primary-900">
              Good Morning, {userName}
            </h2>
            <ChevronDownIcon className="w-5 h-5 text-primary group-hover:translate-y-0.5 transition-transform" />
          </div>
          <p className="text-sm sm:text-base font-medium text-emerald-800/80">
            You have <span className="font-bold text-primary-900 underline decoration-primary/40 decoration-2">{taskCount} Task</span> for today
          </p>
        </div>

        {/* Right: Up Next schedule cards */}
        <div className="md:col-span-5 space-y-2.5">
          <span className="text-[11px] font-bold text-primary-900 uppercase tracking-wider block">
            Up next
          </span>
          <div className="space-y-2">
            {upcomingSchedules.slice(0, 2).map((schedule) => (
              <UpcomingMeetingCard key={schedule.id} schedule={schedule} />
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};
