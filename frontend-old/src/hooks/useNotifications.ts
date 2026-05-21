import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { useAuthStore } from '@/stores/authStore';

export function useNotifications() {
  const user = useAuthStore((state) => state.user);
  
  const query = useQuery({
    queryKey: ['notifications', user?.id],
    queryFn: () => {
      if (!user) return Promise.resolve([]);
      return api.getNotifications(user.id);
    },
    enabled: !!user,
  });

  const unreadCount = query.data?.filter(n => !n.read).length ?? 0;

  return {
    ...query,
    notifications: query.data ?? [],
    unreadCount,
  };
}

export function useMarkNotificationRead() {
  const queryClient = useQueryClient();
  const user = useAuthStore((state) => state.user);

  return useMutation({
    mutationFn: (id: string) => api.markNotificationRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['notifications', user?.id] });
    },
  });
}