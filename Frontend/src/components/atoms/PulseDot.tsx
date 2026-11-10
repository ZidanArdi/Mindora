import type { HTMLAttributes } from 'react';
import { cn } from '../../utils';

export interface PulseDotProps extends HTMLAttributes<HTMLSpanElement> {
  color?: string;
}

export const PulseDot = ({ className, color = 'bg-emerald-500', ...props }: PulseDotProps) => {
  return (
    <span
      className={cn('w-2 h-2 rounded-full animate-pulse', color, className)}
      {...props}
    />
  );
};
