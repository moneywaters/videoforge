import type { Metadata } from 'next';
import { UserNav } from '@/components/layout/user-nav';

export const metadata: Metadata = {
  title: 'VideoForge Dashboard',
  description: 'AI-powered video generation platform',
  robots: { index: false, follow: false }
};

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex min-h-screen">
      <aside className="w-64 shrink-0 border-r bg-background p-4">
        <nav className="space-y-4">
          <a href="/dashboard" className="block rounded-lg px-3 py-2 hover:bg-accent">Dashboard</a>
          <a href="/dashboard/briefs" className="block rounded-lg px-3 py-2 hover:bg-accent">Briefs</a>
          <a href="/dashboard/videos" className="block rounded-lg px-3 py-2 hover:bg-accent">Videos</a>
          <a href="/dashboard/campaigns" className="block rounded-lg px-3 py-2 hover:bg-accent">Campaigns</a>
          <a href="/dashboard/earnings" className="block rounded-lg px-3 py-2 hover:bg-accent">Earnings</a>
          <a href="/dashboard/leaderboard" className="block rounded-lg px-3 py-2 hover:bg-accent">Leaderboard</a>
        </nav>
      </aside>
      <div className="flex flex-1 flex-col">
        <header className="flex h-14 items-center justify-between border-b bg-background px-6">
          <h1 className="text-lg font-semibold">Dashboard</h1>
          <UserNav />
        </header>
        <main className="flex-1 overflow-auto p-6">{children}</main>
      </div>
    </div>
  );
}
