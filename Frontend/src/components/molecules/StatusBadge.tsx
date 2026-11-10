import type { HTMLAttributes } from 'react';
import { Badge, PulseDot } from '../atoms';
import { cn } from '../../utils';

export interface StatusBadgeProps extends HTMLAttributes<HTMLDivElement> {
  label: string;
  dotColor?: string;
}

export const StatusBadge = ({
  label,
  dotColor = 'bg-emerald-500',
  className,
  ...props
}: StatusBadgeProps) => {
  return (
    <Badge className={cn(className)} {...props}>
      <PulseDot color={dotColor} />
      {label}
    </Badge>
  );
};
