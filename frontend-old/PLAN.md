# Implementation Plan: VideoForge Frontend Mockup

## Overview

Build a Vite + React + TypeScript + Tailwind SPA mockup that exposes all 11 backend service functionalities through realistic UI and mocked APIs.

## Architecture Decisions

1. **Vite over Next.js**: SPA is sufficient for mockup; no SSR complexity needed.
2. **Mock API layer**: `src/lib/api.ts` returns Promises with `setTimeout` to simulate network latency.
3. **Zustand for auth**: Lightweight, no provider boilerplate.
4. **React Query for server state**: caching, loading/error states out of the box.
5. **Role-based routing**: Sidebar and dashboard widgets adapt to `client | editor | ad_specialist | admin`.

## Task List

### Phase 1: Foundation (Sequential — must complete first)
- [ ] **Task 1: Project Scaffold**
  - Run `npm create vite@latest . -- --template react-ts`
  - Install deps: tailwindcss postcss autoprefixer react-router-dom @tanstack/react-query zustand lucide-react recharts clsx tailwind-merge @headlessui/react
  - Configure Tailwind, PostCSS, TS path aliases (`@/*`)
  - Create folder structure: `src/{lib,hooks,stores,types,components, pages}`
  - Implement shared types in `src/types/index.ts` (mirrors backend: User, Brief, Video, Campaign, Performance, Payout, etc.)
  - Implement mock API layer `src/lib/api.ts` with all CRUD functions returning realistic mock data
  - Implement UI primitives: `Button`, `Input`, `Card`, `Badge`, `Modal`, `Skeleton`, `Select`, `Textarea`
  - Implement `Shell`, `Sidebar` (role-aware navigation), `TopBar`, `MobileNav`
  - Set up React Router with route stubs for ALL pages (empty placeholder components)
  - Set up Zustand auth store + `useAuth` hook
  - Set up QueryClient provider
  - **Acceptance:** `npm run dev` starts, `npm run typecheck` passes, all routes accessible
  - **Verify:** `npm run build` succeeds
  - **Files touched:** `package.json`, `vite.config.ts`, `tailwind.config.js`, `tsconfig.json`, `src/main.tsx`, `src/App.tsx`, `src/index.css`, `src/types/index.ts`, `src/lib/*`, `src/components/layout/*`, `src/components/ui/*`, `src/stores/*`, `src/hooks/*`, `src/pages/**/index.ts` (stubs)

### Phase 2: Feature Pages (Parallelizable after Task 1)
- [ ] **Task 2: Auth + Dashboard + Notifications**
  - `pages/auth/Login.tsx`, `Register.tsx`, `Onboarding.tsx` (role-specific steps)
  - `pages/dashboard/Dashboard.tsx` with role-based widget grid (client: active briefs, recent videos; editor: available briefs, submissions; admin: stats cards)
  - `TopBar` notification dropdown with mock real-time events
  - **Acceptance:** Can log in as different roles, see different dashboards, see notification count/badge
  - **Verify:** Navigate through auth flow without console errors

- [ ] **Task 3: Briefs + Videos**
  - `pages/briefs/BriefList.tsx` (filters, status badges, bounty display)
  - `pages/briefs/BriefDetail.tsx` (submissions count, status timeline)
  - `pages/briefs/BriefCreate.tsx` with multi-step AI interview mock (chat-style questionnaire)
  - `pages/videos/VideoList.tsx` (editor view + client review view)
  - `pages/videos/VideoDetail.tsx` (approval actions: approve/reject/request-revision, revision history)
  - `pages/videos/Leaderboard.tsx` (blind — only rank and sales numbers)
  - **Acceptance:** Client can create brief via AI interview, editor can view and "submit", client can approve/reject
  - **Verify:** End-to-end brief→video flow navigable

- [ ] **Task 4: Campaigns + Performance + Shopify**
  - `pages/campaigns/CampaignList.tsx`, `CampaignDetail.tsx` (budget, platform, status)
  - `pages/performance/Analytics.tsx` (Recharts: line chart for sales over time, bar chart for comparison)
  - `pages/performance/Leaderboards.tsx` (editor + specialist rankings with mock data)
  - `pages/shopify/Stores.tsx` (connected stores list)
  - `pages/shopify/Links.tsx` (generated links table, UTM params)
  - **Acceptance:** Charts render with mock data, campaign creation form exists, links table shows attribution
  - **Verify:** No Recharts console errors, responsive charts

- [ ] **Task 5: Payouts + Admin + AI Support**
  - `pages/payouts/Earnings.tsx` (balance cards, tier progress: $0/$500 free threshold)
  - `pages/payouts/PayoutHistory.tsx` (table with status badges)
  - `pages/admin/Users.tsx` (searchable table, ban/role actions)
  - `pages/admin/Moderation.tsx` (flagged content queue with approve/reject)
  - `pages/admin/Disputes.tsx` (dispute cards with evidence sidebar)
  - `pages/support/Chat.tsx` (chat interface with mock AI responses + escalation button)
  - **Acceptance:** Earnings show correct tier math, admin tables are sortable/searchable, chat accepts messages and responds
  - **Verify:** Interactive elements work (buttons, modals, chat input)

### Phase 3: Integration & Polish
- [ ] **Task 6: Final Integration**
  - Ensure all imports resolve, no type errors
  - Verify responsive behavior across breakpoints
  - Ensure no linter errors
  - **Acceptance:** `npm run build` succeeds, `npm run typecheck` passes
  - **Verify:** Manual click-through of all major routes

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Parallel agents conflict on shared files | Med | Task 1 creates all stubs; Tasks 2-5 only edit files in their assigned page directories and avoid changing shared primitives |
| Mock data becomes inconsistent across pages | Low | Centralize types + mock data generators in Task 1 |
| Build fails due to missing TS types | Low | Strict `tsconfig` + `typecheck` after each task |

## Parallelization Plan

```
Task 1 (Scaffold)
   │
   ├──→ Task 2 (Auth + Dashboard + Notifications)
   ├──→ Task 3 (Briefs + Videos)
   ├──→ Task 4 (Campaigns + Performance + Shopify)
   ├──→ Task 5 (Payouts + Admin + AI Support)
   │
Task 6 (Integration)
```

## Open Questions

- None at this time — proceeding based on frontend SPEC.
