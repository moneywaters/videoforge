'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useQuery } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';
import type { Brief, Video } from '@/types';

const videoStatusBadgeVariant = (status: Video['status']) => {
  switch (status) {
    case 'approved': return 'default';
    case 'submitted': return 'secondary';
    case 'rejected': return 'destructive';
    case 'needs_revision': return 'secondary';
    default: return 'default';
  }
};

function VideoCard({ video, onClick }: { video: Video; onClick: () => void }) {
  const statusVariant = videoStatusBadgeVariant(video.status);
  const statusLabel = video.status === 'needs_revision' 
    ? 'Revision Requested' 
    : video.status.charAt(0).toUpperCase() + video.status.slice(1);

  const submittedDate = new Date(video.submittedAt).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });

  return (
    <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={onClick}>
      <div className="aspect-video bg-muted rounded-t-lg flex items-center justify-center">
        <Icons.video className="h-12 w-12 text-muted-foreground" />
      </div>
      <CardContent className="space-y-2">
        <div className="flex justify-between items-start">
          <h3 className="font-semibold truncate">{video.briefTitle}</h3>
          <Badge variant={statusVariant}>{statusLabel}</Badge>
        </div>
        <p className="text-sm text-muted-foreground">By {video.editorName}</p>
        <p className="text-xs text-muted-foreground/60">
          {video.duration}s &bull; {video.resolution} &bull; {submittedDate}
        </p>
      </CardContent>
    </Card>
  );
}

function VideoCardSkeleton() {
  return (
    <Card>
      <Skeleton className="aspect-video rounded-t-lg" />
      <CardContent className="space-y-2">
        <Skeleton className="h-5 w-3/4" />
        <Skeleton className="h-4 w-1/2" />
        <Skeleton className="h-3 w-1/3" />
      </CardContent>
    </Card>
  );
}

function BriefCard({ brief, onClick }: { brief: Brief; onClick: () => void }) {
  return (
    <Card className="hover:shadow-md transition-shadow cursor-pointer" onClick={onClick}>
      <CardContent className="space-y-2">
        <h3 className="font-semibold truncate">{brief.title}</h3>
        <div className="flex items-center justify-between text-sm">
          <span className="text-muted-foreground">Bounty</span>
          <span className="font-medium">${brief.bountyBudget.toLocaleString()}</span>
        </div>
        <Badge variant="default">{brief.currentSubmissions}/{brief.submissionLimit} submitted</Badge>
      </CardContent>
    </Card>
  );
}

export default function VideoListPage() {
  const router = useRouter();
  const [activeTab, setActiveTab] = useState('available');

  const { data: videos, isLoading: videosLoading } = useQuery({
    queryKey: ['videos'],
    queryFn: () => api.getVideos(),
  });

  const { data: briefs, isLoading: briefsLoading } = useQuery({
    queryKey: ['briefs'],
    queryFn: () => api.getBriefs(),
  });

  const isLoading = videosLoading || briefsLoading;

  const openBriefs = briefs?.filter((b) => b.status === 'published') ?? [];
  const submittedVideos = videos ?? [];
  const mySubmissions = submittedVideos.filter((v) => v.editorId === 'usr-editor-001');

  const editorTabs = [
    { id: 'available', label: `Available Briefs (${openBriefs.length})` },
    { id: 'submissions', label: `My Submissions (${mySubmissions.length})` },
  ];

  const tabs = editorTabs;

  const handleVideoClick = (videoId: string) => {
    router.push(`/dashboard/videos/${videoId}`);
  };

  const handleBriefClick = (briefId: string) => {
    router.push(`/dashboard/briefs/${briefId}`);
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">Videos</h1>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[...Array(6)].map((_, i) => (
            <VideoCardSkeleton key={i} />
          ))}
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Videos</h1>

      <Tabs value={activeTab} onValueChange={setActiveTab}>
        <TabsList>
          <TabsTrigger value="available">Available Briefs ({openBriefs.length})</TabsTrigger>
          <TabsTrigger value="submissions">My Submissions ({mySubmissions.length})</TabsTrigger>
        </TabsList>
        <TabsContent value="available">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {openBriefs.length > 0 ? (
              openBriefs.map((brief) => (
                <BriefCard key={brief.id} brief={brief} onClick={() => handleBriefClick(brief.id)} />
              ))
            ) : (
              <div className="col-span-full text-center py-12 text-muted-foreground">
                No briefs available yet. Check back later!
              </div>
            )}
          </div>
        </TabsContent>
        <TabsContent value="submissions">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {mySubmissions.length > 0 ? (
              mySubmissions.map((video) => (
                <VideoCard key={video.id} video={video} onClick={() => handleVideoClick(video.id)} />
              ))
            ) : (
              <div className="col-span-full text-center py-12 text-muted-foreground">
                No submissions yet. Start by picking up a brief!
              </div>
            )}
          </div>
        </TabsContent>
      </Tabs>
    </div>
  );
}