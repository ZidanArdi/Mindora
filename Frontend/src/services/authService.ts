import { API_BASE_URL, apiFetch, setAccessToken, clearAuth } from './api';
import type { AuthResult, LoginPayload, RegisterPayload, UserResponse } from '../types';

export const authService = {
  /**
   * Log in via email and password with Cloudflare Turnstile token.
   */
  login: async (payload: LoginPayload): Promise<AuthResult> => {
    const res = await apiFetch('/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: payload.email,
        password: payload.password,
        turnstile_token: payload.turnstile_token || 'dummy-test-token',
      }),
    });

    const json = await res.json();
    if (!res.ok || !json.success) {
      throw new Error(json.message || 'Login failed');
    }

    setAccessToken(json.data.access_token);
    return json.data;
  },

  /**
   * Register a new user account with enforced 'user' role.
   */
  register: async (payload: RegisterPayload): Promise<AuthResult> => {
    const res = await apiFetch('/auth/register', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: payload.name,
        email: payload.email,
        password: payload.password,
        turnstile_token: payload.turnstile_token || 'dummy-test-token',
      }),
    });

    const json = await res.json();
    if (!res.ok || !json.success) {
      throw new Error(json.message || 'Registration failed');
    }

    setAccessToken(json.data.access_token);
    return json.data;
  },

  /**
   * Refresh session using HttpOnly cookie and operation ID for distributed idempotency.
   */
  refreshToken: async (): Promise<AuthResult> => {
    const operationId = typeof crypto !== 'undefined' && crypto.randomUUID
      ? crypto.randomUUID()
      : 'op_' + Math.random().toString(36).substring(2) + Date.now().toString(36);

    const res = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'X-Operation-ID': operationId,
      },
      credentials: 'include',
      body: JSON.stringify({ operation_id: operationId }),
    });

    const json = await res.json();
    if (!res.ok || !json.success) {
      clearAuth();
      throw new Error(json.message || 'Session refresh failed');
    }

    setAccessToken(json.data.access_token);
    return json.data;
  },

  /**
   * Revoke session on backend and clear local state.
   */
  logout: async (): Promise<void> => {
    try {
      await fetch(`${API_BASE_URL}/auth/logout`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
      });
    } catch (e) {
      console.warn('Logout network notice:', e);
    } finally {
      clearAuth();
    }
  },

  /**
   * Retrieve the current authenticated user's profile and permissions.
   */
  getProfile: async (): Promise<UserResponse> => {
    const res = await apiFetch('/user/profile', {
      method: 'GET',
    });

    const json = await res.json();
    if (!res.ok || !json.success) {
      throw new Error(json.message || 'Failed to fetch user profile');
    }

    return json.data;
  },

  /**
   * Redirect browser to backend Google OAuth 2.0 PKCE initiation endpoint.
   */
  redirectToGoogle: (turnstileToken?: string) => {
    const params = new URLSearchParams();
    if (turnstileToken) {
      params.append('turnstile_token', turnstileToken);
    }
    const queryString = params.toString();
    window.location.href = `${API_BASE_URL}/auth/google${queryString ? `?${queryString}` : ''}`;
  },
};
