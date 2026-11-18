import { createContext } from 'react';
import type { UserResponse, AuthResult, LoginPayload, RegisterPayload } from '../types';

export interface AuthContextType {
  user: UserResponse | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  activeWorkspace: 'user' | 'admin';
  setActiveWorkspace: (ws: 'user' | 'admin') => void;
  login: (payload: LoginPayload) => Promise<AuthResult>;
  register: (payload: RegisterPayload) => Promise<AuthResult>;
  logout: () => Promise<void>;
  loginWithGoogle: (turnstileToken?: string) => void;
  refreshUser: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);
