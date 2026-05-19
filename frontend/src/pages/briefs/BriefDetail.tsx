import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardBody, CardHeader, CardTitle } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import type { BriefStatus, Video } from '@/types/index';

const statusBadgeVariant = (status: BriefStatus) => {
  switch (status) {
    case 'published': return 'success';
    case 'closed': return 'danger';
    case 'draft': return 'warning';
    default: return 'default';
  }
};

const statusLabels: Record<BriefStatus, string> = {
  draft: 'Draft',
  published: 'Open',
  closed: 'Closed',
};

const videoStatusBadgeVariant = (status: Video['status']) => {
  switch (status) {
    case 'approved': return 'success';
    case 'submitted': return 'info';
    case 'rejected': return 'danger';
    case 'needs_revision': return 'warning';
    default: return 'default';
  }
};

function SubmissionItem({ video }: { video: Video }) {
  const statusLabel = video.status === 'needs_revision' ? 'Revision Requested' : video.status.charAt(0).toUpperCase() + video.status.slice(1);

  return (
    <div className="flex items-center justify-between p-4 border border-gray-200 rounded-lg hover:bg-gray-50 transition-colors">
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-gray-900 truncate">Submission #{video.id.slice(-4)}</p>
        <p className="text-sm text-gray-500">
          By {video.editorName} • {new Date(video.submittedAt).toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })}
        </p>
      </div>
      <Badge variant={videoStatusBadgeVariant(video.status)}>{statusLabel}</Badge>
    </div>
  );
}

export function BriefDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: brief, isLoading: briefLoading } = useQuery({
    queryKey: ['brief', id],
    queryFn: () => api.getBrief(id!),
    enabled: !!id,
  });

  const { data: videos, isLoading: videosLoading } = useQuery({
    queryKey: ['videos', 'brief', id],
    queryFn: () => api.getVideos(),
    enabled: !!id,
  });

  const briefVideos = videos?.filter((v) => v.briefId === id) ?? [];

  if (briefLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-24" />
        <Skeleton className="h-64 w-full" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  if (!brief) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Brief not found</p>
        <Button className="mt-4" onClick={() => navigate('/briefs')}>Back to Briefs</Button>
      </div>
    );
  }

  const statusVariant = statusBadgeVariant(brief.status);
  const statusLabel = statusLabels[brief.status];
  const deadlineDate = brief.deadline 
    ? new Date(brief.deadline).toLocaleDateString('en-US', { month: 'long', day: 'numeric', year: 'numeric' })
    : 'No deadline';

  return (
    <div className="space-y-6">
      <Button variant="ghost" onClick={() => navigate('/briefs')}>
        ← Back to Briefs
      </Button>

      <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-6 space-y-4">
        <div className="flex justify-between items-start">
          <h1 className="text-2xl font-bold text-gray-900">{brief.title}</h1>
          <Badge variant={statusVariant}>{statusLabel}</Badge>
        </div>

        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
          <div>
            <p className="text-gray-500">Bounty Budget</p>
            <p className="font-semibold text-gray-900">${brief.bountyBudget.toLocaleString()}</p>
          </div>
          <div>
            <p className="text-gray-500">Submissions</p>
            <p className="font-semibold text-gray-900">{brief.currentSubmissions}/{brief.submissionLimit}</p>
          </div>
          <div>
            <p className="text-gray-500">Client</p>
            <p className="font-medium text-gray-900">{brief.clientName}</p>
          </div>
          <div>
            <p className="text-gray-500">Deadline</p>
            <p className="font-medium text-gray-900">{deadlineDate}</p>
          </div>
        </div>

        <div className="flex flex-wrap gap-2">
          {brief.tags.map((tag) => (
            <Badge key={tag} variant="info">{tag}</Badge>
          ))}
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Description</CardTitle>
        </CardHeader>
        <CardBody>
          <p className="text-gray-700 whitespace-pre-wrap">{brief.description}</p>
        </CardBody>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Submissions ({briefVideos.length})</CardTitle>
        </CardHeader>
        <CardBody>
          {videosLoading ? (
            <div className="space-y-3">
              {[...Array(3)].map((_, i) => (
                <Skeleton key={i} className="h-16 w-full" />
              ))}
            </div>
          ) : briefVideos.length === 0 ? (
            <p className="text-gray-500 text-center py-8">No submissions yet</p>
          ) : (
            <div className="space-y-3">
              {briefVideos.map((video) => (
                <SubmissionItem key={video.id} video={video} />
              ))}
            </div>
          )}
        </CardBody>
      </Card>

      {brief.status === 'published' && (
        <Button className="w-full">Request More Submissions</Button>
      )}
    </div>
  );
}