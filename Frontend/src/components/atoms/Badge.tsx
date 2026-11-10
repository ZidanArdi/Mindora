import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../utils';

export interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
  variant?: 'emerald' | 'amber' | 'blue' | 'slate' | 'mint' | 'danger';
}

export const Badge = ({
  children,
  variant = 'emerald',
  className,
  ...props
}: BadgeProps) => {
  const variantStyles = {
    emerald: 'bg-emerald-50 border-emerald-200 text-emerald-700',
    amber: 'bg-amber-50 border-amber-200 text-amber-700',
    blue: 'bg-blue-50 border-blue-200 text-blue-700',
    slate: 'bg-slate-50 border-slate-200 text-slate-700',
    mint: 'bg-nature-mint border-primary-200 text-primary-900',
    danger: 'bg-red-50 border-red-200 text-red-700',
  };

  return (
    <div
      className={cn(
        'inline-flex items-center gap-1.5 px-3 py-1 rounded-full border text-xs font-semibold',
        variantStyles[variant],
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
};
