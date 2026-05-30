"use client";

import { useState, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { generateThumbnail } from './thumbnail-utils';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { UploadList } from '@/components/upload/UploadList';
import type { UploadFile } from '@/components/upload/UploadItem';
import { BriefFileExplorer, type BriefFile } from './file-explorer';
import { VideoPreviewModal } from './video-preview-modal';
import type { BriefStatus, Video } from '@/types/index';

const statusBadgeVariant = (status: BriefStatus): 'default' | 'secondary' | 'destructive' | 'outline' => {
  switch (status) {
    case 'published':
      return 'default';
    case 'closed':
      return 'destructive';
    case 'draft':
      return 'secondary';
    default:
      return 'outline';
  }
};

const statusLabels: Record<BriefStatus, string> = {
  draft: 'Draft',
  published: 'Open',
  closed: 'Closed',
};

const videoStatusBadgeVariant = (status: Video['status']): 'default' | 'secondary' | 'destructive' | 'outline' => {
  switch (status) {
    case 'approved':
      return 'default';
    case 'submitted':
      return 'outline';
    case 'rejected':
      return 'destructive';
    case 'needs_revision':
      return 'secondary';
    default:
      return 'outline';
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

interface UploadController {
  abortController: AbortController;
}

export default function BriefDetailPage() {
  const params = useParams();
  const id = params?.id as string;
  const router = useRouter();
  const queryClient = useQueryClient();
  const uploadControllersRef = useRef<Map<string, UploadController>>(new Map());

  const [uploads, setUploads] = useState<UploadFile[]>([]);
  const [briefFiles, setBriefFiles] = useState<BriefFile[]>([]);
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [previewTitle, setPreviewTitle] = useState<string>('');

  const isUploading = uploads.some((u) => u.status === 'uploading');

  const { data: brief, isLoading: briefLoading } = useQuery({
    queryKey: ['brief', id],
    queryFn: () => api.getBrief(id),
    enabled: !!id,
  });

  const { data: videos, isLoading: videosLoading } = useQuery({
    queryKey: ['videos', 'brief', id],
    queryFn: () => api.getVideos(),
    enabled: !!id,
  });

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length === 0) return;
    if (!id) return;

    const newUploads: UploadFile[] = files.map((file) => ({
      id: crypto.randomUUID(),
      file,
      progress: 0,
      status: 'pending',
    }));

    setUploads((prev) => [...prev, ...newUploads]);

    let anyCompleted = false;
    const errors: string[] = [];

    for (const uploadItem of newUploads) {
      const uploadId = uploadItem.id;
      const file = uploadItem.file;

      const abortController = new AbortController();
      uploadControllersRef.current.set(uploadId, { abortController });

      setUploads((prev) =>
        prev.map((u) =>
          u.id === uploadId ? { ...u, status: 'uploading' } : u
        )
      );

      try {
        const { url, uploadId: storjKey } = await api.getUploadUrl(id, file.name, file.type);

        if (!url) {
          throw new Error('Failed to get upload URL');
        }

        await api.uploadToPresignedUrl(url, file, {
          signal: abortController.signal,
          onProgress: (loaded, total) => {
            const progress = Math.round((loaded / total) * 100);
            setUploads((prev) =>
              prev.map((u) =>
                u.id === uploadId ? { ...u, progress } : u
              )
            );
          },
        });

await api.confirmUpload(id, storjKey, {
          name: file.name,
          type: file.type,
          size: file.size,
        });

        // Fetch the presigned download URL for preview
        let fileUrl: string | undefined;
        try {
          const resp = await api.getDownloadUrl(id);
          fileUrl = resp.download_url;
        } catch {
          // ignore, fileUrl stays undefined
        }

        anyCompleted = true;
        setUploads((prev) =>
          prev.map((u) =>
            u.id === uploadId ? { ...u, status: 'completed', progress: 100 } : u
          )
        );

        // Generate thumbnail for images and videos asynchronously
        const thumbnailUrl = await generateThumbnail(file);

        // Add completed upload to file explorer
        setBriefFiles((prev) => [
          ...prev,
          {
            id: crypto.randomUUID(),
            name: file.name,
            type: file.type,
            size: file.size,
            uploadedAt: new Date().toISOString(),
            thumbnailUrl: thumbnailUrl ?? undefined,
          },
        ]);
      } catch (error) {
        const isCancelled =
          error instanceof Error && error.message === 'Upload cancelled';
        setUploads((prev) =>
          prev.map((u) =>
            u.id === uploadId
              ? {
                  ...u,
                  status: isCancelled ? 'cancelled' : 'error',
                  error: isCancelled ? undefined : error instanceof Error ? error.message : 'Upload failed',
                }
              : u
          )
        );
        if (!isCancelled) {
          errors.push(error instanceof Error ? error.message : 'Unknown error');
        }
      } finally {
        uploadControllersRef.current.delete(uploadId);
      }
    }

    if (anyCompleted) {
      queryClient.invalidateQueries({ queryKey: ['videos', 'brief', id] });
      alert('Upload successful!');
    }

    if (errors.length > 0) {
      alert(`Some uploads failed:\n${errors.join('\n')}`);
    }
  };

  const handleDownload = async () => {
    try {
      const resp = await api.getDownloadUrl(id);
      if (resp.download_url) {
        window.open(resp.download_url, '_blank', 'noopener,noreferrer');
      }
    } catch (error: unknown) {
      alert(`Download failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    }
  };

  const handleCancel = (uploadId: string) => {
    const controller = uploadControllersRef.current.get(uploadId);
    if (controller) {
      controller.abortController.abort();
      setUploads((prev) =>
        prev.map((u) =>
          u.id === uploadId ? { ...u, status: 'cancelled' } : u
        )
      );
      setTimeout(() => {
        setUploads((prev) => prev.filter((u) => u.id !== uploadId));
      }, 500);
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
        <Button className="mt-4" onClick={() => router.push('/dashboard/briefs')}>Back to Briefs</Button>
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
      <Button variant="ghost" onClick={() => router.push('/dashboard/briefs')}>
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
            <Badge key={tag} variant="secondary">{tag}</Badge>
          ))}
        </div>

        <div className="flex gap-4 pt-4 border-t border-gray-100">
          <div>
            <input
              type="file"
              id="upload-footage"
              multiple
              className="hidden"
              onChange={handleFileChange}
              disabled={isUploading}
            />
            <label htmlFor="upload-footage" className="cursor-pointer inline-block">
              <div className={`px-4 py-2 rounded-md font-medium text-sm transition-colors ${isUploading ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-gray-100 text-gray-900 hover:bg-gray-200'}`}>
                {isUploading ? 'Uploading…' : 'Upload Raw Footage'}
              </div>
            </label>
          </div>
          <Button variant="secondary" onClick={handleDownload}>
            Download Raw Footage
          </Button>
        </div>

        {uploads.length > 0 && (
          <div className="mt-4">
            <UploadList uploads={uploads} onCancel={handleCancel} />
          </div>
        )}

        <BriefFileExplorer
          files={briefFiles}
          onDownload={(file) => file.url && window.open(file.url, '_blank', 'noopener,noreferrer')}
          onPreview={(file) => {
            setPreviewTitle(file.name);
            setPreviewUrl(file.url ?? null);
          }}
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Description</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="text-gray-700 whitespace-pre-wrap">{brief.description}</p>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Submissions ({briefVideos.length})</CardTitle>
        </CardHeader>
        <CardContent>
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
        </CardContent>
      </Card>

      {brief.status === 'published' && (
        <Button className="w-full">Request More Submissions</Button>
      )}

      <VideoPreviewModal
        url={previewUrl}
        title={previewTitle}
        onClose={() => setPreviewUrl(null)}
      />
    </div>
  );
}