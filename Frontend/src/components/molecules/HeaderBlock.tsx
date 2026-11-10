import type { HTMLAttributes } from 'react';
import { Heading, Text } from '../atoms';
import { cn } from '../../utils';

export interface HeaderBlockProps extends HTMLAttributes<HTMLDivElement> {
  title: string;
  subtitle?: string;
}

export const HeaderBlock = ({
  title,
  subtitle,
  className,
  ...props
}: HeaderBlockProps) => {
  return (
    <div className={cn('space-y-2', className)} {...props}>
      <Heading level={1}>{title}</Heading>
      {subtitle && <Text size="sm">{subtitle}</Text>}
    </div>
  );
};
