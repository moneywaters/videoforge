import { create } from 'zustand';
import { persist, createJSONStorage } from 'zustand/middleware';
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
  _hasHydrated: boolean;
  login: (email: string, password: string) => Promise<void>;
  loginWithGoogle: () => void;
  handleGoogleCallback: (token: string) => void;
  logout: () => void;
  setUser: (user: User | null) => void;
  setLoading: (loading: boolean) => void;
  setHasHydrated: (hasHydrated: boolean) => void;
}

// SSR-safe localStorage access
function getLocalStorage(): Storage | null {
  if (typeof window === 'undefined') return null;
  return localStorage;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      isLoading: false,
      _hasHydrated: false,
      login: async (_email: string, _password: string) => {
        throw new Error('Not implemented');
      },
      loginWithGoogle: () => {
        if (typeof window === 'undefined') return;
        const targetUrl = process.env.NEXT_PUBLIC_API_URL
          ? `${process.env.NEXT_PUBLIC_API_URL}/auth/google/login`
          : 'https://videoforge-gateway.fly.dev/api/v1/auth/google/login';
        window.location.href = targetUrl;
      },
      handleGoogleCallback: (token: string) => {
        console.log('[OAuth] handleGoogleCallback called, token length:', token?.length);
        if (typeof window === 'undefined') return;
        const payload = decodeJWT(token);
        console.log('[OAuth] JWT payload:', payload);
        if (!payload.sub) {
          console.error('[OAuth] Invalid token: missing sub claim');
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
        console.log('[OAuth] Creating user object:', user);
        getLocalStorage()?.setItem('token', token);
        set({ user, isLoading: false });
        setTimeout(() => {
          console.log('[OAuth] Store state after set:', useAuthStore.getState().user);
          console.log('[OAuth] localStorage auth-storage:', localStorage.getItem('auth-storage'));
        }, 500);
      },
      logout: () => {
        getLocalStorage()?.removeItem('token');
        set({ user: null });
      },
      setUser: (user: User | null) => set({ user }),
      setLoading: (loading: boolean) => set({ isLoading: loading }),
      setHasHydrated: (hasHydrated: boolean) => set({ _hasHydrated: hasHydrated }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ user: state.user }),
      storage: createJSONStorage(() => {
        if (typeof window === 'undefined') {
          return {
            getItem: () => null,
            setItem: () => {},
            removeItem: () => {},
          };
        }
        return localStorage;
      }),
      onRehydrateStorage: () => (state) => {
        if (state) {
          state._hasHydrated = true;
        }
      },
    }
  )
);

export const hasHydrated = () => useAuthStore.getState()._hasHydrated;

export const isClient = (): boolean =>
  useAuthStore.getState().user?.role === 'client';
export const isEditor = (): boolean =>
  useAuthStore.getState().user?.role === 'editor';
export const isSpecialist = (): boolean =>
  useAuthStore.getState().user?.role === 'ad_specialist';
