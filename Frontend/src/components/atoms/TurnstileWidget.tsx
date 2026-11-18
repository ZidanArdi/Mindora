import { useEffect, useRef } from 'react';
import { ShieldIcon } from './Icons';
import { FEATURES } from '../../config/features';

interface TurnstileWidgetProps {
  siteKey?: string;
  action?: string;
  onVerify: (token: string) => void;
  className?: string;
}

declare global {
  interface Window {
    turnstile?: {
      render: (
        container: HTMLElement | string,
        params: {
          sitekey: string;
          action?: string;
          callback?: (token: string) => void;
          'error-callback'?: () => void;
          'expired-callback'?: () => void;
          theme?: 'light' | 'dark' | 'auto';
        }
      ) => string;
      reset: (widgetId: string) => void;
      remove: (widgetId: string) => void;
    };
  }
}

export const TurnstileWidget = ({
  siteKey = import.meta.env.VITE_TURNSTILE_SITE_KEY || '',
  action = 'login',
  onVerify,
  className = '',
}: TurnstileWidgetProps) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const widgetIdRef = useRef<string | null>(null);
  // Disabled if feature flag is false or no real siteKey configured
  const isBypassed = !FEATURES.ENABLE_TURNSTILE || !siteKey || siteKey === 'your_turnstile_site_key_here';

  useEffect(() => {
    // If Turnstile is temporarily disabled for development, provide standard bypass token
    if (isBypassed) {
      onVerify('1x0000000000000000000000000000000AA');
      return;
    }

    // If Turnstile script is loaded in window
    if (window.turnstile && containerRef.current) {
      try {
        widgetIdRef.current = window.turnstile.render(containerRef.current, {
          sitekey: siteKey,
          action,
          callback: (token: string) => {
            onVerify(token);
          },
          'error-callback': () => {
            console.warn('Turnstile challenge failed or encountered an error');
          },
          'expired-callback': () => {
            onVerify('');
          },
          theme: 'light',
        });
      } catch (err) {
        console.warn('Turnstile render warning:', err);
      }
    } else {
      // In local dev without script, provide valid test token
      onVerify('1x0000000000000000000000000000000AA');
    }

    return () => {
      if (widgetIdRef.current && window.turnstile) {
        try {
          window.turnstile.remove(widgetIdRef.current);
        } catch {
          // ignore cleanup errors
        }
      }
    };
  }, [siteKey, action, onVerify, isBypassed]);

  return (
    <div className={`flex flex-col items-center justify-center my-2 ${className}`}>
      <div ref={containerRef} />
      {isBypassed && (
        <div className="flex items-center gap-1.5 py-1 px-2.5 rounded-lg bg-surface-muted border border-border text-[11px] text-text-muted">
          <ShieldIcon className="w-3.5 h-3.5 text-primary" />
          <span>Cloudflare Turnstile (Dev Mode: Bypassed)</span>
        </div>
      )}
    </div>
  );
};
