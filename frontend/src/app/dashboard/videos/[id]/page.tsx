'use client';

import { useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { Textarea } from '@/components/ui/textarea';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';
import type { VideoStatus } from '@/types';

const videoStatusBadgeVariant = (status: VideoStatus) => {
  switch (status) {
    case 'approved': return 'default';
    case 'submitted': return 'secondary';
    case 'rejected': return 'destructive';
    case 'needs_revision': return 'secondary';
    default: return 'default';
  }
};

const statusLabels: Record<VideoStatus, string> = {
  submitted: 'Submitted',
  approved: 'Approved',
  rejected: 'Rejected',
  needs_revision: 'Revision Requested',
};

export default function VideoDetailPage() {
  const params = useParams();
  const router = useRouter();
  const queryClient = useQueryClient();
  const id = params.id as string;
  const [feedback, setFeedback] = useState('');
  const [actionType, setActionType] = useState<'approve' | 'reject' | 'revision' | null>(null);

  const { data: video, isLoading } = useQuery({
    queryKey: ['video', id],
    queryFn: () => api.getVideo(id),
    enabled: !!id,
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ status, feedbackText }: { status: VideoStatus; feedbackText?: string }) =>
      api.updateVideoStatus(id, status, feedbackText),
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
        <p className="text-muted-foreground">Video not found</p>
        <Button className="mt-4" onClick={() => router.push('/dashboard/videos')}>
          Back to Videos
        </Button>
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
      <Button variant="ghost" onClick={() => router.push('/dashboard/videos')}>
        <Icons.chevronLeft className="mr-2 h-4 w-4" />
        Back to Videos
      </Button>

      <div className="aspect-video bg-muted rounded-lg flex items-center justify-center">
        <Icons.video className="h-16 w-16 text-muted-foreground opacity-80" />
      </div>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-start">
            <CardTitle>{video.briefTitle}</CardTitle>
            <Badge variant={statusVariant}>{statusLabel}</Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <p className="text-muted-foreground">Duration</p>
              <p className="font-medium">{video.duration}s</p>
            </div>
            <div>
              <p className="text-muted-foreground">Resolution</p>
              <p className="font-medium">{video.resolution}</p>
            </div>
            <div>
              <p className="text-muted-foreground">Submitted By</p>
              <p className="font-medium">{video.editorName}</p>
            </div>
            <div>
              <p className="text-muted-foreground">Date</p>
              <p className="font-medium">{submittedDate}</p>
            </div>
          </div>

          {video.feedback && userRole === 'client' && (
            <div className="border-t pt-4">
              <p className="text-sm font-medium mb-1">Client Feedback</p>
              <p className="text-sm text-muted-foreground bg-amber-50 dark:bg-amber-950 p-3 rounded border border-amber-200 dark:border-amber-800">
                {video.feedback}
              </p>
            </div>
          )}

          {video.revisions.length > 0 && (
            <div className="border-t pt-4">
              <p className="text-sm font-medium mb-2">Revision History</p>
              <div className="space-y-2">
                {video.revisions.map((rev) => (
                  <div key={rev.id} className="flex justify-between text-sm p-2 bg-muted rounded">
                    <span>Version {rev.version}</span>
                    <span className="text-muted-foreground">
                      {new Date(rev.createdAt).toLocaleDateString()}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </CardContent>
      </Card>

      {userRole === 'client' && (
        <Card>
          <CardHeader>
            <CardTitle>Actions</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {actionType ? (
              <div className="space-y-3">
                <label className="text-sm font-medium">
                  {actionType === 'approve' ? 'Notes (optional)' : 'Feedback (required)'}
                </label>
                <Textarea
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
                <Button variant="destructive" onClick={() => setActionType('reject')}>
                  Reject
                </Button>
                <Button variant="secondary" onClick={() => setActionType('revision')}>
                  Request Revision
                </Button>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {userRole === 'editor' && video.status === 'needs_revision' && (
        <Card>
          <CardHeader>
            <CardTitle>Feedback</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {video.feedback && (
              <div className="bg-amber-50 dark:bg-amber-950 p-4 rounded border border-amber-200 dark:border-amber-800">
                <p className="text-sm font-medium text-amber-800 dark:text-amber-200">Client Feedback:</p>
                <p className="text-sm text-amber-700 dark:text-amber-300 mt-1">{video.feedback}</p>
              </div>
            )}
            <Button>Submit New Revision</Button>
          </CardContent>
        </Card>
      )}
    </div>
  );
}