"use client";

import { IconFile, IconVideo, IconX } from '@tabler/icons-react';

export type UploadStatus = 'pending' | 'uploading' | 'completed' | 'error' | 'cancelled';

export interface UploadFile {
  id: string;
  file: globalThis.File;
  progress: number;
  status: UploadStatus;
  error?: string;
}

export interface UploadItemProps {
  upload: UploadFile;
  onCancel: (id: string) => void;
}

const statusLabels: Record<UploadStatus, string> = {
  pending: 'Pending',
  uploading: 'Uploading…',
  completed: 'Complete',
  error: 'Error',
  cancelled: 'Cancelled',
};

const statusColors: Record<UploadStatus, string> = {
  pending: 'text-gray-500',
  uploading: 'text-blue-600',
  completed: 'text-green-600',
  error: 'text-red-600',
  cancelled: 'text-gray-400',
};

function getFileIcon(file: globalThis.File) {
  const isVideo = file.type?.startsWith('video/') ?? false;
  return isVideo ? IconVideo : IconFile;
}

export function UploadItem({ upload, onCancel }: UploadItemProps) {
  const { id, file, progress, status, error } = upload;
  const FileDisplayIcon = getFileIcon(file);
  // Show X button for pending (waiting), uploading, error, or cancelled - but NOT completed
  const showCancel = status !== 'completed';

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-3 space-y-2">
      <div className="flex items-center gap-2">
        <FileDisplayIcon className="w-5 h-5 flex-shrink-0 text-gray-400" />
        <span
          className="text-sm text-gray-700 truncate flex-1 min-w-0"
          title={file.name}
        >
          {file.name}
        </span>
        {showCancel && (
          <button
            type="button"
            onClick={() => onCancel(id)}
            aria-label={status === 'pending' ? 'Remove from queue' : 'Cancel upload'}
            className="p-1 rounded hover:bg-red-100 transition-colors flex-shrink-0"
          >
            <IconX className="w-3.5 h-3.5 text-gray-500 hover:text-red-600" />
          </button>
        )}
      </div>

      {/* Show progress bar only for uploading/completed, show waiting state for pending */}
      {status === 'pending' ? (
        <div className="h-6 bg-gray-100 rounded-full flex items-center justify-center">
          <span className="text-xs text-gray-500">Waiting...</span>
        </div>
      ) : (
        <div className="relative">
          <div
            role="progressbar"
            aria-valuenow={progress}
            aria-valuemin={0}
            aria-valuemax={100}
            aria-label={`Upload progress: ${progress}%`}
            className="h-6 bg-gray-200 rounded-full overflow-hidden"
          >
            <div
              className="h-full bg-blue-500 rounded-full transition-all duration-200"
              style={{ width: `${progress}%` }}
            />
            <div className="absolute inset-0 flex items-center justify-center">
              <span className="text-xs font-medium text-white drop-shadow-sm z-10">
                {progress}%
              </span>
            </div>
          </div>
        </div>
      )}

      <div className="flex items-center justify-between">
        <span className={`text-xs ${statusColors[status]}`}>
          {status === 'error' && error ? error : statusLabels[status]}
        </span>
      </div>
    </div>
  );
}