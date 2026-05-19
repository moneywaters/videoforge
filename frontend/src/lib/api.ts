import type {
  User,
  Brief,
  Video,
  Campaign,
  PerformanceMetric,
  PlatformPerformance,
  Payout,
  EarningsSummary,
  Notification,
  ShopifyStore,
  VideoLink,
  Dispute,
  ModerationItem,
  ChatMessage,
  OnboardingStep,
  LeaderboardEntry,
  VideoStatus,
  AdPlatform,
} from '@/types/index';

// Helper to generate UUIDs
const uuid = () => 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
  const r = Math.random() * 16 | 0;
  const v = c === 'x' ? r : (r & 0x3 | 0x8);
  return v.toString(16);
});

// Mock data storage
let currentUser: User | null = null;

const mockUsers: User[] = [
  {
    id: 'usr-client-001',
    email: 'client@acme.com',
    name: 'John Smith',
    role: 'client',
    createdAt: '2025-01-15T10:00:00Z',
    onboardingComplete: true,
  },
  {
    id: 'usr-editor-001',
    email: 'jane@editor.com',
    name: 'Jane Doe',
    role: 'editor',
    createdAt: '2025-02-20T14:30:00Z',
    onboardingComplete: true,
  },
  {
    id: 'usr-specialist-001',
    email: 'alex@adpro.com',
    name: 'Alex Martinez',
    role: 'ad_specialist',
    createdAt: '2025-03-10T09:15:00Z',
    onboardingComplete: true,
  },
  {
    id: 'usr-admin-001',
    email: 'admin@videoforge.com',
    name: 'Admin User',
    role: 'admin',
    createdAt: '2024-12-01T08:00:00Z',
    onboardingComplete: true,
  },
  {
    id: 'usr-support-001',
    email: 'support@videoforge.com',
    name: 'AI Support',
    role: 'support_ai',
    createdAt: '2025-04-01T12:00:00Z',
    onboardingComplete: true,
  },
];

const mockBriefs: Brief[] = [
  {
    id: 'brf-001',
    title: 'Summer Sale TikTok Promo',
    description: 'Create an engaging 30-second TikTok promoting our summer sale. Must include product showcase and discount code.',
    status: 'published',
    clientId: 'usr-client-001',
    clientName: 'John Smith',
    bountyBudget: 1500,
    submissionLimit: 10,
    currentSubmissions: 4,
    tags: ['fashion', 'summer', 'sale'],
    createdAt: '2025-05-01T10:00:00Z',
    deadline: '2025-06-15T23:59:59Z',
  },
  {
    id: 'brf-002',
    title: 'Product Launch Reel',
    description: '15-30 second Instagram Reel for new product launch. Show product features and benefits.',
    status: 'published',
    clientId: 'usr-client-001',
    clientName: 'John Smith',
    bountyBudget: 2000,
    submissionLimit: 5,
    currentSubmissions: 2,
    tags: ['tech', 'product', 'launch'],
    createdAt: '2025-05-10T14:00:00Z',
    deadline: '2025-06-20T23:59:59Z',
  },
  {
    id: 'brf-003',
    title: 'Holiday Gift Guide',
    description: 'Create a gift guide video showcasing top products for the holiday season.',
    status: 'published',
    clientId: 'usr-client-001',
    clientName: 'John Smith',
    bountyBudget: 3000,
    submissionLimit: 8,
    currentSubmissions: 6,
    tags: ['holiday', 'gift', 'beauty'],
    createdAt: '2025-04-20T09:00:00Z',
    deadline: '2025-12-01T23:59:59Z',
  },
  {
    id: 'brf-004',
    title: 'Fitness App Promo',
    description: '30-second promotional video for fitness mobile app. Highlight key features and transformations.',
    status: 'closed',
    clientId: 'usr-client-001',
    clientName: 'John Smith',
    bountyBudget: 1200,
    submissionLimit: 5,
    currentSubmissions: 5,
    tags: ['fitness', 'app', 'health'],
    createdAt: '2025-03-15T11:00:00Z',
    deadline: '2025-04-30T23:59:59Z',
  },
  {
    id: 'brf-005',
    title: 'Beauty Tutorial Series',
    description: 'Step-by-step makeup tutorial showcasing our new product line.',
    status: 'published',
    clientId: 'usr-client-001',
    clientName: 'John Smith',
    bountyBudget: 2500,
    submissionLimit: 10,
    currentSubmissions: 3,
    tags: ['beauty', 'tutorial', 'makeup'],
    createdAt: '2025-05-18T16:00:00Z',
    deadline: '2025-07-10T23:59:59Z',
  },
  {
    id: 'brf-006',
    title: 'Tech Review Video',
    description: 'In-depth product review for our latest tech gadget. Must be professional and engaging.',
    status: 'draft',
    clientId: 'usr-client-001',
    clientName: 'John Smith',
    bountyBudget: 500,
    submissionLimit: 3,
    currentSubmissions: 0,
    tags: ['tech', 'review', 'gadget'],
    createdAt: '2025-05-19T08:00:00Z',
  },
];

