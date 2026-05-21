import { NavGroup } from '@/types';

export const navGroups: NavGroup[] = [
  {
    label: 'Overview',
    items: [
      {
        title: 'Dashboard',
        url: '/dashboard',
        icon: 'dashboard',
        isActive: false,
        shortcut: ['d', 'd'],
        items: [],
      },
      {
        title: 'Briefs',
        url: '/dashboard/briefs',
        icon: 'page',
        shortcut: ['b', 'b'],
        isActive: false,
        items: [],
      },
      {
        title: 'Videos',
        url: '/dashboard/videos',
        icon: 'media',
        isActive: false,
        items: [],
      },
      {
        title: 'Campaigns',
        url: '/dashboard/campaigns',
        icon: 'trendingUp',
        isActive: false,
        items: [],
      },
      {
        title: 'Leaderboard',
        url: '/dashboard/leaderboard',
        icon: 'award',
        isActive: false,
        items: [],
      },
    ],
  },
  {
    label: 'Business',
    items: [
      {
        title: 'Performance',
        url: '/dashboard/performance',
        icon: 'chart',
        isActive: false,
        items: [],
      },
      {
        title: 'Earnings',
        url: '/dashboard/earnings',
        icon: 'creditCard',
        isActive: false,
        items: [],
      },
      {
        title: 'Payouts',
        url: '/dashboard/payouts',
        icon: 'dollarSign',
        isActive: false,
        items: [],
      },
    ],
  },
  {
    label: 'Admin',
    items: [
      {
        title: 'Users',
        url: '/dashboard/users',
        icon: 'teams',
        isActive: false,
        items: [],
        access: { role: 'admin' },
      },
      {
        title: 'Moderation',
        url: '/dashboard/moderation',
        icon: 'shield',
        isActive: false,
        items: [],
        access: { role: 'admin' },
      },
    ],
  },
  {
    label: '',
    items: [
      {
        title: 'Account',
        url: '#',
        icon: 'account',
        isActive: true,
        items: [
          {
            title: 'Profile',
            url: '/dashboard/profile',
            icon: 'profile',
          },
          {
            title: 'Logout',
            url: '/login',
            icon: 'logout',
          },
        ],
      },
    ],
  },
];
