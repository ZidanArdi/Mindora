import type { InputHTMLAttributes } from 'react';
import { CheckIcon } from './Icons';
import { cn } from '../../utils';

export interface CheckboxProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'type'> {
  label?: string;
}

export const Checkbox = ({
  label,
  checked,
  onChange,
  className,
  id,
  ...props
}: CheckboxProps) => {
  const inputId = id || (label ? `cb_${label.replace(/\s+/g, '_').toLowerCase()}` : undefined);

  return (
    <label htmlFor={inputId} className={cn('inline-flex items-center gap-2.5 cursor-pointer select-none text-sm text-text-secondary hover:text-text-primary transition-colors', className)}>
      <div className="relative flex items-center justify-center">
        <input
          id={inputId}
          type="checkbox"
          checked={checked}
          onChange={onChange}
          className="peer sr-only"
          {...props}
        />
        <div className="w-4 h-4 rounded border border-border bg-white transition-all peer-checked:bg-primary peer-checked:border-primary peer-focus-visible:ring-2 peer-focus-visible:ring-primary/30 flex items-center justify-center">
          <CheckIcon className={cn('w-3 h-3 text-white transition-opacity', checked ? 'opacity-100' : 'opacity-0')} />
        </div>
      </div>
      {label && <span>{label}</span>}
    </label>
  );
};
