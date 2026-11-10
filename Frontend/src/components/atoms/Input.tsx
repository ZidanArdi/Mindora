import type { InputHTMLAttributes, ReactNode } from 'react';
import { cn } from '../../utils';

export interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  leftIcon?: ReactNode;
  rightIcon?: ReactNode;
  error?: string;
}

export const Input = ({
  leftIcon,
  rightIcon,
  error,
  className,
  ...props
}: InputProps) => {
  return (
    <div className="w-full">
      <div className="relative flex items-center">
        {leftIcon && (
          <span className="absolute left-3.5 text-text-secondary pointer-events-none flex items-center">
            {leftIcon}
          </span>
        )}
        <input
          className={cn(
            'w-full rounded-xl border border-border bg-white text-sm text-text-primary placeholder:text-text-secondary/70 transition-colors focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary disabled:bg-slate-50 disabled:text-slate-400',
            leftIcon ? 'pl-10' : 'pl-3.5',
            rightIcon ? 'pr-10' : 'pr-3.5',
            'py-2.5',
            error && 'border-red-400 focus:ring-red-200 focus:border-red-500',
            className
          )}
          {...props}
        />
        {rightIcon && (
          <span className="absolute right-3.5 text-text-secondary flex items-center">
            {rightIcon}
          </span>
        )}
      </div>
      {error && <p className="mt-1.5 text-xs text-red-600 font-medium">{error}</p>}
    </div>
  );
};
