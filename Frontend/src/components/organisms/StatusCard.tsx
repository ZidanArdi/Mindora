import type { HTMLAttributes } from 'react';
import { HeaderBlock, StatusBadge } from '../molecules';
import { cn } from '../../utils';

export interface StatusCardProps extends HTMLAttributes<HTMLDivElement> {
  title?: string;
  subtitle?: string;
  statusText?: string;
}

export const StatusCard = ({
  title = 'Mindora',
  subtitle = 'Student Productivity Application',
  statusText = 'Frontend is running',
  className,
  ...props
}: StatusCardProps) => {
  return (
    <div
      className={cn(
        'bg-white rounded-2xl shadow-sm border border-slate-200 p-8 text-center space-y-6',
        className
      )}
      {...props}
    >
      <HeaderBlock title={title} subtitle={subtitle} />
      <StatusBadge label={statusText} />
    </div>
  );
};
