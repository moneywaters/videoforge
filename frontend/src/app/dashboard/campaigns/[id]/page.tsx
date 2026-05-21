'use client';

import { useParams, useRouter } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';

export default function CampaignDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const { data: campaign, isLoading } = useQuery({
    queryKey: ['campaign', id],
    queryFn: () => api.getCampaign(id),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-24" />
        <Skeleton className="h-64 w-full" />
      </div>
    );
  }

  if (!campaign) {
    return (
      <div className="text-center py-12">
        <p className="text-muted-foreground">Campaign not found</p>
        <Button className="mt-4" onClick={() => router.push('/dashboard/campaigns')}>
          Back to Campaigns
        </Button>
      </div>
    );
  }

  const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'long',
      day: 'numeric',
      year: 'numeric',
    });
  };

  return (
    <div className="space-y-6">
      <Button variant="ghost" onClick={() => router.push('/dashboard/campaigns')}>
        <Icons.chevronLeft className="mr-2 h-4 w-4" />
        Back to Campaigns
      </Button>

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{campaign.videoTitle}</h1>
          <p className="text-muted-foreground mt-1">
            {campaign.platform.toUpperCase()} &bull; {campaign.status}
          </p>
        </div>
        <div className="flex gap-2">
          {campaign.status === 'active' && (
            <Button variant="secondary">Pause</Button>
          )}
          {campaign.status === 'paused' && (
            <Button>Resume</Button>
          )}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <Card>
          <CardContent>
            <p className="text-sm text-muted-foreground">Budget</p>
            <p className="text-2xl font-semibold">${campaign.budget.toLocaleString()}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <p className="text-sm text-muted-foreground">Spent</p>
            <p className="text-2xl font-semibold">${campaign.spent.toLocaleString()}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <p className="text-sm text-muted-foreground">Target CPA</p>
            <p className="text-2xl font-semibold">${campaign.targetCpa}</p>
          </CardContent>
        </Card>
        <Card>
          <CardContent>
            <p className="text-sm text-muted-foreground">Started</p>
            <p className="text-2xl font-semibold">{formatDate(campaign.startedAt)}</p>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Campaign Performance</CardTitle>
          <CardDescription>Track how your ad is performing</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-64 flex items-center justify-center bg-muted rounded-lg">
            <div className="text-center">
              <Icons.trendingUp className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <p className="text-muted-foreground">Chart placeholder</p>
              <p className="text-sm text-muted-foreground/60">Campaign performance chart will appear here</p>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}