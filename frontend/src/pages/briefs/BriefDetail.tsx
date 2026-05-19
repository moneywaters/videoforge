import { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
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
  const queryClient = useQueryClient();

  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);

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

  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      setUploading(true);
      setUploadProgress(0);
      try {
        // 1. Get presigned URL
        const { url, uploadId } = await api.getUploadUrl(id!, file.name, file.type);
        
        // 2. Upload directly to Storj via presigned URL with progress
        await new Promise<void>((resolve, reject) => {
          const xhr = new XMLHttpRequest();
          xhr.open('PUT', url);
          xhr.setRequestHeader('Content-Type', file.type);
          
          xhr.upload.onprogress = (e) => {
            if (e.lengthComputable) {
              setUploadProgress(Math.round((e.loaded / e.total) * 100));
            }
          };
          
          xhr.onload = () => {
            if (xhr.status >= 200 && xhr.status < 300) {
              resolve();
            } else {
              reject(new Error(`Upload failed with status ${xhr.status}`));
            }
          };
          
          xhr.onerror = () => reject(new Error('Upload failed'));
          xhr.send(file);
        });

        // 3. Confirm upload
        await api.confirmUpload(id!, uploadId);
      } finally {
        setUploading(false);
        setUploadProgress(0);
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['videos', 'brief', id] });
      alert('Upload successful!');
    },
    onError: (error: Error) => {
      alert(`Upload failed: ${error.message}`);
    }
  });

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      uploadMutation.mutate(file);
    }
    // reset input
    e.target.value = '';
  };

  const handleDownload = async () => {
    try {
      const { url } = await api.getDownloadUrl(id!);
      window.open(url, '_blank', 'noopener,noreferrer');
    } catch (error: unknown) {
      alert(`Download failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

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
        
        <div className="flex gap-4 pt-4 border-t border-gray-100">
           <div>
             <input
               type="file"
               id="upload-footage"
               className="hidden"
               onChange={handleFileChange}
               disabled={uploading}
             />
             <label htmlFor="upload-footage" className="cursor-pointer inline-block">
               <div className={`px-4 py-2 rounded-md font-medium text-sm transition-colors ${uploading ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-gray-100 text-gray-900 hover:bg-gray-200'}`}>
                 {uploading ? `Uploading... ${uploadProgress}%` : 'Upload Raw Footage'}
               </div>
             </label>
           </div>
           <Button variant="secondary" onClick={handleDownload}>
             Download Raw Footage
           </Button>
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