import { useEffect, useRef, useState } from 'react';
import { ShieldIcon, CheckIcon } from './Icons';

interface TurnstileWidgetProps {
  siteKey?: string;
  action?: string;
  onVerify: (token: string) => void;
  resetKey?: number | string;
  enabled?: boolean;
  disabledMessage?: string;
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
          appearance?: 'always' | 'execute' | 'interaction-only';
          callback?: (token: string) => void;
          'error-callback'?: (errorCode?: string) => void;
          'expired-callback'?: () => void;
          'timeout-callback'?: () => void;
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
  enabled = true,
  disabledMessage = 'Lengkapi input terlebih dahulu untuk mengaktifkan verifikasi keamanan.',
  className = '',
}: TurnstileWidgetProps) => {
  const containerRef = useRef<HTMLDivElement>(null);
  const widgetIdRef = useRef<string | null>(null);
  const [status, setStatus] = useState<'loading' | 'verified' | 'error' | 'expired'>('loading');

  // Keep a ref to onVerify so callback changes don't re-trigger widget lifecycle
  const onVerifyRef = useRef(onVerify);
  useEffect(() => {
    onVerifyRef.current = onVerify;
  }, [onVerify]);

  const hasKey = Boolean(siteKey && siteKey !== 'your_turnstile_site_key_here');

  useEffect(() => {
    const containerEl = containerRef.current;

    if (!enabled || !hasKey) {
      if (!hasKey && enabled && import.meta.env.DEV) {
        console.warn(
          '[Turnstile] VITE_TURNSTILE_SITE_KEY is not configured. Configure a Cloudflare Turnstile Sitekey in Frontend/.env.'
        );
      }
      onVerifyRef.current('');
      const timer = setTimeout(() => {
        setStatus((prev) => (prev !== 'loading' ? 'loading' : prev));
      }, 0);
      return () => clearTimeout(timer);
    }

    let isMounted = true;
    let pollInterval: ReturnType<typeof setInterval> | null = null;
    let pollTimeout: ReturnType<typeof setTimeout> | null = null;

    const statusTimer = setTimeout(() => {
      if (isMounted) setStatus('loading');
    }, 0);

    const render = () => {
      if (!isMounted || !containerEl || !window.turnstile) return;
      if (widgetIdRef.current) return;

      try {
        containerEl.innerHTML = '';
        const id = window.turnstile.render(containerEl, {
          sitekey: siteKey,
          action,
          appearance: 'always',
          theme: 'light',
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
          'timeout-callback': () => {
            if (isMounted) {
              console.warn('[Turnstile] Challenge timed out');
              setStatus('error');
              onVerifyRef.current('');
            }
          },
          'expired-callback': () => {
            if (isMounted) {
              console.warn('[Turnstile] Challenge token expired');
              setStatus('expired');
              onVerifyRef.current('');
              if (widgetIdRef.current && window.turnstile) {
                try {
                  window.turnstile.reset(widgetIdRef.current);
                } catch {
                  // ignore reset error
                }
              }
            }
          },
        });
        widgetIdRef.current = id;
      } catch (err) {
        console.warn('[Turnstile] Render error:', err);
        if (isMounted) {
          setStatus('error');
          onVerifyRef.current('');
        }
      }
    };

    if (window.turnstile) {
      render();
    } else {
      let script = document.querySelector(
        'script[src*="challenges.cloudflare.com/turnstile"]'
      ) as HTMLScriptElement | null;
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
          onVerifyRef.current('');
        }
      }, 8000);
    }

    return () => {
      isMounted = false;
      clearTimeout(statusTimer);
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
  }, [siteKey, action, hasKey, enabled]);

  // Handle widget reset when parent form triggers an update
  useEffect(() => {
    if (resetKey === undefined || !enabled) return;
    if (widgetIdRef.current && window.turnstile) {
      try {
        window.turnstile.reset(widgetIdRef.current);
      } catch {
        // ignore reset errors
      }
    }
    onVerifyRef.current('');
    const timer = setTimeout(() => {
      setStatus((prev) => (prev !== 'loading' ? 'loading' : prev));
    }, 0);
    return () => clearTimeout(timer);
  }, [resetKey, enabled]);

  const handleRetry = () => {
    onVerifyRef.current('');
    setStatus('loading');
    if (widgetIdRef.current && window.turnstile) {
      try {
        window.turnstile.reset(widgetIdRef.current);
      } catch {
        // ignore reset error
      }
    }
  };

  if (!enabled) {
    return (
      <div className={`flex flex-col items-center justify-center my-2 ${className}`}>
        <div className="w-full flex items-center gap-2.5 py-3 px-3.5 rounded-xl bg-surface-muted border border-border text-xs text-text-secondary font-medium">
          <ShieldIcon className="w-4 h-4 text-text-muted flex-shrink-0" />
          <span>{disabledMessage}</span>
        </div>
      </div>
    );
  }

  return (
    <div className={`flex flex-col items-center justify-center my-2 space-y-2 ${className}`}>
      <div ref={containerRef} className="min-h-[65px] flex items-center justify-center" />

      {!hasKey && (
        <div className="w-full flex items-center gap-2 py-2 px-3 rounded-xl bg-amber-50 border border-amber-200 text-[11px] text-amber-700 font-medium">
          <ShieldIcon className="w-3.5 h-3.5 text-amber-600 flex-shrink-0" />
          <span>Turnstile Sitekey belum dikonfigurasi di Frontend/.env</span>
        </div>
      )}

      {hasKey && status === 'loading' && (
        <div className="flex items-center gap-1.5 py-1 px-3 rounded-xl bg-surface-muted border border-border text-[11px] text-text-secondary font-medium">
          <span className="w-2 h-2 rounded-full bg-primary animate-pulse" />
          <span>Menunggu verifikasi keamanan...</span>
        </div>
      )}

      {hasKey && status === 'verified' && (
        <div className="flex items-center gap-1.5 py-1 px-3 rounded-xl bg-nature-mint border border-primary/20 text-[11px] text-primary-900 font-medium animate-fadeIn">
          <CheckIcon className="w-3.5 h-3.5 text-primary flex-shrink-0" />
          <span>Verifikasi keamanan berhasil.</span>
        </div>
      )}

      {hasKey && status === 'error' && (
        <div className="flex items-center gap-2 py-1.5 px-3 rounded-xl bg-red-50 border border-red-200 text-[11px] text-red-600 font-medium animate-fadeIn">
          <ShieldIcon className="w-3.5 h-3.5 text-red-500 flex-shrink-0" />
          <span>Verifikasi gagal. Silakan coba lagi.</span>
          <button
            type="button"
            onClick={handleRetry}
            className="underline font-semibold text-red-700 hover:text-red-800 cursor-pointer"
          >
            Muat ulang
          </button>
        </div>
      )}

      {hasKey && status === 'expired' && (
        <div className="flex items-center gap-1.5 py-1.5 px-3 rounded-xl bg-amber-50 border border-amber-200 text-[11px] text-amber-700 font-medium animate-fadeIn">
          <ShieldIcon className="w-3.5 h-3.5 text-amber-600 flex-shrink-0" />
          <span>Verifikasi kedaluwarsa. Silakan verifikasi ulang.</span>
        </div>
      )}
    </div>
  );
};
