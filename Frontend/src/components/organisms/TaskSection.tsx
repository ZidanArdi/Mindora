import { useState } from 'react';
import type { TaskItem } from '../../types';
import { PlusIcon, SendIcon, ChevronDownIcon } from '../atoms';
import { TaskItemRow } from '../molecules';
import { cn } from '../../utils';

export interface TaskSectionProps {
  initialTasks: TaskItem[];
  className?: string;
}

export const TaskSection = ({ initialTasks, className }: TaskSectionProps) => {
  const [tasks, setTasks] = useState<TaskItem[]>(initialTasks);
  const [newTaskTitle, setNewTaskTitle] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('Inbox');
  const [selectedWhen, setSelectedWhen] = useState('Now');

  const chips = ['Now', 'Tomorrow', 'Next week', 'Custom'];

  const handleAddTask = () => {
    if (!newTaskTitle.trim()) return;

    const newTask: TaskItem = {
      id: `tsk_${Date.now()}`,
      title: newTaskTitle.trim(),
      category: selectedCategory,
      categoryColor:
        selectedCategory === 'Work'
          ? '#10B981'
          : selectedCategory === 'Study'
          ? '#F59E0B'
          : selectedCategory === 'Health'
          ? '#EC4899'
          : '#3B82F6',
      time: 'Just now',
      completed: false,
    };

    setTasks([newTask, ...tasks]);
    setNewTaskTitle('');
  };

  const handleToggleTask = (id: string, completed: boolean) => {
    setTasks(tasks.map((t) => (t.id === id ? { ...t, completed } : t)));
  };

  return (
    <div className={cn('space-y-6', className)}>
      {/* Quick Add Task Card */}
      <div className="bg-white rounded-2xl border border-border p-4 sm:p-5 shadow-soft-sm space-y-3.5">
        <div className="flex items-center gap-2">
          <input
            type="text"
            value={newTaskTitle}
            onChange={(e) => setNewTaskTitle(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleAddTask()}
            placeholder="Add task for today"
            className="flex-1 text-sm text-text-primary placeholder:text-text-secondary/70 focus:outline-none bg-transparent py-1"
          />
          <button
            type="button"
            onClick={handleAddTask}
            className="p-1.5 rounded-lg text-text-secondary hover:text-primary transition-colors cursor-pointer"
            aria-label="Send or tag task"
          >
            <SendIcon className="w-4 h-4" />
          </button>
          <button
            type="button"
            onClick={handleAddTask}
            className="p-1.5 rounded-lg bg-primary hover:bg-primary-hover text-white transition-colors cursor-pointer"
            aria-label="Add task button"
          >
            <PlusIcon className="w-4 h-4" />
          </button>
        </div>

        {/* Chips row */}
        <div className="flex flex-wrap items-center gap-2 pt-2 border-t border-border-subtle text-xs">
          {/* Category Dropdown simulation */}
          <div
            onClick={() => {
              const cats = ['Inbox', 'Work', 'Study', 'Health'];
              const next = cats[(cats.indexOf(selectedCategory) + 1) % cats.length];
              setSelectedCategory(next);
            }}
            className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full bg-surface-muted border border-border font-medium text-text-primary cursor-pointer hover:bg-slate-100"
            title="Click to change category"
          >
            <span className="w-2 h-2 rounded-full bg-blue-500" />
            <span>{selectedCategory}</span>
            <ChevronDownIcon className="w-3 h-3 text-text-secondary" />
          </div>

          {/* Time chips */}
          {chips.map((chip) => {
            const isSelected = selectedWhen === chip;
            return (
              <button
                key={chip}
                type="button"
                onClick={() => setSelectedWhen(chip)}
                className={cn(
                  'px-3 py-1 rounded-full border transition-all text-xs font-medium cursor-pointer',
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
      </div>

      {/* Today's Tasks List */}
      <div className="bg-white rounded-2xl border border-border p-5 sm:p-6 shadow-soft-sm space-y-3">
        <div className="flex items-center justify-between pb-2">
          <h4 className="text-base font-bold text-text-primary tracking-tight">
            Today's task
          </h4>
          <span className="text-xs text-text-secondary font-medium">
            {tasks.filter((t) => t.completed).length} of {tasks.length} completed
          </span>
        </div>

        <div className="divide-y divide-border-subtle -mx-2">
          {tasks.map((task) => (
            <TaskItemRow
              key={task.id}
              task={task}
              onToggle={handleToggleTask}
            />
          ))}
        </div>
      </div>
    </div>
  );
};
