import { useAuthStore } from '@/stores/authStore';

export function useAuth() {
  const user = useAuthStore((state) => state.user);
  const isLoading = useAuthStore((state) => state.isLoading);
  const login = useAuthStore((state) => state.login);
  const loginWithGoogle = useAuthStore((state) => state.loginWithGoogle);
  const registerWithGoogle = useAuthStore((state) => state.registerWithGoogle);
  const logout = useAuthStore((state) => state.logout);

  return {
    user,
    isLoading,
    login,
    loginWithGoogle,
    registerWithGoogle,
    logout,
    isAdmin: user?.role === 'admin',
    isClient: user?.role === 'client',
    isEditor: user?.role === 'editor',
    isSpecialist: user?.role === 'ad_specialist',
  };
}
