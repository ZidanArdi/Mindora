import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../utils';

export interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  variant?: 'emerald' | 'slate';
}

export const Badge = ({
  children,
  variant = 'emerald',
  className,
  ...props
}: BadgeProps) => {
  const variantStyles = {
    emerald: 'bg-emerald-50 border-emerald-200 text-emerald-700',
    slate: 'bg-slate-50 border-slate-200 text-slate-700',
  };

  return (
    <div
      className={cn(
        'inline-flex items-center gap-2.5 px-3.5 py-1.5 rounded-full border text-xs font-semibold',
        variantStyles[variant],
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
};
