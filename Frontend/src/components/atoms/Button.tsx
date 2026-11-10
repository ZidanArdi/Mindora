import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { cn } from '../../utils';

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  children: ReactNode;
  variant?: 'primary' | 'secondary' | 'outline' | 'ghost';
  size?: 'sm' | 'md' | 'lg';
}

export const Button = ({
  children,
  variant = 'primary',
  size = 'md',
  className,
  ...props
}: ButtonProps) => {
  const variantStyles = {
    primary: 'bg-primary hover:bg-primary-hover text-white shadow-soft-sm',
    secondary: 'bg-nature-mint hover:bg-primary-200 text-primary-900',
    outline: 'border border-border bg-white hover:bg-surface-muted text-text-primary',
    ghost: 'text-text-secondary hover:text-text-primary hover:bg-surface-muted',
  };

  const sizeStyles = {
    sm: 'px-3 py-1.5 text-xs rounded-lg',
    md: 'px-4 py-2.5 text-sm rounded-xl',
    lg: 'px-5 py-3 text-base rounded-xl font-semibold',
  };

  return (
    <button
      className={cn(
        'inline-flex items-center justify-center font-medium transition-colors select-none focus:outline-none focus:ring-2 focus:ring-primary/20 disabled:opacity-50 disabled:pointer-events-none cursor-pointer',
        sizeStyles[size],
        variantStyles[variant],
        className
      )}
      {...props}
    >
      {children}
    </button>
  );
};