const mockVideos: Video[] = [
  {
    id: 'vid-001',
    briefId: 'brf-001',
    briefTitle: 'Summer Sale TikTok Promo',
    editorId: 'usr-editor-001',
    editorName: 'Jane Doe',
    status: 'approved',
    url: 'https://cdn.videoforge.com/videos/vid-001.mp4',
    thumbnail: 'https://cdn.videoforge.com/thumbs/vid-001.jpg',
    duration: 30,
    resolution: '1080x1920',
    submittedAt: '2025-05-12T10:00:00Z',
    revisions: [],
  },
  {
    id: 'vid-002',
    briefId: 'brf-001',
    briefTitle: 'Summer Sale TikTok Promo',
    editorId: 'usr-editor-001',
    editorName: 'Jane Doe',
    status: 'submitted',
    url: 'https://cdn.videoforge.com/videos/vid-002.mp4',
    thumbnail: 'https://cdn.videoforge.com/thumbs/vid-002.jpg',
    duration: 25,
    resolution: '1080x1920',
    submittedAt: '2025-05-15T14:30:00Z',
    revisions: [],
  },
  {
    id: 'vid-003',
    briefId: 'brf-002',
    briefTitle: 'Product Launch Reel',
    editorId: 'usr-editor-001',
    editorName: 'Jane Doe',
    status: 'needs_revision',
    url: 'https://cdn.videoforge.com/videos/vid-003.mp4',
    thumbnail: 'https://cdn.videoforge.com/thumbs/vid-003.jpg',
    duration: 20,
    resolution: '1080x1920',
    submittedAt: '2025-05-14T09:00:00Z',
    revisions: [
      {
        id: 'rev-001',
        version: 1,
        url: 'https://cdn.videoforge.com/videos/vid-003-v1.mp4',
        createdAt: '2025-05-14T09:00:00Z',
        notes: 'Initial submission',
      },
    ],
    feedback: 'Please add more product close-ups and improve the transition timing.',
  },
  {
    id: 'vid-004',
    briefId: 'brf-003',
    briefTitle: 'Holiday Gift Guide',
    editorId: 'usr-editor-001',
    editorName: 'Jane Doe',
    status: 'rejected',
    url: 'https://cdn.videoforge.com/videos/vid-004.mp4',
    thumbnail: 'https://cdn.videoforge.com/thumbs/vid-004.jpg',
    duration: 45,
    resolution: '1080x1920',
    submittedAt: '2025-04-25T11:00:00Z',
    revisions: [],
    feedback: 'Video does not meet content guidelines. Please review requirements.',
  },
  {
    id: 'vid-005',
    briefId: 'brf-003',
    briefTitle: 'Holiday Gift Guide',
    editorId: 'usr-editor-001',
    editorName: 'Jane Doe',
    status: 'approved',
    url: 'https://cdn.videoforge.com/videos/vid-005.mp4',
    thumbnail: 'https://cdn.videoforge.com/thumbs/vid-005.jpg',
    duration: 60,
    resolution: '1080x1920',
    submittedAt: '2025-05-01T15:00:00Z',
    revisions: [],
  },
  {
    id: 'vid-006',
    briefId: 'brf-004',
    briefTitle: 'Fitness App Promo',
    editorId: 'usr-editor-001',
    editorName: 'Jane Doe',
    status: 'approved',
    url: 'https://cdn.videoforge.com/videos/vid-006.mp4',
    thumbnail: 'https://cdn.videoforge.com/thumbs/vid-006.jpg',
    duration: 30,
    resolution: '1080x1920',
    submittedAt: '2025-04-10T12:00:00Z',
    revisions: [],
  },
  {
    id: 'vid-007',
    briefId: 'brf-005',
    briefTitle: 'Beauty Tutorial Series',
    editorId: 'usr-editor-001',
    editorName: 'Jane Doe',
    status: 'submitted',
    url: 'https://cdn.videoforge.com/videos/vid-007.mp4',
    thumbnail: 'https://cdn.videoforge.com/thumbs/vid-007.jpg',
    duration: 180,
    resolution: '1080x1920',
    submittedAt: '2025-05-19T10:00:00Z',
    revisions: [],
  },
  {
    id: 'vid-008',
    briefId: 'brf-005',
    briefTitle: 'Beauty Tutorial Series',
    editorId: 'usr-editor-001',
    editorName: 'Jane Doe',
    status: 'submitted',
    url: 'https://cdn.videoforge.com/videos/vid-008.mp4',
    thumbnail: 'https://cdn.videoforge.com/thumbs/vid-008.jpg',
    duration: 150,
    resolution: '1080x1920',
    submittedAt: '2025-05-19T11:30:00Z',
    revisions: [],
  },
];

