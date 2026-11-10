import type { HTMLAttributes } from 'react';
import { cn } from '../../utils';

export interface AvatarProps extends HTMLAttributes<HTMLDivElement> {
  initials?: string;
  src?: string;
  size?: 'sm' | 'md' | 'lg';
  variant?: 'primary' | 'secondary' | 'neutral';
}

export const Avatar = ({
  initials = 'JD',
  src,
  size = 'md',
  variant = 'primary',
  className,
  ...props
}: AvatarProps) => {
  const sizeStyles = {
    sm: 'w-7 h-7 text-xs',
    md: 'w-9 h-9 text-xs',
    lg: 'w-11 h-11 text-sm font-semibold',
  };

  const variantStyles = {
    primary: 'bg-primary-100 text-primary-900 border border-primary-200',
    secondary: 'bg-secondary/15 text-primary-900 border border-secondary/30',
    neutral: 'bg-slate-100 text-slate-700 border border-slate-200',
  };

  return (
    <div
      className={cn(
        'rounded-full flex items-center justify-center font-medium select-none overflow-hidden shrink-0',
        sizeStyles[size],
        variantStyles[variant],
        className
      )}
      {...props}
    >
      {src ? (
        <img src={src} alt="avatar" className="w-full h-full object-cover" />
      ) : (
        <span>{initials}</span>
      )}
    </div>
  );
};
