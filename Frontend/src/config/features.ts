/**
 * Mindora Feature Flags & Development Configuration
 * 
 * Provides explicit control over external third-party integrations
 * so the application can run in offline or local development mode
 * without requiring live Google Cloud Console or Cloudflare Turnstile credentials.
 */

export const FEATURES = {
  /**
   * Google OAuth 2.0 PKCE Integration Flag
   * Defaults to false in development unless explicitly enabled via VITE_ENABLE_GOOGLE_OAUTH=true.
   */
  ENABLE_GOOGLE_OAUTH: import.meta.env.VITE_ENABLE_GOOGLE_OAUTH === 'true',

  /**
   * Cloudflare Turnstile Integration Flag
   * Enabled when VITE_TURNSTILE_SITE_KEY is provided and not set to placeholder,
   * or if explicitly enabled via VITE_ENABLE_TURNSTILE=true.
   * Can be explicitly disabled via VITE_ENABLE_TURNSTILE=false.
   */
  ENABLE_TURNSTILE:
    import.meta.env.VITE_ENABLE_TURNSTILE === 'false'
      ? false
      : Boolean(
          (import.meta.env.VITE_TURNSTILE_SITE_KEY &&
            import.meta.env.VITE_TURNSTILE_SITE_KEY !== 'your_turnstile_site_key_here') ||
            import.meta.env.VITE_ENABLE_TURNSTILE === 'true'
        ),
};
