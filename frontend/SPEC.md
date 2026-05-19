# Spec: VideoForge Frontend (Mockup)

## Objective

Build an interactive mockup SPA for VideoForge that demonstrates all 11 backend microservice functionalities. The mockup uses realistic mocked APIs and data so stakeholders can navigate the full user journey without a running backend.

## Tech Stack

| Layer | Choice |
|-------|--------|
| Framework | React 18 |
| Language | TypeScript 5.x |
| Build Tool | Vite 5 |
| Styling | Tailwind CSS 3.4 |
| Routing | React Router 6 |
| State | React Query (TanStack Query) for server state, Zustand for client state |
| Icons | Lucide React |
| Charts | Recharts |
| UI Primitives | Headless UI + Tailwind |

## Commands

```bash
# Install dependencies
npm install

# Dev server
npm run dev              # http://localhost:5173

# Build
npm run build            # dist/

# Type check
npm run typecheck        # tsc --noEmit

# Lint
npm run lint             # eslint
```

## Project Structure

```
frontend/
├── public/              # Static assets
├── src/
│   ├── main.tsx         # Entry point
│   ├── App.tsx          # Router + providers
│   ├── index.css        # Tailwind directives + base styles
│   ├── lib/
│   │   ├── api.ts       # Mock API layer (mirrors backend endpoints)
│   │   ├── utils.ts     # cn() and helpers
│   │   └── constants.ts # App constants
│   ├── hooks/
│   │   ├── useAuth.ts   # Auth state + login/logout
│   │   ├── useMockApi.ts # Generic mock query/mutation hook
│   │   └── useNotifications.ts
│   ├── stores/
│   │   └── authStore.ts # Zustand auth store
│   ├── types/
│   │   └── index.ts     # Shared TypeScript types (mirrors backend models)
│   ├── components/
│   │   ├── layout/
│   │   │   ├── Shell.tsx      # Main app shell (sidebar + header)
│   │   │   ├── Sidebar.tsx    # Navigation sidebar (role-aware)
│   │   │   ├── TopBar.tsx     # Header with search + notifications
│   │   │   └── MobileNav.tsx  # Mobile bottom/sheet nav
│   │   ├── ui/            # Primitives (Button, Input, Card, Badge, Modal, Skeleton)
│   │   └── shared/        # Shared composite components
│   └── pages/
│       ├── auth/
│       │   ├── Login.tsx
│       │   ├── Register.tsx
│       │   └── Onboarding.tsx
│       ├── dashboard/
│       │   └── Dashboard.tsx  # Role-based landing dashboard
│       ├── briefs/
│       │   ├── BriefList.tsx
│       │   ├── BriefDetail.tsx
│       │   └── BriefCreate.tsx # Includes AI interview mock
│       ├── videos/
│       │   ├── VideoList.tsx
│       │   ├── VideoDetail.tsx # Approval workflow, revisions
│       │   └── Leaderboard.tsx # Blind leaderboard
│       ├── campaigns/
│       │   ├── CampaignList.tsx
│       │   └── CampaignDetail.tsx
│       ├── performance/
│       │   ├── Analytics.tsx
│       │   └── Leaderboards.tsx
│       ├── payouts/
│       │   ├── Earnings.tsx
│       │   └── PayoutHistory.tsx
│       ├── shopify/
│       │   ├── Stores.tsx
│       │   └── Links.tsx
│       ├── admin/
│       │   ├── Users.tsx
│       │   ├── Moderation.tsx
│       │   └── Disputes.tsx
│       ├── support/
│       │   └── Chat.tsx         # AI Support chat
│       └── NotFound.tsx
├── index.html
├── tailwind.config.js
├── postcss.config.js
├── tsconfig.json
├── tsconfig.app.json
├── vite.config.ts
└── package.json
```

## Code Style

```tsx
// Component: PascalCase, props interface, default export
export function BriefCard({ brief }: BriefCardProps) {
  return (
    <div className="rounded-lg border bg-white p-4 shadow-sm">
      <h3 className="text-base font-semibold text-gray-900">{brief.title}</h3>
      <p className="mt-1 text-sm text-gray-500">{brief.description}</p>
    </div>
  );
}

// Hooks: camelCase, use prefix
function useBriefs() {
  return useQuery({ queryKey: ['briefs'], queryFn: mockApi.getBriefs });
}

// File paths match route segments for colocation
```

## Testing Strategy

- Mockup phase: **No automated tests** (manual navigation verification)
- Build must pass `npm run build` and `npm run typecheck`

## Boundaries

- **Always:** Run `npm run typecheck` after file changes. Use Tailwind spacing scale.
- **Ask first:** Adding new dependencies. Changing routing structure.
- **Never:** Commit secrets. Use raw hex colors (use Tailwind palette). Leave empty pages without at least a skeleton.

## Success Criteria

- [ ] `npm run dev` starts without errors
- [ ] `npm run build` completes successfully
- [ ] All user roles can log in and see role-appropriate sidebar items
- [ ] Client can complete AI interview mock → create brief → see brief list
- [ ] Editor can see briefs, submit videos (mock), see blind leaderboard
- [ ] Ad specialist can create campaign for approved videos
- [ ] Performance shows charts and leaderboards with mock data
- [ ] Admin can view user list, moderation queue, disputes
- [ ] AI Support chat interface works with mock responses
- [ ] Responsive: usable at 320px, 768px, 1024px, 1440px
- [ ] No "AI aesthetic" — clean SaaS grays, purposeful layout, real content

## Open Questions

1. Should we include dark mode toggle in mockup? → **Not required for MVP mockup**
2. Should we mock WebSocket real-time updates? → **Yes, simulate with interval polling for notifications**