const mockCampaigns: Campaign[] = [
  {
    id: 'cmp-001',
    videoId: 'vid-001',
    videoTitle: 'Summer Sale TikTok Promo',
    specialistId: 'usr-specialist-001',
    specialistName: 'Alex Martinez',
    platform: 'tiktok',
    budget: 1000,
    spent: 450,
    status: 'active',
    startedAt: '2025-05-13T00:00:00Z',
    targetCpa: 25,
  },
  {
    id: 'cmp-002',
    videoId: 'vid-005',
    videoTitle: 'Holiday Gift Guide',
    specialistId: 'usr-specialist-001',
    specialistName: 'Alex Martinez',
    platform: 'meta',
    budget: 2000,
    spent: 1800,
    status: 'active',
    startedAt: '2025-05-02T00:00:00Z',
    targetCpa: 30,
  },
  {
    id: 'cmp-003',
    videoId: 'vid-006',
    videoTitle: 'Fitness App Promo',
    specialistId: 'usr-specialist-001',
    specialistName: 'Alex Martinez',
    platform: 'meta',
    budget: 800,
    spent: 800,
    status: 'ended',
    startedAt: '2025-04-11T00:00:00Z',
    endedAt: '2025-04-30T23:59:59Z',
    targetCpa: 20,
  },
  {
    id: 'cmp-004',
    videoId: 'vid-001',
    videoTitle: 'Summer Sale TikTok Promo',
    specialistId: 'usr-specialist-001',
    specialistName: 'Alex Martinez',
    platform: 'meta',
    budget: 1500,
    spent: 200,
    status: 'paused',
    startedAt: '2025-05-14T00:00:00Z',
    targetCpa: 25,
  },
];

// Generate 30 days of performance data
const generatePerformanceData = (): PerformanceMetric[] => {
  const metrics: PerformanceMetric[] = [];
  const baseDate = new Date('2025-04-20');
  
  for (let day = 0; day < 30; day++) {
    const date = new Date(baseDate);
    date.setDate(date.getDate() + day);
    const dateStr = date.toISOString().split('T')[0];
    
    // Generate metrics for each video
    mockVideos.forEach((video) => {
      const baseSales = 5 + Math.floor(Math.random() * 15);
      const revenue = baseSales * (10 + Math.random() * 40);
      const conversions = Math.floor(baseSales * 0.7);
      
      metrics.push({
        date: dateStr,
        videoId: video.id,
        sales: baseSales,
        revenue: Math.round(revenue * 100) / 100,
        conversions,
        roas: 2 + Math.random() * 4,
      });
    });
  }
  
  return metrics;
};

const mockPerformance = generatePerformanceData();

