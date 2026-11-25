import type { InputHTMLAttributes } from 'react';
import { SearchIcon, CloseIcon } from '../atoms';
import { cn } from '../../utils';

export interface SearchInputProps extends InputHTMLAttributes<HTMLInputElement> {
  containerClassName?: string;
  onClear?: () => void;
  isLoading?: boolean;
}

export const SearchInput = ({
  containerClassName,
  className,
  placeholder = 'Search task...',
  value,
  onClear,
  isLoading,
  ...props
}: SearchInputProps) => {
  return (
    <div className={cn('relative flex items-center w-full max-w-md', containerClassName)}>
      <SearchIcon className="absolute left-3.5 w-4 h-4 text-text-secondary pointer-events-none" />
      <input
        type="text"
        placeholder={placeholder}
        value={value}
        className={cn(
          'w-full pl-10 pr-9 py-2 text-sm rounded-xl border border-border bg-white text-text-primary placeholder:text-text-secondary/70 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all shadow-soft-sm',
          className
        )}
        {...props}
      />
      {isLoading && (
        <div className="absolute right-3 w-4 h-4 border-2 border-primary border-t-transparent rounded-full animate-spin" />
      )}
      {!isLoading && onClear && value && String(value).length > 0 && (
        <button
          type="button"
          onClick={onClear}
          className="absolute right-3 p-0.5 rounded text-text-muted hover:text-text-primary cursor-pointer transition-colors"
          aria-label="Bersihkan pencarian"
        >
          <CloseIcon className="w-3.5 h-3.5" />
        </button>
      )}
    </div>
  );
};
