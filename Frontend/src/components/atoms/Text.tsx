import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../utils';

export interface TextProps extends HTMLAttributes<HTMLParagraphElement> {
  children: ReactNode;
  size?: 'xs' | 'sm' | 'base';
  variant?: 'muted' | 'default';
}

export const Text = ({
  children,
  size = 'sm',
  variant = 'muted',
  className,
  ...props
}: TextProps) => {
  const sizeStyles = {
    xs: 'text-xs',
    sm: 'text-sm',
    base: 'text-base',
  };

  const variantStyles = {
    muted: 'text-slate-500 font-medium',
    default: 'text-slate-800',
  };

  return (
    <p
      className={cn(sizeStyles[size], variantStyles[variant], className)}
      {...props}
    >
      {children}
    </p>
  );
};
