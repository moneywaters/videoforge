import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Badge } from '@/components/ui/Badge';
import { Card, CardBody } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import { Tabs } from '@/components/ui/Tabs';
import type { Video } from '@/types/index';

const videoStatusBadgeVariant = (status: Video['status']) => {
  switch (status) {
    case 'approved': return 'success';
    case 'submitted': return 'info';
    case 'rejected': return 'danger';
    case 'needs_revision': return 'warning';
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
      <div className="aspect-video bg-gray-100 rounded-t-lg flex items-center justify-center">
        <svg className="w-12 h-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </div>
      <CardBody className="space-y-2">
        <div className="flex justify-between items-start">
          <h3 className="font-semibold text-gray-900 truncate">{video.briefTitle}</h3>
          <Badge variant={statusVariant}>{statusLabel}</Badge>
        </div>
        <p className="text-sm text-gray-500">By {video.editorName}</p>
        <p className="text-xs text-gray-400">
          {video.duration}s • {video.resolution} • {submittedDate}
        </p>
      </CardBody>
    </Card>
  );
}

function VideoCardSkeleton() {
  return (
    <Card>
      <Skeleton className="aspect-video rounded-t-lg" />
      <CardBody className="space-y-2">
        <Skeleton className="h-5 w-3/4" />
        <Skeleton className="h-4 w-1/2" />
        <Skeleton className="h-3 w-1/3" />
      </CardBody>
    </Card>
  );
}

export function VideoList() {
  const navigate = useNavigate();
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
    navigate(`/videos/${videoId}`);
  };

  const handleBriefClick = (briefId: string) => {
    navigate(`/briefs/${briefId}`);
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold text-gray-900">Videos</h1>
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
      <h1 className="text-2xl font-bold text-gray-900">Videos</h1>

      <Tabs tabs={tabs} activeTab={activeTab} onChange={setActiveTab}>
        {activeTab === 'available' ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {openBriefs.map((brief) => (
              <Card
                key={brief.id}
                className="hover:shadow-md transition-shadow cursor-pointer"
                onClick={() => handleBriefClick(brief.id)}
              >
                <CardBody className="space-y-2">
                  <h3 className="font-semibold text-gray-900 truncate">{brief.title}</h3>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-gray-500">Bounty</span>
                    <span className="font-medium text-gray-900">${brief.bountyBudget.toLocaleString()}</span>
                  </div>
                  <Badge variant="success">{brief.currentSubmissions}/{brief.submissionLimit} submitted</Badge>
                </CardBody>
              </Card>
            ))}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {mySubmissions.map((video) => (
              <VideoCard key={video.id} video={video} onClick={() => handleVideoClick(video.id)} />
            ))}
          </div>
        )}
      </Tabs>
    </div>
  );
}