const mockPayouts: Payout[] = [
  {
    id: 'pyt-001',
    userId: 'usr-editor-001',
    userName: 'Jane Doe',
    amount: 1500,
    fee: 75,
    netAmount: 1425,
    status: 'completed',
    createdAt: '2025-04-15T10:00:00Z',
    processedAt: '2025-04-17T14:30:00Z',
  },
  {
    id: 'pyt-002',
    userId: 'usr-editor-001',
    userName: 'Jane Doe',
    amount: 2200,
    fee: 110,
    netAmount: 2090,
    status: 'processing',
    createdAt: '2025-05-10T10:00:00Z',
  },
  {
    id: 'pyt-003',
    userId: 'usr-specialist-001',
    userName: 'Alex Martinez',
    amount: 800,
    fee: 40,
    netAmount: 760,
    status: 'pending',
    createdAt: '2025-05-18T10:00:00Z',
  },
];

const mockNotifications: Notification[] = [
  {
    id: 'ntf-001',
    userId: 'usr-client-001',
    event: 'new_submission',
    message: 'New video submission for Summer Sale TikTok Promo',
    read: false,
    createdAt: '2025-05-19T10:00:00Z',
  },
  {
    id: 'ntf-002',
    userId: 'usr-client-001',
    event: 'video_approved',
    message: 'Your video for Holiday Gift Guide has been approved',
    read: true,
    createdAt: '2025-05-02T14:00:00Z',
  },
  {
    id: 'ntf-003',
    userId: 'usr-editor-001',
    event: 'brief_new',
    message: 'New brief published: Beauty Tutorial Series',
    read: false,
    createdAt: '2025-05-18T16:00:00Z',
  },
  {
    id: 'ntf-004',
    userId: 'usr-editor-001',
    event: 'revision_requested',
    message: 'Revision requested for Product Launch Reel',
    read: false,
    createdAt: '2025-05-14T09:30:00Z',
  },
  {
    id: 'ntf-005',
    userId: 'usr-specialist-001',
    event: 'campaign_active',
    message: 'Your campaign is now live',
    read: true,
    createdAt: '2025-05-13T00:00:00Z',
  },
];

const mockStores: ShopifyStore[] = [
  {
    id: 'str-001',
    name: 'Acme Official Store',
    domain: 'acme-store.myshopify.com',
    connectedAt: '2025-01-20T10:00:00Z',
    status: 'active',
  },
  {
    id: 'str-002',
    name: 'Acme Beauty',
    domain: 'acme-beauty.myshopify.com',
    connectedAt: '2025-02-15T14:30:00Z',
    status: 'active',
  },
  {
    id: 'str-003',
    name: 'Acme Tech',
    domain: 'acme-tech.myshopify.com',
    connectedAt: '2025-03-10T09:00:00Z',
    status: 'disconnected',
  },
];

const mockLinks: VideoLink[] = [
  {
    id: 'lnk-001',
    videoId: 'vid-001',
    campaignId: 'cmp-001',
    url: 'https://acme-store.myshopify.com/summer-sale',
    discountCode: 'SUMMER20',
    utmSource: 'tiktok',
    utmMedium: 'video',
    utmCampaign: 'summer_sale_2025',
    clicks: 1240,
    conversions: 156,
    revenue: 4680,
  },
  {
    id: 'lnk-002',
    videoId: 'vid-005',
    campaignId: 'cmp-002',
    url: 'https://acme-store.myshopify.com/holiday-gift',
    discountCode: 'GIFT25',
    utmSource: 'facebook',
    utmMedium: 'video',
    utmCampaign: 'holiday_gift_2025',
    clicks: 3450,
    conversions: 420,
    revenue: 12600,
  },
];

const mockDisputes: Dispute[] = [
  {
    id: 'dsp-001',
    reporterId: 'usr-client-001',
    reporterName: 'John Smith',
    targetId: 'vid-004',
    targetName: 'Holiday Gift Guide',
    reason: 'Quality not as promised',
    evidence: ['https://example.com/screenshot1.jpg'],
    status: 'resolved',
    createdAt: '2025-04-26T10:00:00Z',
    resolvedAt: '2025-04-28T14:00:00Z',
    resolution: 'Editor agreed to resubmit',
  },
];

const mockModeration: ModerationItem[] = [
  {
    id: 'mod-001',
    type: 'video',
    contentId: 'vid-004',
    flaggedBy: 'usr-admin-001',
    reason: 'Content guideline violation',
    status: 'rejected',
    createdAt: '2025-04-26T10:00:00Z',
  },
  {
    id: 'mod-002',
    type: 'brief',
    contentId: 'brf-002',
    flaggedBy: 'system',
    reason: 'Inappropriate content detected',
    status: 'approved',
    createdAt: '2025-05-11T08:00:00Z',
  },
];

