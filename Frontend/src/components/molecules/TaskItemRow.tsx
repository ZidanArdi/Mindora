import { useState } from 'react';
import type { TaskItem } from '../../types';
import { Checkbox, ChevronRightIcon } from '../atoms';
import { cn } from '../../utils';

export interface TaskItemRowProps {
  task: TaskItem;
  onToggle?: (id: string, completed: boolean) => void;
  className?: string;
}

export const TaskItemRow = ({ task, onToggle, className }: TaskItemRowProps) => {
  const [completed, setCompleted] = useState(task.completed);

  const handleToggle = () => {
    const nextState = !completed;
    setCompleted(nextState);
    onToggle?.(task.id, nextState);
  };

  return (
    <div
      className={cn(
        'group flex items-center justify-between p-3.5 rounded-xl border border-transparent hover:border-border hover:bg-surface-subtle transition-all cursor-pointer select-none',
        completed && 'opacity-60',
        className
      )}
      onClick={handleToggle}
    >
      <div className="flex items-center gap-3.5 min-w-0">
        <Checkbox
          checked={completed}
          onChange={handleToggle}
          onClick={(e) => e.stopPropagation()}
        />
        <div className="space-y-0.5 min-w-0">
          <p
            className={cn(
              'text-sm font-semibold text-text-primary truncate transition-all',
              completed && 'line-through text-text-secondary'
            )}
          >
            {task.title}
          </p>
          <div className="flex items-center gap-2">
            <span
              className="inline-block w-2 h-2 rounded-full shrink-0"
              style={{ backgroundColor: task.categoryColor }}
            />
            <span className="text-xs text-text-secondary font-medium">
              {task.category}
            </span>
          </div>
        </div>
      </div>

      <div className="flex items-center gap-3 shrink-0 ml-4">
        <span className="text-xs font-medium text-text-secondary bg-surface-muted px-2.5 py-1 rounded-lg border border-border-subtle">
          {task.time}
        </span>
        <ChevronRightIcon className="w-4 h-4 text-text-secondary opacity-0 group-hover:opacity-100 transition-opacity" />
      </div>
    </div>
  );
};
