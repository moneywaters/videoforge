"use client";

import { UploadItem, UploadFile } from './UploadItem';

export interface UploadListProps {
  uploads: UploadFile[];
  onCancel: (id: string) => void;
}

export function UploadList({ uploads, onCancel }: UploadListProps) {
  if (uploads.length === 0) {
    return null;
  }

  return (
    <div className="space-y-2">
      {uploads.map((upload) => (
        <UploadItem key={upload.id} upload={upload} onCancel={onCancel} />
      ))}
    </div>
  );
}