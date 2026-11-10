import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../utils';

export interface HeadingProps extends HTMLAttributes<HTMLHeadingElement> {
  children: ReactNode;
  level?: 1 | 2 | 3 | 4;
}

export const Heading = ({
  children,
  level = 1,
  className,
  ...props
}: HeadingProps) => {
  const baseStyles = 'font-bold tracking-tight text-slate-900';

  if (level === 2) {
    return (
      <h2 className={cn('text-2xl', baseStyles, className)} {...props}>
        {children}
      </h2>
    );
  }
  if (level === 3) {
    return (
      <h3 className={cn('text-xl', baseStyles, className)} {...props}>
        {children}
      </h3>
    );
  }
  if (level === 4) {
    return (
      <h4 className={cn('text-lg', baseStyles, className)} {...props}>
        {children}
      </h4>
    );
  }

  return (
    <h1 className={cn('text-3xl', baseStyles, className)} {...props}>
      {children}
    </h1>
  );
};
