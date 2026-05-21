import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Card, CardBody, CardHeader, CardTitle } from '@/components/ui/Card';
import { Skeleton } from '@/components/ui/Skeleton';
import { Textarea } from '@/components/ui/Textarea';
import type { VideoStatus } from '@/types/index';

const videoStatusBadgeVariant = (status: VideoStatus) => {
  switch (status) {
    case 'approved': return 'success';
    case 'submitted': return 'info';
    case 'rejected': return 'danger';
    case 'needs_revision': return 'warning';
    default: return 'default';
  }
};

const statusLabels: Record<VideoStatus, string> = {
  submitted: 'Submitted',
  approved: 'Approved',
  rejected: 'Rejected',
  needs_revision: 'Revision Requested',
};

export function VideoDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [feedback, setFeedback] = useState('');
  const [actionType, setActionType] = useState<'approve' | 'reject' | 'revision' | null>(null);

  const { data: video, isLoading } = useQuery({
    queryKey: ['video', id],
    queryFn: () => api.getVideo(id!),
    enabled: !!id,
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ status, feedbackText }: { status: VideoStatus; feedbackText?: string }) =>
      api.updateVideoStatus(id!, status, feedbackText),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['video', id] });
      queryClient.invalidateQueries({ queryKey: ['videos'] });
      setActionType(null);
      setFeedback('');
    },
  });

  const handleAction = (action: 'approve' | 'reject' | 'revision') => {
    const statusMap: Record<string, VideoStatus> = {
      approve: 'approved',
      reject: 'rejected',
      revision: 'needs_revision',
    };
    updateStatusMutation.mutate({
      status: statusMap[action],
      feedbackText: action !== 'approve' ? feedback : undefined,
    });
  };

  if (isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-24" />
        <Skeleton className="aspect-video w-full" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  }

  if (!video) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500">Video not found</p>
        <Button className="mt-4" onClick={() => navigate('/videos')}>Back to Videos</Button>
      </div>
    );
  }

  const statusVariant = videoStatusBadgeVariant(video.status);
  const statusLabel = statusLabels[video.status];
  const submittedDate = new Date(video.submittedAt).toLocaleDateString('en-US', {
    month: 'long',
    day: 'numeric',
    year: 'numeric',
  });

  // This would come from auth in a real app
  const userRole = 'client' as 'client' | 'editor';

  return (
    <div className="space-y-6">
      <Button variant="ghost" onClick={() => navigate('/videos')}>
        ← Back to Videos
      </Button>

      <div className="aspect-video bg-gray-900 rounded-lg flex items-center justify-center">
        <svg className="w-16 h-16 text-white opacity-80" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z" />
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
      </div>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-start">
            <CardTitle>{video.briefTitle}</CardTitle>
            <Badge variant={statusVariant}>{statusLabel}</Badge>
          </div>
        </CardHeader>
        <CardBody className="space-y-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <p className="text-gray-500">Duration</p>
              <p className="font-medium text-gray-900">{video.duration}s</p>
            </div>
            <div>
              <p className="text-gray-500">Resolution</p>
              <p className="font-medium text-gray-900">{video.resolution}</p>
            </div>
            <div>
              <p className="text-gray-500">Submitted By</p>
              <p className="font-medium text-gray-900">{video.editorName}</p>
            </div>
            <div>
              <p className="text-gray-500">Date</p>
              <p className="font-medium text-gray-900">{submittedDate}</p>
            </div>
          </div>

          {video.feedback && userRole === 'client' && (
            <div className="border-t pt-4">
              <p className="text-sm font-medium text-gray-700 mb-1">Client Feedback</p>
              <p className="text-sm text-gray-600 bg-amber-50 p-3 rounded border border-amber-200">
                {video.feedback}
              </p>
            </div>
          )}

          {video.revisions.length > 0 && (
            <div className="border-t pt-4">
              <p className="text-sm font-medium text-gray-700 mb-2">Revision History</p>
              <div className="space-y-2">
                {video.revisions.map((rev) => (
                  <div key={rev.id} className="flex justify-between text-sm p-2 bg-gray-50 rounded">
                    <span>Version {rev.version}</span>
                    <span className="text-gray-500">
                      {new Date(rev.createdAt).toLocaleDateString()}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardBody>
      </Card>

      {userRole === 'client' && (
        <Card>
          <CardHeader>
            <CardTitle>Actions</CardTitle>
          </CardHeader>
          <CardBody className="space-y-4">
            {actionType ? (
              <div className="space-y-3">
                <Textarea
                  label={actionType === 'approve' ? 'Notes (optional)' : 'Feedback (required)'}
                  value={feedback}
                  onChange={(e) => setFeedback(e.target.value)}
                  placeholder={
                    actionType === 'reject'
                      ? 'Explain why the video was rejected...'
                      : 'Provide feedback for revision...'
                  }
                />
                <div className="flex gap-2">
                  <Button
                    onClick={() => handleAction(actionType)}
                    disabled={
                      updateStatusMutation.isPending ||
                      (actionType !== 'approve' && !feedback.trim())
                    }
                  >
                    {updateStatusMutation.isPending ? 'Processing...' : 'Confirm'}
                  </Button>
                  <Button variant="secondary" onClick={() => setActionType(null)}>
                    Cancel
                  </Button>
                </div>
              </div>
            ) : (
              <div className="flex flex-wrap gap-2">
                <Button onClick={() => setActionType('approve')}>Approve</Button>
                <Button variant="danger" onClick={() => setActionType('reject')}>
                  Reject
                </Button>
                <Button variant="secondary" onClick={() => setActionType('revision')}>
                  Request Revision
                </Button>
              </div>
            )}
          </CardBody>
        </Card>
      )}

      {userRole === 'editor' && video.status === 'needs_revision' && (
        <Card>
          <CardHeader>
            <CardTitle>Feedback</CardTitle>
          </CardHeader>
          <CardBody className="space-y-4">
            {video.feedback && (
              <div className="bg-amber-50 p-4 rounded border border-amber-200">
                <p className="text-sm font-medium text-amber-800">Client Feedback:</p>
                <p className="text-sm text-amber-700 mt-1">{video.feedback}</p>
              </div>
            )}
            <Button>Submit New Revision</Button>
          </CardBody>
        </Card>
      )}
    </div>
  );
}