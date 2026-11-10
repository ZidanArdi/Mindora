import { useState, useEffect } from 'react';
import { MinusIcon, PlusIcon, PlayIcon, PauseIcon, ChevronDownIcon } from '../atoms';
import { cn } from '../../utils';

export interface FocusTimerWidgetProps {
  initialMinutes?: number;
  initialTaskName?: string;
  className?: string;
}

export const FocusTimerWidget = ({
  initialMinutes = 30,
  initialTaskName = 'Cleaning up desk',
  className,
}: FocusTimerWidgetProps) => {
  const [minutes, setMinutes] = useState(initialMinutes);
  const [seconds, setSeconds] = useState(0);
  const [isRunning, setIsRunning] = useState(false);
  const [taskName] = useState(initialTaskName);

  useEffect(() => {
    let interval: ReturnType<typeof setInterval> | null = null;
    if (isRunning) {
      interval = setInterval(() => {
        if (seconds > 0) {
          setSeconds((prev) => prev - 1);
        } else if (minutes > 0) {
          setMinutes((prev) => prev - 1);
          setSeconds(59);
        } else {
          setIsRunning(false);
        }
      }, 1000);
    }
    return () => {
      if (interval) clearInterval(interval);
    };
  }, [isRunning, minutes, seconds]);

  const handleMinus = () => {
    if (!isRunning && minutes > 5) setMinutes((m) => m - 5);
  };

  const handlePlus = () => {
    if (!isRunning && minutes < 120) setMinutes((m) => m + 5);
  };

  const toggleRun = () => {
    setIsRunning(!isRunning);
  };

  return (
    <div
      className={cn(
        'relative overflow-hidden rounded-2xl bg-linear-to-b from-teal-50/80 to-emerald-100/60 border border-emerald-200/60 p-6 shadow-soft-sm flex flex-col justify-between space-y-6',
        className
      )}
    >
      <div className="space-y-4">
        {/* Header and task selector */}
        <div className="space-y-1">
          <h4 className="text-lg font-bold text-primary-900 tracking-tight">
            Focus timer
          </h4>
          <div className="inline-flex items-center gap-1.5 text-xs font-semibold text-primary cursor-pointer hover:underline">
            <span>{taskName}</span>
            <ChevronDownIcon className="w-3.5 h-3.5" />
          </div>
        </div>

        {/* Counter and control buttons */}
        <div className="flex items-center justify-center gap-5 py-2">
          <button
            type="button"
            onClick={handleMinus}
            disabled={isRunning || minutes <= 5}
            className="w-8 h-8 rounded-full bg-white border border-border flex items-center justify-center text-text-secondary hover:text-text-primary hover:border-primary disabled:opacity-40 transition-colors shadow-soft-sm cursor-pointer"
            aria-label="Decrease time"
          >
            <MinusIcon className="w-4 h-4" />
          </button>

          <div className="text-center min-w-[110px]">
            <span className="text-3xl font-extrabold text-primary-900 tracking-tight">
              {isRunning
                ? `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
                : `${minutes} mins`}
            </span>
          </div>

          <button
            type="button"
            onClick={handlePlus}
            disabled={isRunning || minutes >= 120}
            className="w-8 h-8 rounded-full bg-white border border-border flex items-center justify-center text-text-secondary hover:text-text-primary hover:border-primary disabled:opacity-40 transition-colors shadow-soft-sm cursor-pointer"
            aria-label="Increase time"
          >
            <PlusIcon className="w-4 h-4" />
          </button>
        </div>

        {/* Action Button */}
        <div className="flex justify-center">
          <button
            type="button"
            onClick={toggleRun}
            className={cn(
              'inline-flex items-center gap-2 px-6 py-2.5 rounded-full text-xs font-bold transition-all shadow-soft-sm cursor-pointer',
              isRunning
                ? 'bg-amber-600 hover:bg-amber-700 text-white'
                : 'bg-white hover:bg-surface-muted text-primary-900 border border-emerald-300'
            )}
          >
            {isRunning ? (
              <>
                <PauseIcon className="w-3.5 h-3.5 text-white" />
                <span>Pause</span>
              </>
            ) : (
              <>
                <PlayIcon className="w-3.5 h-3.5 text-primary" />
                <span>Start</span>
              </>
            )}
          </button>
        </div>
      </div>

      {/* Nature Botanical Pine Tree Silhouette (per Dashboard User.pdf) */}
      <div className="relative pt-6 -mx-6 -mb-6 pointer-events-none select-none">
        <svg viewBox="0 0 320 80" fill="none" className="w-full h-16 object-cover">
          {/* Rolling green hills */}
          <path d="M0 65 Q 80 40 160 55 T 320 50 L 320 80 L 0 80 Z" fill="#6F9F78" opacity="0.6" />
          <path d="M0 60 Q 120 70 240 50 T 320 62 L 320 80 L 0 80 Z" fill="#20B8A5" opacity="0.4" />
          
          {/* Pine trees */}
          {/* Tree 1 */}
          <path d="M40 30 L48 45 L32 45 Z" fill="#168F7A" />
          <path d="M40 40 L50 56 L30 56 Z" fill="#183B35" />
          <rect x="38" y="56" width="4" height="6" fill="#183B35" />

          {/* Tree 2 */}
          <path d="M75 22 L85 40 L65 40 Z" fill="#168F7A" />
          <path d="M75 35 L88 52 L62 52 Z" fill="#183B35" />
          <rect x="73" y="52" width="4" height="6" fill="#183B35" />

          {/* Tree 3 */}
          <path d="M110 32 L118 47 L102 47 Z" fill="#168F7A" />
          <path d="M110 42 L120 57 L100 57 Z" fill="#183B35" />
          <rect x="108" y="57" width="4" height="6" fill="#183B35" />

          {/* Tree 4 */}
          <path d="M150 18 L162 38 L138 38 Z" fill="#168F7A" />
          <path d="M150 32 L165 52 L135 52 Z" fill="#183B35" />
          <rect x="148" y="52" width="4" height="6" fill="#183B35" />

          {/* Tree 5 */}
          <path d="M195 26 L205 42 L185 42 Z" fill="#168F7A" />
          <path d="M195 38 L208 55 L182 55 Z" fill="#183B35" />
          <rect x="193" y="55" width="4" height="6" fill="#183B35" />

          {/* Tree 6 */}
          <path d="M235 22 L245 40 L225 40 Z" fill="#168F7A" />
          <path d="M235 35 L248 53 L222 53 Z" fill="#183B35" />
          <rect x="233" y="53" width="4" height="6" fill="#183B35" />

          {/* Tree 7 */}
          <path d="M275 30 L284 46 L266 46 Z" fill="#168F7A" />
          <path d="M275 42 L286 58 L264 58 Z" fill="#183B35" />
          <rect x="273" y="58" width="4" height="6" fill="#183B35" />
        </svg>
      </div>
    </div>
  );
};
