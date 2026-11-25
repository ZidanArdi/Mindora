import { Toaster as SonnerToaster } from 'sonner';

type ToasterProps = React.ComponentProps<typeof SonnerToaster>;

export const Toaster = ({ ...props }: ToasterProps) => {
  return (
    <SonnerToaster
      position="top-right"
      richColors
      closeButton
      duration={3500}
      className="toaster group"
      toastOptions={{
        classNames: {
          toast:
            'group toast font-sans rounded-2xl border border-border bg-white text-text-primary shadow-soft-lg text-xs p-4 flex items-center gap-3',
          description: 'text-text-secondary text-[11px] leading-relaxed',
          actionButton:
            'bg-primary text-white rounded-xl text-xs font-semibold px-3 py-1.5 hover:bg-primary-hover transition-colors',
          cancelButton:
            'bg-surface-muted text-text-secondary rounded-xl text-xs font-semibold px-3 py-1.5 hover:bg-slate-200 transition-colors',
          closeButton:
            'border-border text-text-muted hover:text-text-primary rounded-lg transition-colors',
          success:
            '!bg-emerald-50/90 !border-emerald-200 !text-emerald-950 font-medium',
          error:
            '!bg-rose-50/90 !border-rose-200 !text-rose-950 font-medium',
          warning:
            '!bg-amber-50/90 !border-amber-200 !text-amber-950 font-medium',
          info:
            '!bg-teal-50/90 !border-teal-200 !text-teal-950 font-medium',
        },
      }}
      {...props}
    />
  );
};
