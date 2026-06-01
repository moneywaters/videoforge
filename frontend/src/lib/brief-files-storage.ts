function getStorage(): Storage | null {
  if (typeof window === 'undefined') return null;
  return localStorage;
}

export interface StoredBriefFile {
  id: string;
  name: string;
  type: string;
  size: number;
  uploadedAt: string;
  url?: string;
  thumbnailUrl?: string;
}

export function loadBriefFiles(briefId: string): StoredBriefFile[] {
  const raw = getStorage()?.getItem(`brief-files-${briefId}`);
  return raw ? JSON.parse(raw) : [];
}

export function saveBriefFiles(briefId: string, files: StoredBriefFile[]) {
  getStorage()?.setItem(`brief-files-${briefId}`, JSON.stringify(files));
}