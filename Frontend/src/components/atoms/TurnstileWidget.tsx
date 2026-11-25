import { useEffect, useRef, useState } from 'react';
import { ShieldIcon } from './Icons';

interface TurnstileWidgetProps {
  siteKey?: string;
  action?: string;
  onVerify: (token: string) => void;
  resetKey?: number | string;
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
          'error-callback'?: (errorCode?: string) => void;
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
  resetKey,
  className = '',
}: TurnstileWidgetProps) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const widgetIdRef = useRef<string | null>(null);
  const [status, setStatus] = useState<'idle' | 'verified' | 'error' | 'expired'>('idle');

  // Keep a ref to onVerify so callback changes don't re-trigger widget lifecycle
  const onVerifyRef = useRef(onVerify);
  useEffect(() => {
    onVerifyRef.current = onVerify;
  }, [onVerify]);

  const hasKey = Boolean(siteKey && siteKey !== 'your_turnstile_site_key_here');

  useEffect(() => {
    if (!hasKey) {
      if (import.meta.env.DEV) {
        console.warn(
          '[Turnstile] VITE_TURNSTILE_SITE_KEY is not configured. For development testing, configure a Cloudflare Turnstile Sitekey in Frontend/.env.'
        );
      }
      return;
    }

    let isMounted = true;
    let pollInterval: ReturnType<typeof setInterval> | null = null;
    let pollTimeout: ReturnType<typeof setTimeout> | null = null;

    const containerEl = containerRef.current;
    const render = () => {
      if (!isMounted || !containerEl || !window.turnstile) return;
      if (widgetIdRef.current) return;

      try {
        containerEl.innerHTML = '';
        const id = window.turnstile.render(containerEl, {
          sitekey: siteKey,
          action,
          callback: (token: string) => {
            if (isMounted) {
              setStatus('verified');
              onVerifyRef.current(token);
            }
          },
          'error-callback': () => {
            if (isMounted) {
              console.warn('[Turnstile] Challenge failed or encountered an error');
              setStatus('error');
              onVerifyRef.current('');
            }
          },
          'expired-callback': () => {
            if (isMounted) {
              console.warn('[Turnstile] Challenge token expired');
              setStatus('expired');
              onVerifyRef.current('');
            }
          },
          theme: 'light',
        });
        widgetIdRef.current = id;
      } catch (err) {
        console.warn('[Turnstile] Render error:', err);
        if (isMounted) setStatus('error');
      }
    };

    if (window.turnstile) {
      render();
    } else {
      let script = document.querySelector('script[src*="challenges.cloudflare.com/turnstile"]') as HTMLScriptElement | null;
      if (!script) {
        script = document.createElement('script');
        script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit';
        script.async = true;
        script.defer = true;
        document.head.appendChild(script);
      }

      pollInterval = setInterval(() => {
        if (window.turnstile) {
          if (pollInterval) clearInterval(pollInterval);
          if (pollTimeout) clearTimeout(pollTimeout);
          render();
        }
      }, 100);

      pollTimeout = setTimeout(() => {
        if (pollInterval) clearInterval(pollInterval);
        if (isMounted && !widgetIdRef.current) {
          console.error('[Turnstile] Script loading timed out');
          setStatus('error');
        }
      }, 8000);
    }

    return () => {
      isMounted = false;
      if (pollInterval) clearInterval(pollInterval);
      if (pollTimeout) clearTimeout(pollTimeout);
      if (widgetIdRef.current && window.turnstile) {
        try {
          window.turnstile.remove(widgetIdRef.current);
        } catch {
          // ignore cleanup errors
        }
        widgetIdRef.current = null;
      }
      if (containerEl) {
        containerEl.innerHTML = '';
      }
    };
  }, [siteKey, action, hasKey]);

  // Handle widget reset when parent form triggers an update
  useEffect(() => {
    if (resetKey !== undefined && widgetIdRef.current && window.turnstile) {
      try {
        window.turnstile.reset(widgetIdRef.current);
      } catch {
        // ignore reset errors
      }
      onVerifyRef.current('');
      setStatus((prev) => (prev !== 'idle' ? 'idle' : prev));
    }
  }, [resetKey]);

  return (
    <div className={`flex flex-col items-center justify-center my-2 ${className}`}>
      <div ref={containerRef} className="min-h-[65px] flex items-center justify-center" />
      {!hasKey && (
        <div className="flex items-center gap-1.5 py-1.5 px-3 rounded-xl bg-amber-50 border border-amber-200 text-[11px] text-amber-700 font-medium">
          <ShieldIcon className="w-3.5 h-3.5 text-amber-600 flex-shrink-0" />
          <span>Turnstile Sitekey belum dikonfigurasi di Frontend/.env</span>
        </div>
      )}
      {hasKey && status === 'error' && (
        <div className="flex items-center gap-1.5 py-1.5 px-3 rounded-xl bg-red-50 border border-red-200 text-[11px] text-red-600 font-medium animate-fadeIn">
          <ShieldIcon className="w-3.5 h-3.5 text-red-500 flex-shrink-0" />
          <span>Verifikasi keamanan gagal. Silakan muat ulang atau coba lagi.</span>
        </div>
      )}
      {hasKey && status === 'expired' && (
        <div className="flex items-center gap-1.5 py-1.5 px-3 rounded-xl bg-amber-50 border border-amber-200 text-[11px] text-amber-700 font-medium animate-fadeIn">
          <ShieldIcon className="w-3.5 h-3.5 text-amber-600 flex-shrink-0" />
          <span>Verifikasi kedaluwarsa. Silakan centang ulang tantangan keamanan.</span>
        </div>
      )}
    </div>
  );
};
