import ReactMarkdown from 'react-markdown';
import { cn } from '../../utils';

export interface MarkdownContentProps {
  content: string;
  className?: string;
}

export const MarkdownContent = ({ content, className }: MarkdownContentProps) => {
  return (
    <div className={cn('markdown-body leading-relaxed text-inherit', className)}>
      <ReactMarkdown
        components={{
          p: ({ children }) => <p className="mb-2 last:mb-0 leading-relaxed">{children}</p>,
          strong: ({ children }) => <strong className="font-semibold text-inherit">{children}</strong>,
          em: ({ children }) => <em className="italic">{children}</em>,
          ul: ({ children }) => <ul className="list-disc pl-5 my-2 space-y-1">{children}</ul>,
          ol: ({ children }) => <ol className="list-decimal pl-5 my-2 space-y-1">{children}</ol>,
          li: ({ children }) => <li className="leading-relaxed">{children}</li>,
          h1: ({ children }) => <h1 className="text-base font-bold my-2">{children}</h1>,
          h2: ({ children }) => <h2 className="text-sm font-bold my-2">{children}</h2>,
          h3: ({ children }) => <h3 className="text-xs font-bold my-1.5">{children}</h3>,
          code: ({ children }) => (
            <code className="px-1.5 py-0.5 rounded bg-black/10 font-mono text-[11px]">{children}</code>
          ),
          blockquote: ({ children }) => (
            <blockquote className="border-l-2 border-emerald-500/40 pl-3 italic my-2 text-inherit opacity-90">
              {children}
            </blockquote>
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
};
