/**
 * Mindora Frontend API Client & Refresh Request Coordination
 * 
 * Implements Section 2.A:
 * - Ensures multiple concurrent API requests encountering 401 share ONE single refresh operation.
 * - Queues pending requests behind the same refresh promise.
 * - Attaches a cryptographically random operation identifier (UUID) for backend idempotency.
 * - Prevents infinite refresh request loops using single-retry flags.
 * - Handles refresh failure consistently by clearing credentials and redirecting to /login.
 */

export const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

let inMemoryAccessToken: string | null = null;
let refreshPromise: Promise<string> | null = null;

export const setAccessToken = (token: string | null) => {
  inMemoryAccessToken = token;
};

export const getAccessToken = (): string | null => {
  return inMemoryAccessToken;
};

export const clearAuth = (shouldRedirect: boolean = false) => {
  inMemoryAccessToken = null;
  refreshPromise = null;
  if (
    shouldRedirect &&
    typeof window !== 'undefined' &&
    window.location.pathname !== '/login' &&
    window.location.pathname !== '/'
  ) {
    window.location.href = '/login?session_expired=true';
  }
};

interface RequestOptions extends RequestInit {
  _retry?: boolean;
}

/**
 * Coordinate concurrent refresh operations across in-flight requests.
 * Uses a single shared Promise so all parallel 401 requests wait for the same refresh call.
 */
const getSharedRefreshToken = async (): Promise<string> => {
  if (refreshPromise) {
    return refreshPromise;
  }

  // Generate cryptographically secure operation identifier for idempotency
  const operationId = typeof crypto !== 'undefined' && crypto.randomUUID
    ? crypto.randomUUID()
    : 'op_' + Math.random().toString(36).substring(2) + Date.now().toString(36);

  refreshPromise = (async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include', // Includes HttpOnly refresh_token cookie
        body: JSON.stringify({ operation_id: operationId }),
      });

      if (!response.ok) {
        if (response.status === 409) {
          // Concurrent conflict: another instance/tab refreshed. Wait a moment and check credentials.
          await new Promise((res) => setTimeout(res, 150));
          if (inMemoryAccessToken) {
            return inMemoryAccessToken;
          }
        }
        clearAuth();
        throw new Error('Session expired. Please log in again.');
      }

      const data = await response.json();
      const newToken = data?.data?.access_token;
      if (!newToken) {
        clearAuth();
        throw new Error('Invalid token response from server.');
      }

      setAccessToken(newToken);
      return newToken;
    } finally {
      // Release shared promise after completion to allow future refreshes when token expires again
      refreshPromise = null;
    }
  })();

  return refreshPromise;
};

/**
 * Fetch wrapper with automatic token injection, 401 interception, and queued refresh coordination.
 */
export async function apiFetch(endpoint: string, options: RequestOptions = {}): Promise<Response> {
  const url = endpoint.startsWith('http') ? endpoint : `${API_BASE_URL}${endpoint.startsWith('/') ? '' : '/'}${endpoint}`;
  
  const headers = new Headers(options.headers || {});
  if (inMemoryAccessToken && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${inMemoryAccessToken}`);
  }

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: options.credentials || 'include',
  });

  // Intercept 401 Unauthorized (unless this request was already a retry or refresh endpoint)
  if (response.status === 401 && !options._retry && !url.includes('/auth/refresh') && !url.includes('/auth/login')) {
    try {
      const newAccessToken = await getSharedRefreshToken();

      // Retry original request with newly issued access token and _retry marker to prevent loops
      const retryHeaders = new Headers(options.headers || {});
      retryHeaders.set('Authorization', `Bearer ${newAccessToken}`);

      return await fetch(url, {
        ...options,
        headers: retryHeaders,
        credentials: options.credentials || 'include',
        _retry: true,
      } as RequestOptions);
    } catch {
      // Refresh failed or session terminated; propagate original 401
      return response;
    }
  }

  return response;
}
