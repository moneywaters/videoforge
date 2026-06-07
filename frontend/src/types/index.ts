export type UserRole =
  | 'client'
  | 'editor'
  | 'ad_specialist';

export interface User {
  id: string;
  email: string;
  name: string;
  role: UserRole;
  avatar?: string;
  createdAt: string;
  onboardingComplete: boolean;
}

export type BriefStatus = 'draft' | 'published' | 'closed';

export interface Brief {
  id: string;
  title: string;
  description: string;
  status: BriefStatus;
  clientId: string;
  clientName: string;
  bountyBudget: number;
  submissionLimit: number;
  currentSubmissions: number;
  tags: string[];
  createdAt: string;
  deadline?: string;
  has_raw_footage?: boolean;
  raw_footage_filename?: string;
  raw_footage_storj_key?: string;
  updated_at?: string;
}

export type VideoStatus =
  | 'submitted'
  | 'approved'
  | 'rejected'
  | 'needs_revision';

export interface VideoRevision {
  id: string;
  version: number;
  url: string;
  createdAt: string;
  notes: string;
}

export interface Video {
  id: string;
  briefId: string;
  briefTitle: string;
  editorId: string;
  editorName: string;
  status: VideoStatus;
  url: string;
  thumbnail?: string;
  duration: number;
  resolution: string;
  submittedAt: string;
  revisions: VideoRevision[];
  feedback?: string;
}

export interface LeaderboardEntry {
  rank: number;
  editorName: string;
  videoCount: number;
  totalSales: number;
  revenue: number;
}

export type CampaignStatus = 'draft' | 'active' | 'paused' | 'ended';
export type AdPlatform = 'meta' | 'tiktok' | 'google';

export interface Campaign {
  id: string;
  videoId: string;
  videoTitle: string;
  specialistId: string;
  specialistName: string;
  platform: AdPlatform;
  budget: number;
  spent: number;
  status: CampaignStatus;
  startedAt: string;
  endedAt?: string;
  targetCpa: number;
}

export interface PerformanceMetric {
  date: string;
  videoId: string;
  sales: number;
  revenue: number;
  conversions: number;
  roas: number;
}

export interface PlatformPerformance {
  platform: AdPlatform;
  totalSales: number;
  totalRevenue: number;
  totalSpend: number;
  roas: number;
}

export interface Payout {
  id: string;
  userId: string;
  userName: string;
  amount: number;
  fee: number;
  netAmount: number;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  createdAt: string;
  processedAt?: string;
}

export interface EarningsSummary {
  userId: string;
  totalEarned: number;
  totalPaidOut: number;
  pendingBalance: number;
  lifetimeSales: number;
  tierProgress: number;
  feeRate: number;
}

export interface Notification {
  id: string;
  userId: string;
  event: string;
  message: string;
  read: boolean;
  createdAt: string;
}

export interface ShopifyStore {
  id: string;
  name: string;
  domain: string;
  connectedAt: string;
  status: 'active' | 'disconnected';
}

export interface VideoLink {
  id: string;
  videoId: string;
  campaignId: string;
  url: string;
  discountCode?: string;
  utmSource: string;
  utmMedium: string;
  utmCampaign: string;
  clicks: number;
  conversions: number;
  revenue: number;
}

export interface Dispute {
  id: string;
  reporterId: string;
  reporterName: string;
  targetId: string;
  targetName: string;
  reason: string;
  evidence: string[];
  status: 'open' | 'resolved';
  createdAt: string;
  resolvedAt?: string;
  resolution?: string;
}

export interface ModerationItem {
  id: string;
  type: 'video' | 'brief';
  contentId: string;
  flaggedBy: string;
  reason: string;
  status: 'pending' | 'approved' | 'rejected';
  createdAt: string;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  timestamp: string;
}

export interface OnboardingStep {
  id: string;
  title: string;
  description: string;
  completed: boolean;
}

export interface NavItem {
  title: string;
  url: string;
  icon?: string;
  shortcut?: string[];
  isActive?: boolean;
  items?: NavItem[];
  access?: {
    requireOrg?: boolean;
    permission?: string;
    role?: string;
    plan?: string;
    feature?: string;
  };
}

export interface NavGroup {
  label: string;
  items: NavItem[];
}
