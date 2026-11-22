import { useState, useMemo } from 'react';
import { ChevronLeftIcon, ChevronRightIcon } from '../atoms';
import { cn } from '../../utils';

export interface CalendarWidgetProps {
  className?: string;
  initialDate?: Date;
  onSelectDate?: (date: Date) => void;
}

export const CalendarWidget = ({ className, initialDate, onSelectDate }: CalendarWidgetProps) => {
  const [viewDate, setViewDate] = useState(() => initialDate || new Date());
  const [selectedDate, setSelectedDate] = useState(() => initialDate || new Date());

  const daysOfWeek = ['Sen', 'Sel', 'Rab', 'Kam', 'Jum', 'Sab', 'Min'];

  const monthNames = [
    'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
    'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember',
  ];

  const year = viewDate.getFullYear();
  const month = viewDate.getMonth();

  const handlePrevMonth = () => {
    setViewDate(new Date(year, month - 1, 1));
  };

  const handleNextMonth = () => {
    setViewDate(new Date(year, month + 1, 1));
  };

  const handleSelectDay = (day: number) => {
    const nextDate = new Date(year, month, day);
    setSelectedDate(nextDate);
    onSelectDate?.(nextDate);
  };

  const calendarDays = useMemo(() => {
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);

    // Convert JS Sunday (0) to Monday-based (0 = Mon, 6 = Sun)
    const startDayOfWeek = (firstDay.getDay() + 6) % 7;
    const daysInMonth = lastDay.getDate();
    const daysInPrevMonth = new Date(year, month, 0).getDate();

    const days = [];

    // Previous month padding
    for (let i = startDayOfWeek - 1; i >= 0; i--) {
      days.push({
        day: daysInPrevMonth - i,
        isCurrentMonth: false,
        hasEvent: false,
      });
    }

    // Current month days
    for (let d = 1; d <= daysInMonth; d++) {
      days.push({
        day: d,
        isCurrentMonth: true,
        hasEvent: false,
      });
    }

    // Next month padding to fill remaining row cells (complete multiple of 7)
    const remainder = days.length % 7;
    if (remainder > 0) {
      const nextMonthPadding = 7 - remainder;
      for (let d = 1; d <= nextMonthPadding; d++) {
        days.push({
          day: d,
          isCurrentMonth: false,
          hasEvent: false,
        });
      }
    }

    return days;
  }, [year, month]);

  const [today] = useState(() => new Date());
  const isTodayCurrentMonth = today.getFullYear() === year && today.getMonth() === month;

  return (
    <div className={cn('bg-white rounded-2xl border border-border p-5 shadow-soft-sm', className)}>
      <div className="flex items-center justify-between mb-4">
        <button
          type="button"
          onClick={handlePrevMonth}
          className="p-1 rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-muted transition-colors cursor-pointer"
          aria-label="Bulan sebelumnya"
        >
          <ChevronLeftIcon className="w-4 h-4" />
        </button>

        <div className="text-center">
          <h4 className="text-sm font-bold text-text-primary">{monthNames[month]}</h4>
          <span className="text-[10px] text-text-secondary font-medium tracking-wide">{year}</span>
        </div>

        <button
          type="button"
          onClick={handleNextMonth}
          className="p-1 rounded-lg text-text-secondary hover:text-text-primary hover:bg-surface-muted transition-colors cursor-pointer"
          aria-label="Bulan berikutnya"
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
          const isSelected =
            item.isCurrentMonth &&
            selectedDate.getFullYear() === year &&
            selectedDate.getMonth() === month &&
            selectedDate.getDate() === item.day;

          const isToday =
            isTodayCurrentMonth && item.isCurrentMonth && today.getDate() === item.day;

          return (
            <button
              key={`${item.day}_${index}`}
              type="button"
              onClick={() => item.isCurrentMonth && handleSelectDay(item.day)}
              className={cn(
                'relative h-8 rounded-lg text-xs font-medium flex flex-col items-center justify-center transition-colors',
                !item.isCurrentMonth && 'text-slate-300 pointer-events-none',
                item.isCurrentMonth && !isSelected && 'text-text-primary hover:bg-surface-muted cursor-pointer',
                isSelected && 'bg-primary text-white font-bold shadow-soft-sm cursor-pointer',
                isToday && !isSelected && 'ring-1 ring-primary/40 font-bold'
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
