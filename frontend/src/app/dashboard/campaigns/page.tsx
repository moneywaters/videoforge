'use client';

import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';

export default function CampaignListPage() {
  const { data: campaigns, isLoading } = useQuery({
    queryKey: ['campaigns'],
    queryFn: () => api.getCampaigns(),
  });

  if (isLoading) {
    return (
      <div className="space-y-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-muted rounded w-48"></div>
          <div className="h-64 bg-muted rounded"></div>
        </div>
      </div>
    );
  }

  const activeCampaigns = campaigns?.filter((c) => c.status === 'active') ?? [];
  const pausedCampaigns = campaigns?.filter((c) => c.status === 'paused') ?? [];
  const draftCampaigns = campaigns?.filter((c) => c.status === 'draft') ?? [];
  const endedCampaigns = campaigns?.filter((c) => c.status === 'ended') ?? [];

  const allCampaigns = campaigns ?? [];
  const hasAnyCampaigns = allCampaigns.length > 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Campaigns</h1>
          <p className="text-muted-foreground mt-1">Manage your ad campaigns</p>
        </div>
        <Button>
          <Icons.add className="mr-2 h-4 w-4" />
          New Campaign
        </Button>
      </div>

      {hasAnyCampaigns ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardContent>
              <div className="flex items-center gap-3">
                <div className="p-2 bg-emerald-100 dark:bg-emerald-900 rounded-lg">
                  <Icons.check className="h-5 w-5 text-emerald-600" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Active</p>
                  <p className="text-2xl font-semibold">{activeCampaigns.length}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent>
              <div className="flex items-center gap-3">
                <div className="p-2 bg-amber-100 dark:bg-amber-900 rounded-lg">
                  <Icons.clock className="h-5 w-5 text-amber-600" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Paused</p>
                  <p className="text-2xl font-semibold">{pausedCampaigns.length}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent>
              <div className="flex items-center gap-3">
                <div className="p-2 bg-blue-100 dark:bg-blue-900 rounded-lg">
                  <Icons.edit className="h-5 w-5 text-blue-600" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Draft</p>
                  <p className="text-2xl font-semibold">{draftCampaigns.length}</p>
                </div>
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardContent>
              <div className="flex items-center gap-3">
                <div className="p-2 bg-muted rounded-lg">
                  <Icons.check className="h-5 w-5 text-muted-foreground" />
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Ended</p>
                  <p className="text-2xl font-semibold">{endedCampaigns.length}</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      ) : (
        <Card>
          <CardContent className="text-center py-12">
            <Icons.video className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-semibold mb-2">No campaigns yet</h3>
            <p className="text-muted-foreground mb-4">
              Create your first campaign to start promoting your videos.
            </p>
            <Button>
              <Icons.add className="mr-2 h-4 w-4" />
              Create Campaign
            </Button>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>All Campaigns</CardTitle>
          <CardDescription>View and manage all your campaigns</CardDescription>
        </CardHeader>
        <CardContent>
          {hasAnyCampaigns ? (
            <div className="space-y-4">
              {allCampaigns.map((campaign) => (
                <div
                  key={campaign.id}
                  className="flex items-center justify-between p-4 border rounded-lg hover:bg-muted/50 cursor-pointer"
                >
                  <div className="flex items-center gap-4">
                    <div className="p-2 bg-muted rounded-lg">
                      <Icons.video className="h-5 w-5" />
                    </div>
                    <div>
                      <p className="font-medium">{campaign.videoTitle}</p>
                      <p className="text-sm text-muted-foreground capitalize">{campaign.platform} &bull; {campaign.status}</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-medium">${campaign.spent.toLocaleString()} / ${campaign.budget.toLocaleString()}</p>
                    <p className="text-sm text-muted-foreground">{campaign.status}</p>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-center py-8 text-muted-foreground">
              No campaigns found. Start by creating your first campaign!
            </p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}