import { cn } from '@/lib/utils';
import {
  FileText,
  Film,
  LayoutDashboard,
  MessageSquare,
  TrendingUp,
} from 'lucide-react';
import { NavLink } from 'react-router-dom';

const MOBILE_NAV_ITEMS = [
  { label: 'Dashboard', path: '/', icon: LayoutDashboard },
  { label: 'Briefs', path: '/briefs', icon: FileText },
  { label: 'Videos', path: '/videos', icon: Film },
  { label: 'Performance', path: '/performance', icon: TrendingUp },
  { label: 'Support', path: '/support', icon: MessageSquare },
];

export function MobileNav() {
  return (
    <nav className="fixed bottom-0 left-0 right-0 z-40 border-t border-gray-200 bg-white md:hidden">
      <div className="flex items-center justify-around py-2">
        {MOBILE_NAV_ITEMS.map((item) => {
          const Icon = item.icon;
          return (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                cn(
                  'flex flex-col items-center gap-0.5 px-2 py-1 text-[10px] font-medium transition-colors',
                  isActive
                    ? 'text-gray-900'
                    : 'text-gray-600'
                )
              }
            >
              <Icon className="h-5 w-5" />
              <span>{item.label}</span>
            </NavLink>
          );
        })}
      </div>
    </nav>
  );
}