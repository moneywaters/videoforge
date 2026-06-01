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
} from '@/types';

const BASE_URL =
  process.env.NEXT_PUBLIC_API_URL ||
  'https://videoforge-gateway.fly.dev/api/v1';

export class TimeoutError extends Error {
  constructor(message = 'Request timed out') {
    super(message);
    this.name = 'TimeoutError';
  }
}

async function fetchWithAuth(url: string, options: RequestInit & { timeout?: number } = {}) {
  const { timeout = 10000, ...fetchOptions } = options;
  const token =
    typeof window !== 'undefined' ? localStorage.getItem('token') : null;
  const headers = new Headers(fetchOptions.headers);
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }
  if (!headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const controller = new AbortController();
  const timeoutId = setTimeout(() => controller.abort(), timeout);

  try {
    const response = await fetch(`${BASE_URL}${url}`, {
      ...fetchOptions,
      headers,
      signal: controller.signal,
    });
    clearTimeout(timeoutId);

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.message || `HTTP error ${response.status}`);
    }

    if (response.status === 204) {
      return;
    }

    return response.json();
  } catch (error) {
    clearTimeout(timeoutId);
    if (error instanceof Error && error.name === 'AbortError') {
      throw new TimeoutError(`Request timed out after ${timeout}ms. Please check your connection and try again.`);
    }
    throw error;
  }
}

function transformBriefResponse(b: any): Brief {
  return {
    id: b.id ?? b.ID,
    title: b.title ?? '',
    description: b.description ?? '',
    status: b.status ?? 'draft',
    clientId: typeof b.client_id === 'string' ? b.client_id : b.clientId,
    clientName: b.client_name ?? b.clientName ?? 'Client',
    bountyBudget:
      typeof b.bounty_budget === 'number' ? b.bounty_budget : b.bountyBudget ?? 0,
    submissionLimit: b.submission_limit ?? b.submissionLimit ?? 0,
    currentSubmissions: b.current_submissions ?? b.currentSubmissions ?? 0,
    tags: b.tags ?? [],
    createdAt: b.created_at ?? b.createdAt ?? new Date().toISOString(),
    deadline: b.deadline,
  };
}

