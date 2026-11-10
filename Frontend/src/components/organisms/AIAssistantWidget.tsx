import { useState } from 'react';
import { SparklesIcon, SendIcon } from '../atoms';
import { QuickActionChips } from '../molecules';
import { cn } from '../../utils';

export interface AIAssistantWidgetProps {
  className?: string;
  onAskAI?: (prompt: string) => void;
}

export const AIAssistantWidget = ({
  className,
  onAskAI,
}: AIAssistantWidgetProps) => {
  const [prompt, setPrompt] = useState('');
  const [responseMsg, setResponseMsg] = useState<string | null>(null);

  const quickPrompts = ['Plan my day', 'Prioritize tasks', 'Create focus plan'];

  const handleSend = (textToSend?: string) => {
    const text = textToSend || prompt;
    if (!text.trim()) return;

    if (text === 'Plan my day') {
      setResponseMsg('Based on your schedule, start with "Client meeting prep" at 10:20 AM, followed by 30 mins focused deep work.');
    } else if (text === 'Prioritize tasks') {
      setResponseMsg('Highest priority today: 1. Client meeting prep, 2. Machine Learning research paper, 3. Project 3 draft check.');
    } else if (text === 'Create focus plan') {
      setResponseMsg('Focus recommendation: 2 Pomodoro intervals (25 mins focus + 5 mins break) before lunchtime.');
    } else {
      setResponseMsg(`Planning with Mindora AI for "${text}"... Tasks optimized for your productivity patterns.`);
    }

    onAskAI?.(text);
    setPrompt('');
  };

  return (
    <div
      className={cn(
        'bg-surface rounded-2xl border border-border p-6 shadow-soft-sm space-y-4 hover:border-primary/30 transition-all',
        className
      )}
    >
      {/* Header with Sparkle Badge */}
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-primary-100 text-primary-900 text-xs font-semibold">
          <SparklesIcon className="w-3.5 h-3.5 text-primary" />
          <span>AI PLANNING ASSISTANT</span>
        </div>
        <span className="text-[11px] font-medium text-text-secondary">
          Powered by your preferred LLM
        </span>
      </div>

      {/* Headline */}
      <h3 className="text-lg sm:text-xl font-bold tracking-tight text-text-primary">
        What can I help you finish today?
      </h3>

      {/* Quick Action Chips */}
      <QuickActionChips
        chips={quickPrompts}
        onSelect={(chip) => handleSend(chip)}
        size="sm"
      />

      {/* Input Prompt Box */}
      <div className="relative flex items-center pt-1">
        <input
          type="text"
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSend()}
          placeholder="Ask Mindora AI about your tasks..."
          className="w-full pl-4 pr-12 py-3 rounded-xl border border-border bg-surface-subtle text-sm text-text-primary placeholder:text-text-secondary/70 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary focus:bg-white transition-all shadow-soft-sm"
        />
        <button
          type="button"
          onClick={() => handleSend()}
          className="absolute right-2 p-2 rounded-lg bg-primary hover:bg-primary-hover text-white transition-colors cursor-pointer"
          aria-label="Send to AI Assistant"
        >
          <SendIcon className="w-4 h-4" />
        </button>
      </div>

      {/* Mock AI Response message */}
      {responseMsg && (
        <div className="p-3.5 rounded-xl bg-nature-mint/60 border border-primary-200 text-xs text-primary-900 leading-relaxed font-medium animate-fadeIn">
          <span className="font-bold mr-1">Mindora AI:</span>
          {responseMsg}
        </div>
      )}
    </div>
  );
};
