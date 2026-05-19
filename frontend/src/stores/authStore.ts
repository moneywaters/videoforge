import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { User } from '@/types/index';
import { api } from '@/lib/api';

interface AuthState {
  user: User | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  setUser: (user: User | null) => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      isLoading: false,
      login: async (email: string, password: string) => {
        set({ isLoading: true });
        try {
          const user = await api.login(email, password);
          set({ user, isLoading: false });
        } catch (error) {
          set({ isLoading: false });
          throw error;
        }
      },
      logout: () => {
        api.logout();
        set({ user: null });
      },
      setUser: (user: User | null) => set({ user }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ user: state.user }),
    }
  )
);

export const isAdmin = (): boolean => useAuthStore.getState().user?.role === 'admin';
export const isClient = (): boolean => useAuthStore.getState().user?.role === 'client';
export const isEditor = (): boolean => useAuthStore.getState().user?.role === 'editor';
export const isSpecialist = (): boolean => useAuthStore.getState().user?.role === 'ad_specialist';