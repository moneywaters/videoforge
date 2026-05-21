import { useAuth } from '@/hooks/useAuth';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/Button';
import {
  CreditCard,
  DollarSign,
  FileText,
  Film,
  LayoutDashboard,
  MessageSquare,
  ShieldAlert,
  ShoppingBag,
  Target,
  TrendingUp,
  Trophy,
  Users,
  X,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { NavLink } from 'react-router-dom';

interface NavItem {
  label: string;
  path: string;
  icon: LucideIcon;
  roles?: string[];
}

const NAV_ITEMS: NavItem[] = [
  { label: 'Dashboard', path: '/', icon: LayoutDashboard, roles: ['client', 'editor', 'ad_specialist', 'admin', 'support_ai'] },
  { label: 'Briefs', path: '/briefs', icon: FileText, roles: ['client', 'editor', 'ad_specialist', 'admin', 'support_ai'] },
  { label: 'Videos', path: '/videos', icon: Film, roles: ['client', 'editor', 'ad_specialist', 'admin', 'support_ai'] },
  { label: 'Campaigns', path: '/campaigns', icon: Target, roles: ['client', 'ad_specialist', 'admin'] },
  { label: 'Performance', path: '/performance', icon: TrendingUp, roles: ['client', 'editor', 'ad_specialist', 'admin', 'support_ai'] },
  { label: 'Leaderboards', path: '/leaderboards', icon: Trophy, roles: ['client', 'editor', 'ad_specialist', 'admin', 'support_ai'] },
  { label: 'Shopify Links', path: '/shopify/links', icon: ShoppingBag, roles: ['client', 'admin'] },
  { label: 'Earnings', path: '/earnings', icon: DollarSign, roles: ['editor', 'ad_specialist'] },
  { label: 'Payouts', path: '/payouts', icon: CreditCard, roles: ['admin'] },
  { label: 'Users', path: '/admin/users', icon: Users, roles: ['admin'] },
  { label: 'Moderation', path: '/admin/moderation', icon: ShieldAlert, roles: ['admin'] },
  { label: 'Support', path: '/support', icon: MessageSquare, roles: ['client', 'editor', 'ad_specialist', 'admin', 'support_ai'] },
];

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const { user, isAdmin, isClient, isEditor, isSpecialist } = useAuth();

  const roleMap: Record<string, boolean> = {
    admin: isAdmin,
    client: isClient,
    editor: isEditor,
    ad_specialist: isSpecialist,
  };

  const visibleItems = NAV_ITEMS.filter((item) => {
    if (!item.roles) return true;
    if (!user) return false;
    return item.roles.some((role) => roleMap[role]);
  });

  return (
    <>
      {/* Mobile overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 lg:hidden"
          onClick={onClose}
        />
      )}

      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-40 w-64 transform bg-white transition-transform lg:fixed',
          isOpen ? 'translate-x-0' : '-translate-x-full',
          'lg:translate-x-0'
        )}
      >
        <div className="flex h-full flex-col">
          <div className="flex h-16 items-center justify-between border-b border-gray-200 px-4">
            <span className="text-xl font-bold text-gray-900">
              VideoForge
            </span>
            <Button
              variant="ghost"
              size="sm"
              className="lg:hidden"
              onClick={onClose}
            >
              <X className="h-5 w-5" />
            </Button>
          </div>

          <nav className="flex-1 space-y-1 overflow-y-auto px-2 py-4">
            {visibleItems.map((item) => {
              const Icon = item.icon;
              return (
                <NavLink
                  key={item.path}
                  to={item.path}
                  onClick={() => {
                    if (window.innerWidth < 1024) onClose();
                  }}
                  className={({ isActive }) =>
                    cn(
                      'flex items-center rounded-md px-3 py-2 text-sm transition-colors',
                      isActive
                        ? 'bg-gray-100 text-gray-900 font-medium'
                        : 'text-gray-600 hover:bg-gray-50'
                    )
                  }
                >
                  <Icon className="mr-3 h-5 w-5 shrink-0" />
                  {item.label}
                </NavLink>
              );
            })}
          </nav>
        </div>
      </aside>
    </>
  );
}