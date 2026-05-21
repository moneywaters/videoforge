import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '@/types';

function decodeJWT(token: string): Record<string, unknown> {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    return JSON.parse(jsonPayload);
  } catch {
    return {};
  }
}

interface AuthState {
  user: User | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  loginWithGoogle: () => void;
  handleGoogleCallback: (token: string) => void;
  logout: () => void;
  setUser: (user: User | null) => void;
  setLoading: (loading: boolean) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isLoading: false,
      login: async (_email: string, _password: string) => {
        throw new Error('Not implemented');
      },
      loginWithGoogle: () => {
        window.location.href = process.env.NEXT_PUBLIC_API_URL
          ? `${process.env.NEXT_PUBLIC_API_URL}/auth/google/login`
          : 'https://videoforge-gateway.fly.dev/api/v1/auth/google/login';
      },
      handleGoogleCallback: (token: string) => {
        const payload = decodeJWT(token);
        if (!payload.sub) {
          console.error('Invalid token: missing sub claim');
          return;
        }
        const user: User = {
          id: payload.sub as string,
          email: (payload.email as string) || '',
          name: (payload.name as string) || '',
          role: (payload.role as User['role']) || 'client',
          avatar: payload.picture ? (payload.picture as string) : undefined,
          createdAt: new Date().toISOString(),
          onboardingComplete: false,
        };
        localStorage.setItem('token', token);
        set({ user, isLoading: false });
      },
      logout: () => {
        localStorage.removeItem('token');
        set({ user: null });
      },
      setUser: (user: User | null) => set({ user }),
      setLoading: (loading: boolean) => set({ isLoading: loading }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ user: state.user }),
    }
  )
);

export const isAdmin = (): boolean =>
  useAuthStore.getState().user?.role === 'admin';
export const isClient = (): boolean =>
  useAuthStore.getState().user?.role === 'client';
export const isEditor = (): boolean =>
  useAuthStore.getState().user?.role === 'editor';
export const isSpecialist = (): boolean =>
  useAuthStore.getState().user?.role === 'ad_specialist';
