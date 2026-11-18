import React, { useState, useEffect, useCallback } from 'react';
import type { UserResponse, LoginPayload, RegisterPayload, AuthResult } from '../types';
import { authService } from '../services/authService';
import { getAccessToken } from '../services/api';
import { AuthContext, type AuthContextType } from './authContextDef';

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<UserResponse | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [activeWorkspace, setActiveWorkspace] = useState<'user' | 'admin'>('user');

  const refreshUser = useCallback(async () => {
    try {
      const authResult = await authService.refreshToken();
      setUser(authResult.user);
      if (authResult.user.role === 'admin') {
        setActiveWorkspace('admin');
      } else {
        setActiveWorkspace('user');
      }
    } catch {
      setUser(null);
    }
  }, []);

  useEffect(() => {
    const initAuth = async () => {
      try {
        const searchParams = new URLSearchParams(window.location.search);
        const hasAuthSuccess = searchParams.get('auth') === 'success';

        if (hasAuthSuccess) {
          // Complete Google OAuth code exchange callback
          const authResult = await authService.refreshToken();
          setUser(authResult.user);
          if (authResult.user.role === 'admin') {
            setActiveWorkspace('admin');
          } else {
            setActiveWorkspace('user');
          }
          // Clean up URL query parameters without page reload
          searchParams.delete('auth');
          const cleanQuery = searchParams.toString();
          const cleanURL = window.location.pathname + (cleanQuery ? `?${cleanQuery}` : '');
          window.history.replaceState({}, document.title, cleanURL);
        } else {
          // Standard silent refresh check on boot
          await refreshUser();
        }
      } catch (err) {
        console.debug('No active session found during initial boot:', err);
        setUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    initAuth();
  }, [refreshUser]);

  const login = async (payload: LoginPayload): Promise<AuthResult> => {
    setIsLoading(true);
    try {
      const result = await authService.login(payload);
      setUser(result.user);
      if (result.user.role === 'admin') {
        setActiveWorkspace('admin');
      } else {
        setActiveWorkspace('user');
      }
      return result;
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (payload: RegisterPayload): Promise<AuthResult> => {
    setIsLoading(true);
    try {
      const result = await authService.register(payload);
      setUser(result.user);
      setActiveWorkspace('user'); // Public registration is strictly role: user
      return result;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async (): Promise<void> => {
    setIsLoading(true);
    try {
      await authService.logout();
    } finally {
      setUser(null);
      setIsLoading(false);
    }
  };

  const loginWithGoogle = (turnstileToken?: string) => {
    authService.redirectToGoogle(turnstileToken);
  };

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user && !!getAccessToken(),
    isLoading,
    activeWorkspace,
    setActiveWorkspace,
    login,
    register,
    logout,
    loginWithGoogle,
    refreshUser,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