const mockOnboardingSteps: OnboardingStep[] = [
  {
    id: 'step-001',
    title: 'Create your profile',
    description: 'Set up your account with your details and role',
    completed: true,
  },
  {
    id: 'step-002',
    title: 'Connect your store',
    description: 'Link your Shopify store to start receiving briefs',
    completed: true,
  },
  {
    id: 'step-003',
    title: 'Browse briefs',
    description: 'Explore available briefs and submit proposals',
    completed: true,
  },
  {
    id: 'step-004',
    title: 'Upload your first video',
    description: 'Learn how to submit videos for review',
    completed: false,
  },
];

// API functions
export const api = {
  // Auth
  login: async (email: string, _password: string): Promise<User> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    const user = mockUsers.find(u => u.email === email);
    if (!user) {
      throw new Error('Invalid credentials');
    }
    currentUser = user;
    return user;
  },

  register: async (email: string, name: string, role: User['role']): Promise<User> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    const newUser: User = {
      id: `usr-${uuid()}`,
      email,
      name,
      role,
      createdAt: new Date().toISOString(),
      onboardingComplete: false,
    };
    mockUsers.push(newUser);
    currentUser = newUser;
    return newUser;
  },

  logout: async (): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 200));
    currentUser = null;
  },

  getCurrentUser: async (): Promise<User | null> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    return currentUser;
  },

  // Briefs
  getBriefs: async (): Promise<Brief[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockBriefs;
  },

  getBrief: async (id: string): Promise<Brief | undefined> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    return mockBriefs.find(b => b.id === id);
  },

  createBrief: async (brief: Omit<Brief, 'id' | 'createdAt' | 'currentSubmissions'>): Promise<Brief> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    const newBrief: Brief = {
      ...brief,
      id: `brf-${uuid()}`,
      createdAt: new Date().toISOString(),
      currentSubmissions: 0,
    };
    mockBriefs.push(newBrief);
    return newBrief;
  },

  updateBrief: async (id: string, updates: Partial<Brief>): Promise<Brief> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    const brief = mockBriefs.find(b => b.id === id);
    if (!brief) throw new Error('Brief not found');
    Object.assign(brief, updates);
    return brief;
  },

  // Videos
  getVideos: async (): Promise<Video[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockVideos;
  },

  getVideo: async (id: string): Promise<Video | undefined> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    return mockVideos.find(v => v.id === id);
  },

  submitVideo: async (video: Omit<Video, 'id' | 'submittedAt' | 'revisions'>): Promise<Video> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    const newVideo: Video = {
      ...video,
      id: `vid-${uuid()}`,
      submittedAt: new Date().toISOString(),
      revisions: [],
    };
    mockVideos.push(newVideo);
    // Update brief submissions
    const brief = mockBriefs.find(b => b.id === video.briefId);
    if (brief) brief.currentSubmissions++;
    return newVideo;
  },

  updateVideoStatus: async (id: string, status: VideoStatus, feedback?: string): Promise<Video> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    const video = mockVideos.find(v => v.id === id);
    if (!video) throw new Error('Video not found');
    video.status = status;
    if (feedback) video.feedback = feedback;
    return video;
  },

  // Campaigns
  getCampaigns: async (): Promise<Campaign[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockCampaigns;
  },

  getCampaign: async (id: string): Promise<Campaign | undefined> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    return mockCampaigns.find(c => c.id === id);
  },

  createCampaign: async (campaign: Omit<Campaign, 'id' | 'startedAt' | 'spent'>): Promise<Campaign> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    const newCampaign: Campaign = {
      ...campaign,
      id: `cmp-${uuid()}`,
      spent: 0,
      startedAt: new Date().toISOString(),
    };
    mockCampaigns.push(newCampaign);
    return newCampaign;
  },

  // Performance
  getPerformance: async (_videoId?: string): Promise<PerformanceMetric[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockPerformance;
  },

  getPlatformPerformance: async (): Promise<PlatformPerformance[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    const platforms: AdPlatform[] = ['meta', 'tiktok', 'google'];
    return platforms.map(platform => {
      const platformMetrics = mockPerformance.filter(m => {
        const campaign = mockCampaigns.find(c => c.platform === platform);
        return campaign && campaign.videoId === m.videoId;
      });
      return {
        platform,
        totalSales: platformMetrics.reduce((sum, m) => sum + m.sales, 0),
        totalRevenue: platformMetrics.reduce((sum, m) => sum + m.revenue, 0),
        totalSpend: mockCampaigns
          .filter(c => c.platform === platform)
          .reduce((sum, c) => sum + c.spent, 0),
        roas: platformMetrics.length > 0 
          ? platformMetrics.reduce((sum, m) => sum + m.roas, 0) / platformMetrics.length 
          : 0,
      };
    });
  },

  getLeaderboard: async (): Promise<LeaderboardEntry[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return [
      { rank: 1, editorName: 'Jane Doe', videoCount: 12, totalSales: 156, revenue: 15600 },
      { rank: 2, editorName: 'Mike Chen', videoCount: 10, totalSales: 134, revenue: 13400 },
      { rank: 3, editorName: 'Sarah Kim', videoCount: 8, totalSales: 98, revenue: 9800 },
      { rank: 4, editorName: 'Tom Wilson', videoCount: 6, totalSales: 72, revenue: 7200 },
      { rank: 5, editorName: 'Lisa Brown', videoCount: 5, totalSales: 65, revenue: 6500 },
    ];
  },

  // Payouts & Earnings
  getPayouts: async (_userId: string): Promise<Payout[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockPayouts;
  },

  getEarnings: async (userId: string): Promise<EarningsSummary> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    const userPayouts = mockPayouts.filter(p => p.userId === userId);
    const totalEarned = userPayouts.reduce((sum, p) => sum + p.amount, 0);
    const totalPaidOut = userPayouts
      .filter(p => p.status === 'completed')
      .reduce((sum, p) => sum + p.netAmount, 0);
    const pendingBalance = userPayouts
      .filter(p => p.status === 'pending' || p.status === 'processing')
      .reduce((sum, p) => sum + p.netAmount, 0);
    
    return {
      userId,
      totalEarned,
      totalPaidOut,
      pendingBalance,
      lifetimeSales: Math.floor(totalEarned / 100),
      tierProgress: Math.min(500, Math.floor(totalEarned / 50)),
      feeRate: totalEarned > 10000 ? 0 : 0.05,
    };
  },

  // Users
  getUsers: async (): Promise<User[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockUsers;
  },

  updateUserRole: async (userId: string, role: User['role']): Promise<User> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    const user = mockUsers.find(u => u.id === userId);
    if (!user) throw new Error('User not found');
    user.role = role;
    return user;
  },

  // Notifications
  getNotifications: async (userId: string): Promise<Notification[]> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    return mockNotifications.filter(n => n.userId === userId);
  },

  markNotificationRead: async (id: string): Promise<Notification> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    const notification = mockNotifications.find(n => n.id === id);
    if (!notification) throw new Error('Notification not found');
    notification.read = true;
    return notification;
  },

  markAllNotificationsRead: async (userId: string): Promise<void> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    mockNotifications
      .filter(n => n.userId === userId)
      .forEach(n => { n.read = true; });
  },

  // Stores
  getStores: async (): Promise<ShopifyStore[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockStores;
  },

  getLinks: async (): Promise<VideoLink[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockLinks;
  },

  // Disputes
  getDisputes: async (): Promise<Dispute[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockDisputes;
  },

  // Moderation
  getModerationQueue: async (): Promise<ModerationItem[]> => {
    await new Promise(resolve => setTimeout(resolve, 400));
    return mockModeration;
  },

  // Chat
  sendChatMessage: async (content: string): Promise<ChatMessage> => {
    await new Promise(resolve => setTimeout(resolve, 500));
    return {
      id: `msg-${uuid()}`,
      role: 'assistant',
      content,
      timestamp: new Date().toISOString(),
    };
  },

  // Onboarding
  getOnboardingSteps: async (): Promise<OnboardingStep[]> => {
    await new Promise(resolve => setTimeout(resolve, 300));
    return mockOnboardingSteps;
  },
};