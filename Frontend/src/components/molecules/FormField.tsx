import type { ReactNode } from 'react';
import { Input, type InputProps } from '../atoms';
import { cn } from '../../utils';

export interface FormFieldProps extends InputProps {
  label: string;
  actionElement?: ReactNode;
  containerClassName?: string;
}

export const FormField = ({
  label,
  actionElement,
  containerClassName,
  id,
  ...props
}: FormFieldProps) => {
  const inputId = id || `field_${label.replace(/\s+/g, '_').toLowerCase()}`;

  return (
    <div className={cn('space-y-1.5 w-full', containerClassName)}>
      <div className="flex items-center justify-between">
        <label htmlFor={inputId} className="text-xs font-semibold text-text-primary tracking-wide">
          {label}
        </label>
        {actionElement}
      </div>
      <Input id={inputId} {...props} />
    </div>
  );
};
