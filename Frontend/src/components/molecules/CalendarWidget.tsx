import { useState } from 'react';
import { ChevronLeftIcon, ChevronRightIcon } from '../atoms';
import { cn } from '../../utils';

export interface CalendarWidgetProps {
  className?: string;
}

export const CalendarWidget = ({ className }: CalendarWidgetProps) => {
  const [selectedDay, setSelectedDay] = useState(2); // Matches prototype Dec 2nd

  const daysOfWeek = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

  // December 2026 grid representation (or 31-day prototype alignment)
  // Matching the visual structure from Dashboard User.pdf:
  // Starts with prev month 31, 30, then 1, 2, 3, 4, 5...
  const calendarDays = [
    { day: 31, isCurrentMonth: false, hasEvent: false },
    { day: 30, isCurrentMonth: false, hasEvent: false },
    { day: 1, isCurrentMonth: true, hasEvent: true },
    { day: 2, isCurrentMonth: true, hasEvent: true },
    { day: 3, isCurrentMonth: true, hasEvent: false },
    { day: 4, isCurrentMonth: true, hasEvent: false },
    { day: 5, isCurrentMonth: true, hasEvent: false },
    { day: 6, isCurrentMonth: true, hasEvent: false },
    { day: 7, isCurrentMonth: true, hasEvent: true },
    { day: 8, isCurrentMonth: true, hasEvent: false },
    { day: 9, isCurrentMonth: true, hasEvent: false },
    { day: 10, isCurrentMonth: true, hasEvent: false },
    { day: 11, isCurrentMonth: true, hasEvent: false },
    { day: 12, isCurrentMonth: true, hasEvent: false },
    { day: 13, isCurrentMonth: true, hasEvent: false },
    { day: 14, isCurrentMonth: true, hasEvent: false },
    { day: 15, isCurrentMonth: true, hasEvent: false },
    { day: 16, isCurrentMonth: true, hasEvent: false },
    { day: 17, isCurrentMonth: true, hasEvent: false },
    { day: 18, isCurrentMonth: true, hasEvent: false },
    { day: 19, isCurrentMonth: true, hasEvent: false },
    { day: 20, isCurrentMonth: true, hasEvent: false },
    { day: 21, isCurrentMonth: true, hasEvent: false },
    { day: 22, isCurrentMonth: true, hasEvent: false },
    { day: 23, isCurrentMonth: true, hasEvent: false },
    { day: 24, isCurrentMonth: true, hasEvent: false },
    { day: 25, isCurrentMonth: true, hasEvent: false },
    { day: 26, isCurrentMonth: true, hasEvent: false },
    { day: 27, isCurrentMonth: true, hasEvent: false },
    { day: 28, isCurrentMonth: true, hasEvent: false },
    { day: 29, isCurrentMonth: true, hasEvent: false },
    { day: 30, isCurrentMonth: true, hasEvent: false },
    { day: 31, isCurrentMonth: true, hasEvent: false },
    { day: 1, isCurrentMonth: false, hasEvent: false },
    { day: 2, isCurrentMonth: false, hasEvent: false },
  ];

  return (
    <div className={cn('bg-white rounded-2xl border border-border p-5 shadow-soft-sm', className)}>
      <div className="flex items-center justify-between mb-4">
        <button
          type="button"
          className="p-1 rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-muted transition-colors"
          aria-label="Previous month"
        >
          <ChevronLeftIcon className="w-4 h-4" />
        </button>

        <div className="text-center">
          <h4 className="text-sm font-bold text-text-primary">December</h4>
          <span className="text-[10px] text-text-secondary font-medium tracking-wide">2026</span>
        </div>

        <button
          type="button"
          className="p-1 rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-muted transition-colors"
          aria-label="Next month"
        >
          <ChevronRightIcon className="w-4 h-4" />
        </button>
      </div>

      <div className="grid grid-cols-7 gap-1 text-center mb-2">
        {daysOfWeek.map((day) => (
          <span key={day} className="text-[11px] font-semibold text-text-secondary/70 py-1">
            {day}
          </span>
        ))}
      </div>

      <div className="grid grid-cols-7 gap-1 text-center">
        {calendarDays.map((item, index) => {
          const isSelected = item.isCurrentMonth && item.day === selectedDay;
          return (
            <button
              key={`${item.day}_${index}`}
              type="button"
              onClick={() => item.isCurrentMonth && setSelectedDay(item.day)}
              className={cn(
                'relative h-8 rounded-lg text-xs font-medium flex flex-col items-center justify-center transition-colors',
                !item.isCurrentMonth && 'text-slate-300 pointer-events-none',
                item.isCurrentMonth && !isSelected && 'text-text-primary hover:bg-surface-muted',
                isSelected && 'bg-primary text-white font-bold shadow-soft-sm'
              )}
            >
              <span>{item.day}</span>
              {item.hasEvent && !isSelected && (
                <span className="w-1 h-1 rounded-full bg-primary mt-0.5" />
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
};
