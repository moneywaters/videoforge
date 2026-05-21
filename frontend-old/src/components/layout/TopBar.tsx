import { useState, useRef, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { useNotifications } from '@/hooks/useNotifications';
import { useMarkNotificationRead } from '@/hooks/useNotifications';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Avatar } from '@/components/ui/Avatar';
import { Badge } from '@/components/ui/Badge';
import { Bell, Search, Menu, LogOut } from 'lucide-react';

interface TopBarProps {
  onMenuClick: () => void;
}

export function TopBar({ onMenuClick }: TopBarProps) {
  const { user, logout } = useAuth();
  const { notifications, unreadCount } = useNotifications();
  const markAsReadMutation = useMarkNotificationRead();
  const [notificationsOpen, setNotificationsOpen] = useState(false);
  const [userDropdownOpen, setUserDropdownOpen] = useState(false);
  const notificationsRef = useRef<HTMLDivElement>(null);
  const userDropdownRef = useRef<HTMLDivElement>(null);

  const recentNotifications = notifications.slice(0, 5);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (notificationsRef.current && !notificationsRef.current.contains(event.target as Node)) {
        setNotificationsOpen(false);
      }
      if (userDropdownRef.current && !userDropdownRef.current.contains(event.target as Node)) {
        setUserDropdownOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleMarkAsRead = (id: string) => {
    markAsReadMutation.mutate(id);
  };

  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-gray-200 bg-white px-4">
      <div className="flex items-center gap-4">
        <Button
          variant="ghost"
          className="lg:hidden"
          onClick={onMenuClick}
        >
          <Menu className="h-5 w-5" />
        </Button>
        <span className="text-xl font-bold text-gray-900">
          VideoForge
        </span>
      </div>

      <div className="flex items-center gap-4">
        <div className="relative hidden md:block">
          <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
          <Input
            placeholder="Search..."
            className="w-64 pl-8"
            type="search"
          />
        </div>

        <div className="relative" ref={notificationsRef}>
          <button
            className="relative rounded-full p-2 text-gray-600 hover:bg-gray-100"
            aria-label="Notifications"
            onClick={() => setNotificationsOpen((v) => !v)}
          >
            <Bell className="h-5 w-5" />
            {unreadCount > 0 && (
              <span className="absolute right-1 top-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-600 text-[10px] font-medium text-white">
                {unreadCount > 9 ? '9+' : unreadCount}
              </span>
            )}
          </button>

          {notificationsOpen && (
            <div className="absolute right-0 mt-2 w-80 rounded-md border border-gray-200 bg-white py-1 shadow-lg">
              <div className="border-b border-gray-100 px-4 py-2">
                <p className="text-sm font-medium text-gray-900">Notifications</p>
              </div>
              {recentNotifications.length > 0 ? (
                <div className="max-h-80 overflow-y-auto">
                  {recentNotifications.map((notification) => (
                    <button
                      key={notification.id}
                      className={`flex w-full items-start px-4 py-3 text-left text-sm hover:bg-gray-50 ${
                        !notification.read ? 'bg-blue-50' : ''
                      }`}
                      onClick={() => handleMarkAsRead(notification.id)}
                    >
                      <div className="flex-1">
                        <p className="text-gray-900">{notification.message}</p>
                        <p className="text-xs text-gray-500 mt-1">
                          {new Date(notification.createdAt).toLocaleDateString()}
                        </p>
                      </div>
                      {!notification.read && (
                        <span className="h-2 w-2 rounded-full bg-blue-600 mt-2" />
                      )}
                    </button>
                  ))}
                </div>
              ) : (
                <div className="px-4 py-8 text-center text-sm text-gray-500">
                  No notifications
                </div>
              )}
            </div>
          )}
        </div>

        <div className="relative" ref={userDropdownRef}>
          <button
            className="flex items-center gap-2 rounded-full p-1 hover:bg-gray-100"
            onClick={() => setUserDropdownOpen((v) => !v)}
          >
            <Avatar
              src={user?.avatar}
              name={user?.name || 'Guest'}
            />
            <span className="hidden md:block text-sm text-gray-700">
              {user?.name || 'Guest'}
            </span>
          </button>

          {userDropdownOpen && (
            <div className="absolute right-0 mt-2 w-56 rounded-md border border-gray-200 bg-white py-1 shadow-lg">
              <div className="border-b border-gray-100 px-4 py-3">
                <p className="text-sm font-medium text-gray-900">
                  {user?.name || 'Guest'}
                </p>
                <div className="mt-1">
                  <Badge variant="default" className="capitalize">
                    {user?.role?.replace('_', ' ') || 'User'}
                  </Badge>
                </div>
              </div>
              <button
                className="flex w-full items-center gap-2 px-4 py-2 text-sm text-red-600 hover:bg-gray-50"
                onClick={() => {
                  logout();
                  setUserDropdownOpen(false);
                }}
              >
                <LogOut className="h-4 w-4" />
                Log out
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  );
}