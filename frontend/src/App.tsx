import { BrowserRouter, Navigate, Outlet, Route, Routes, useLocation } from "react-router-dom";
import { useEffect } from "react";
import { Shell } from "@/components/layout/Shell.tsx";
import { useAuthStore } from "@/stores/authStore.ts";

import Login from "@/pages/auth/Login.tsx";
import Register from "@/pages/auth/Register.tsx";
import Onboarding from "@/pages/auth/Onboarding.tsx";
import Dashboard from "@/pages/dashboard/Dashboard.tsx";
import { BriefList } from "@/pages/briefs/BriefList.tsx";
import { BriefDetail } from "@/pages/briefs/BriefDetail.tsx";
import { BriefCreate } from "@/pages/briefs/BriefCreate.tsx";
import { VideoList } from "@/pages/videos/VideoList.tsx";
import { VideoDetail } from "@/pages/videos/VideoDetail.tsx";
import { Leaderboard } from "@/pages/videos/Leaderboard.tsx";
import CampaignList from "@/pages/campaigns/CampaignList.tsx";
import CampaignDetail from "@/pages/campaigns/CampaignDetail.tsx";
import Analytics from "@/pages/performance/Analytics.tsx";
import Leaderboards from "@/pages/performance/Leaderboards.tsx";
import Stores from "@/pages/shopify/Stores.tsx";
import Links from "@/pages/shopify/Links.tsx";
import Earnings from "@/pages/payouts/Earnings.tsx";
import PayoutHistory from "@/pages/payouts/PayoutHistory.tsx";
import Users from "@/pages/admin/Users.tsx";
import Moderation from "@/pages/admin/Moderation.tsx";
import Disputes from "@/pages/admin/Disputes.tsx";
import Chat from "@/pages/support/Chat.tsx";
import NotFound from "@/pages/NotFound.tsx";

const MOCK_USER = {
  id: "mock-user-id",
  email: "dev@videoforge.local",
  name: "Dev User",
  role: "client" as const,
  avatar: undefined,
  createdAt: new Date().toISOString(),
  onboardingComplete: true,
};

function MockUserProvider({ children }: { children: React.ReactNode }) {
  const user = useAuthStore((state) => state.user);
  const setUser = useAuthStore((state) => state.setUser);

  useEffect(() => {
    if (user === null) {
      setUser(MOCK_USER);
    }
  }, [user, setUser]);

  return <>{children}</>;
}

function RequireAuth() {
  const user = useAuthStore((state) => state.user);
  const location = useLocation();

  if (!user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return (
    <Shell>
      <Outlet />
    </Shell>
  );
}

function AppRoutes() {
  return (
    <Routes>
      {/* Public auth routes — no Shell */}
      <Route path="/login" element={<Login />} />
      <Route path="/register" element={<Register />} />
      <Route path="/onboarding" element={<Onboarding />} />

      {/* Authenticated routes — wrapped in Shell */}
      <Route element={<RequireAuth />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/briefs" element={<BriefList />} />
        <Route path="/briefs/:id" element={<BriefDetail />} />
        <Route path="/briefs/new" element={<BriefCreate />} />
        <Route path="/videos" element={<VideoList />} />
        <Route path="/videos/:id" element={<VideoDetail />} />
        <Route path="/leaderboard" element={<Leaderboard />} />
        <Route path="/campaigns" element={<CampaignList />} />
        <Route path="/campaigns/:id" element={<CampaignDetail />} />
        <Route path="/performance" element={<Analytics />} />
        <Route path="/leaderboards" element={<Leaderboards />} />
        <Route path="/shopify/stores" element={<Stores />} />
        <Route path="/shopify/links" element={<Links />} />
        <Route path="/earnings" element={<Earnings />} />
        <Route path="/payouts" element={<PayoutHistory />} />
        <Route path="/admin/users" element={<Users />} />
        <Route path="/admin/moderation" element={<Moderation />} />
        <Route path="/admin/disputes" element={<Disputes />} />
        <Route path="/support" element={<Chat />} />
      </Route>

      {/* Catch-all */}
      <Route path="*" element={<NotFound />} />
    </Routes>
  );
}

export default function App() {
  return (
    <BrowserRouter>
      <MockUserProvider>
        <AppRoutes />
      </MockUserProvider>
    </BrowserRouter>
  );
}
