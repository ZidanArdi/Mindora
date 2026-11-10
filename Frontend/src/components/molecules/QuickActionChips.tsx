import { useState } from 'react';
import { cn } from '../../utils';

export interface QuickActionChipsProps {
  chips: string[];
  defaultSelected?: string;
  onSelect?: (chip: string) => void;
  className?: string;
  size?: 'sm' | 'xs';
}

export const QuickActionChips = ({
  chips,
  defaultSelected,
  onSelect,
  className,
  size = 'xs',
}: QuickActionChipsProps) => {
  const [selected, setSelected] = useState(defaultSelected || chips[0]);

  const handleSelect = (chip: string) => {
    setSelected(chip);
    onSelect?.(chip);
  };

  return (
    <div className={cn('flex items-center gap-2 overflow-x-auto no-scrollbar py-1', className)}>
      {chips.map((chip) => {
        const isSelected = selected === chip;
        return (
          <button
            key={chip}
            type="button"
            onClick={() => handleSelect(chip)}
            className={cn(
              'rounded-full border transition-all whitespace-nowrap font-medium select-none cursor-pointer',
              size === 'xs' ? 'px-3 py-1 text-xs' : 'px-3.5 py-1.5 text-xs',
              isSelected
                ? 'bg-primary-100 border-primary-300 text-primary-900 font-semibold'
                : 'bg-white border-border text-text-secondary hover:text-text-primary hover:bg-surface-muted'
            )}
          >
            {chip}
          </button>
        );
      })}
    </div>
  );
};
