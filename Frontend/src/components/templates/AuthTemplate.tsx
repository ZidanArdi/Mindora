import type { ReactNode } from 'react';
import { Avatar } from '../atoms';
import { cn } from '../../utils';

export interface AuthTemplateProps {
  children: ReactNode;
  className?: string;
}

export const AuthTemplate = ({ children, className }: AuthTemplateProps) => {
  return (
    <div className={cn('min-h-screen bg-slate-100/60 flex items-center justify-center p-4 sm:p-6 lg:p-10', className)}>
      {/* Main Split Card (Matches Login.pdf) */}
      <div className="w-full max-w-5xl bg-white rounded-3xl shadow-soft-lg border border-border overflow-hidden grid grid-cols-1 lg:grid-cols-12 min-h-[640px]">
        {/* Left: Nature Inspired Branding Banner */}
        <div className="lg:col-span-6 bg-linear-to-b from-teal-50/90 via-emerald-100/60 to-nature-mint/80 p-8 sm:p-10 flex flex-col justify-between relative overflow-hidden border-b lg:border-b-0 lg:border-r border-emerald-200/50">
          {/* Subtle Organic Background shapes */}
          <div className="absolute top-0 right-0 w-80 h-80 bg-primary/10 rounded-full blur-3xl pointer-events-none -mr-20 -mt-20" />
          <div className="absolute bottom-16 left-0 w-64 h-64 bg-secondary/15 rounded-full blur-2xl pointer-events-none -ml-16" />

          {/* Top Branding */}
          <div className="relative z-10 space-y-6">
            <div className="flex items-center gap-2.5">
              <img src="/logo.png" alt="Mindora" className="w-8 h-8 object-contain" />
              <span className="text-xl font-bold tracking-tight text-primary-900">
                Mindora
              </span>
            </div>

            <div className="space-y-3 pt-4 sm:pt-8">
              <span className="text-xs font-bold text-primary tracking-widest uppercase block">
                Plan · Focus · Achieve
              </span>
              <h2 className="text-3xl sm:text-4xl font-extrabold text-primary-900 tracking-tight leading-tight">
                Make every<br />minute feel<br />meaningful.
              </h2>
              <p className="text-sm sm:text-base font-medium text-emerald-900/80 max-w-sm leading-relaxed">
                Organize your day, protect your focus, and move your best work forward with one calm workspace.
              </p>
            </div>
          </div>

          {/* Testimonial Card */}
          <div className="relative z-10 pt-8 sm:pt-12">
            <div className="p-4 rounded-2xl bg-white/80 backdrop-blur-md border border-white/90 shadow-soft-sm space-y-2 max-w-md">
              <div className="flex items-center gap-3">
                <Avatar initials="AK" size="sm" variant="secondary" />
                <div className="min-w-0">
                  <p className="text-xs font-bold text-primary-900">
                    Alicia
                  </p>
                  <p className="text-[10px] text-text-secondary font-medium">
                    Product Designer
                  </p>
                </div>
              </div>
              <p className="text-xs text-text-primary leading-relaxed italic">
                "You don't have to be perfect every day. Even one small step today will bring you closer to your goal."
              </p>
            </div>
          </div>

          {/* Botanical Pine Trees Silhouette at Bottom */}
          <div className="absolute bottom-0 left-0 right-0 pointer-events-none select-none opacity-80">
            <svg viewBox="0 0 400 60" fill="none" className="w-full h-12 object-cover">
              <path d="M0 45 Q 100 30 200 42 T 400 38 L 400 60 L 0 60 Z" fill="#6F9F78" opacity="0.4" />
              <path d="M30 20 L38 35 L22 35 Z" fill="#168F7A" />
              <path d="M70 15 L80 32 L60 32 Z" fill="#183B35" />
              <path d="M120 25 L128 38 L112 38 Z" fill="#168F7A" />
              <path d="M160 12 L172 32 L148 32 Z" fill="#183B35" />
              <path d="M210 22 L220 36 L200 36 Z" fill="#168F7A" />
              <path d="M260 18 L270 34 L250 34 Z" fill="#183B35" />
              <path d="M310 24 L318 38 L302 38 Z" fill="#168F7A" />
              <path d="M355 14 L366 32 L344 32 Z" fill="#183B35" />
            </svg>
          </div>
        </div>

        {/* Right: Interactive Login Form Container */}
        <div className="lg:col-span-6 p-8 sm:p-12 lg:p-14 flex items-center justify-center bg-white">
          {children}
        </div>
      </div>
    </div>
  );
};
