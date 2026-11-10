import type { HTMLAttributes, ReactNode } from 'react';
import { cn } from '../../utils';

export interface RootTemplateProps extends HTMLAttributes<HTMLDivElement> {
  children: ReactNode;
}

export const RootTemplate = ({
  children,
  className,
  ...props
}: RootTemplateProps) => {
  return (
    <div
      className={cn(
        'min-h-screen bg-slate-50 text-slate-900 flex flex-col justify-center items-center p-6',
        className
      )}
      {...props}
    >
      <main className="w-full max-w-lg">{children}</main>
    </div>
  );
};