export const api = {
  login: async (email: string, password: string): Promise<User> => {
    const data = await fetchWithAuth('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
      timeout: 60000, // 60 seconds for login (NeonDB cold-start)
    });
    localStorage.setItem('token', data.token);
    return data.user;
  },

  register: async (
    email: string,
    password: string,
    firstName: string,
    lastName: string,
    role: string
  ): Promise<void> => {
    await fetchWithAuth('/auth/register', {
      method: 'POST',
      body: JSON.stringify({
        email,
        password,
        first_name: firstName,
        last_name: lastName,
        role,
      }),
    });
  },

  logout: async (): Promise<void> => {
    try {
      await fetchWithAuth('/auth/logout', { method: 'POST' }).catch(() => {});
    } finally {
      localStorage.removeItem('token');
    }
  },

  loginWithGoogle: async (): Promise<void> => {
    window.location.href = `${BASE_URL}/auth/google/login`;
  },

  registerWithGoogle: async (): Promise<void> => {
    window.location.href = `${BASE_URL}/auth/google/login`;
  },

  getUploadUrl: async (
    briefId: string,
    filename: string,
    contentType: string
  ): Promise<{ url: string; uploadId: string }> => {
    const resp = await fetchWithAuth(`/briefs/${briefId}/raw-footage/upload-url`, {
      method: 'POST',
      body: JSON.stringify({ filename, contentType }),
    });
    return {
      url: resp.upload_url,
      uploadId: resp.storj_key,
    };
  },

  confirmUpload: async (briefId: string, uploadId: string): Promise<void> => {
    return fetchWithAuth(`/briefs/${briefId}/raw-footage/confirm`, {
      method: 'POST',
      body: JSON.stringify({ storj_key: uploadId }),
    });
  },

  getDownloadUrl: async (briefId: string): Promise<{ download_url: string; expires_in: number }> => {
    return fetchWithAuth(`/briefs/${briefId}/raw-footage/download-url`, {
      method: 'GET',
    });
  },

  uploadToPresignedUrl: async (
    url: string,
    file: File,
    options?: {
      onProgress?: (loaded: number, total: number) => void;
      signal?: AbortSignal;
    }
  ): Promise<void> => {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest();
      xhr.open('PUT', url);
      xhr.setRequestHeader('Content-Type', file.type);

      let onAbort: (() => void) | undefined;
      if (options?.signal) {
        if (options.signal.aborted) {
          reject(new Error('Upload cancelled'));
          return;
        }
        onAbort = () => {
          xhr.abort();
          reject(new Error('Upload cancelled'));
        };
        options.signal.addEventListener('abort', onAbort, { once: true });
      }

      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable && options?.onProgress) {
          options.onProgress(e.loaded, e.total);
        }
      };

      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve();
        } else {
          reject(new Error(`Upload failed: ${xhr.status}`));
        }
      };

      xhr.onerror = () => reject(new Error('Upload failed'));
      xhr.onloadend = () => {
        if (onAbort && options?.signal) {
          options.signal.removeEventListener('abort', onAbort);
        }
      };
      xhr.send(file);
    });
  },

  getCurrentUser: async (): Promise<User | null> => {
    if (typeof window === 'undefined') return null;
    const token = localStorage.getItem('token');
    if (!token) return null;
    try {
      const data = await fetchWithAuth('/auth/me');
      return data.user;
    } catch {
      return null;
    }
  },

  getBriefs: async (): Promise<Brief[]> => {
    const resp = await fetchWithAuth('/briefs');
    const briefs = Array.isArray(resp) ? resp : resp.briefs ?? [];
    return briefs.map(transformBriefResponse);
  },

  getBrief: async (id: string): Promise<Brief | undefined> => {
    const resp = await fetchWithAuth(`/briefs/${id}`);
    return transformBriefResponse(resp);
  },

  createBrief: async (
    brief: Omit<Brief, 'id' | 'createdAt' | 'currentSubmissions'>
  ): Promise<Brief> => {
    const resp = await fetchWithAuth('/briefs', {
      method: 'POST',
      body: JSON.stringify({
        title: brief.title,
        description: brief.description,
        goals: brief.description,
        target_audience: '',
        tone: '',
        style_preferences: '',
        cta: '',
        bounty_budget: brief.bountyBudget,
        submissions_limit: brief.submissionLimit,
        is_blind: false,
        tags: brief.tags,
      }),
    });
    return transformBriefResponse(resp);
  },

  updateBrief: async (
    id: string,
    updates: Partial<Brief>
  ): Promise<Brief> => {
    const body: Record<string, unknown> = {};
    if (updates.title !== undefined) body.title = updates.title;
    if (updates.description !== undefined)
      body.description = updates.description;
    if (updates.bountyBudget !== undefined)
      body.bounty_budget = updates.bountyBudget;
    if (updates.submissionLimit !== undefined)
      body.submissions_limit = updates.submissionLimit;
    if (updates.tags !== undefined) body.tags = updates.tags;
    const resp = await fetchWithAuth(`/briefs/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(body),
    });
    return transformBriefResponse(resp);
  },

  getVideos: async (): Promise<Video[]> => {
    // TODO: replace with real endpoint
    return [];
  },

  getVideo: async (_id: string): Promise<Video | undefined> => {
    return undefined;
  },

  getCampaigns: async (): Promise<Campaign[]> => {
    return [];
  },

  getCampaign: async (_id: string): Promise<Campaign | undefined> => {
    return undefined;
  },

  createCampaign: async (
    _campaign: Omit<Campaign, 'id' | 'startedAt' | 'spent'>
  ): Promise<Campaign> => {
    throw new Error('Not implemented');
  },

  getPerformance: async (_videoId?: string): Promise<PerformanceMetric[]> => {
    return [];
  },

  getPlatformPerformance: async (): Promise<PlatformPerformance[]> => {
    const platforms: AdPlatform[] = ['meta', 'tiktok', 'google'];
    return platforms.map((platform) => ({
      platform,
      totalSales: 0,
      totalRevenue: 0,
      totalSpend: 0,
      roas: 0,
    }));
  },

  getLeaderboard: async (): Promise<LeaderboardEntry[]> => {
    return [];
  },

  getPayouts: async (_userId: string): Promise<Payout[]> => {
    return [];
  },

  getEarnings: async (_userId: string): Promise<EarningsSummary> => {
    return {
      userId: '',
      totalEarned: 0,
      totalPaidOut: 0,
      pendingBalance: 0,
      lifetimeSales: 0,
      tierProgress: 0,
      feeRate: 0,
    };
  },

  getUsers: async (): Promise<User[]> => {
    return [];
  },

  updateUserRole: async (_userId: string, _role: User['role']): Promise<User> => {
    throw new Error('Not implemented');
  },

  getNotifications: async (_userId: string): Promise<Notification[]> => {
    return [];
  },

  getStores: async (): Promise<ShopifyStore[]> => {
    return [];
  },

  getLinks: async (): Promise<VideoLink[]> => {
    return [];
  },

  getDisputes: async (): Promise<Dispute[]> => {
    return [];
  },

  getModerationQueue: async (): Promise<ModerationItem[]> => {
    return [];
  },

  sendChatMessage: async (_content: string): Promise<ChatMessage> => {
    throw new Error('Not implemented');
  },

  getOnboardingSteps: async (): Promise<OnboardingStep[]> => {
    return [];
  },

  updateVideoStatus: async (_videoId: string, _status: VideoStatus, _feedback?: string): Promise<Video> => {
    throw new Error('Not implemented');
  },
};
