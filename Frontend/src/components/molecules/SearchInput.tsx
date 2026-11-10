import type { InputHTMLAttributes } from 'react';
import { SearchIcon } from '../atoms';
import { cn } from '../../utils';

export interface SearchInputProps extends InputHTMLAttributes<HTMLInputElement> {
  containerClassName?: string;
}

export const SearchInput = ({
  containerClassName,
  className,
  placeholder = 'Search task...',
  ...props
}: SearchInputProps) => {
  return (
    <div className={cn('relative flex items-center w-full max-w-md', containerClassName)}>
      <SearchIcon className="absolute left-3.5 w-4 h-4 text-text-secondary pointer-events-none" />
      <input
        type="text"
        placeholder={placeholder}
        className={cn(
          'w-full pl-10 pr-4 py-2 text-sm rounded-xl border border-border bg-white text-text-primary placeholder:text-text-secondary/70 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all shadow-soft-sm',
          className
        )}
        {...props}
      />
    </div>
  );
};
