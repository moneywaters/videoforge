"use client";

import { useState, useMemo } from 'react';
import {
  IconFolder,
  IconPhoto,
  IconVideo,
  IconFile,
  IconFileMusic,
  IconFileZip,
  IconFileTypePdf,
  IconFileCode,
  IconLayoutGrid,
  IconList,
  IconArrowsSort,
  IconCalendar,
  IconRuler,
} from '@tabler/icons-react';
import { Button } from '@/components/ui/button';
import { Slider } from '@/components/ui/slider';
import { cn } from '@/lib/utils';

export interface BriefFile {
  id: string;
  name: string;
  type: string;
  size: number;
  uploadedAt: string;
  url?: string;
}

interface BriefFileExplorerProps {
  files: BriefFile[];
  onDownload?: (file: BriefFile) => void;
  onPreview?: (file: BriefFile) => void;
}

type SortKey = 'name' | 'size' | 'date';
type SortDir = 'asc' | 'desc';
type ViewMode = 'grid' | 'list';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  return `${(bytes / 1024 ** i).toFixed(i === 0 ? 0 : 2)} ${units[i]}`;
}

function getFileIcon(type: string) {
  if (type.startsWith('image/')) return IconPhoto;
  if (type.startsWith('video/')) return IconVideo;
  if (type.startsWith('audio/')) return IconFileMusic;
  if (type === 'application/pdf') return IconFileTypePdf;
  if (type.includes('zip') || type.includes('compressed')) return IconFileZip;
  if (type.includes('json') || type.includes('xml') || type.includes('javascript') || type.includes('typescript')) return IconFileCode;
  if (type.includes('folder')) return IconFolder;
  return IconFile;
}

export function BriefFileExplorer({ files, onDownload, onPreview }: BriefFileExplorerProps) {
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [iconSize, setIconSize] = useState(48);
  const [sortKey, setSortKey] = useState<SortKey>('date');
  const [sortDir, setSortDir] = useState<SortDir>('desc');
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [infoWidth, setInfoWidth] = useState(280);
  const [isResizing, setIsResizing] = useState(false);

  const sorted = useMemo(() => {
    const list = [...files];
    list.sort((a, b) => {
      let cmp = 0;
      if (sortKey === 'name') cmp = a.name.localeCompare(b.name);
      else if (sortKey === 'size') cmp = a.size - b.size;
      else cmp = new Date(a.uploadedAt).getTime() - new Date(b.uploadedAt).getTime();
      return sortDir === 'asc' ? cmp : -cmp;
    });
    return list;
  }, [files, sortKey, sortDir]);

  const selected = useMemo(() => files.find((f) => f.id === selectedId) ?? null, [files, selectedId]);

  const toggleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDir((d) => (d === 'asc' ? 'desc' : 'asc'));
    } else {
      setSortKey(key);
      setSortDir('asc');
    }
  };

  const handleResizeStart = () => setIsResizing(true);

  return (
    <div className="space-y-3">
      {/* Toolbar */}
      <div className="flex flex-wrap items-center gap-3 bg-muted/50 rounded-lg p-2">
        {/* View Toggle */}
        <div className="flex border rounded-md overflow-hidden">
          <button
            onClick={() => setViewMode('grid')}
            className={cn(
              'px-3 py-1.5 text-sm flex items-center gap-1.5 transition-colors',
              viewMode === 'grid' ? 'bg-background shadow-sm' : 'text-muted-foreground hover:bg-background/50'
            )}
          >
            <IconLayoutGrid className="w-4 h-4" /> Grid
          </button>
          <button
            onClick={() => setViewMode('list')}
            className={cn(
              'px-3 py-1.5 text-sm flex items-center gap-1.5 transition-colors',
              viewMode === 'list' ? 'bg-background shadow-sm' : 'text-muted-foreground hover:bg-background/50'
            )}
          >
            <IconList className="w-4 h-4" /> List
          </button>
        </div>

        {/* Sort Selectors */}
        <div className="flex items-center gap-1">
          {(['name', 'size', 'date'] as SortKey[]).map((key) => (
            <button
              key={key}
              onClick={() => toggleSort(key)}
              className={cn(
                'px-2.5 py-1.5 text-xs rounded-md flex items-center gap-1 transition-colors',
                sortKey === key
                  ? 'bg-primary text-primary-foreground'
                  : 'text-muted-foreground hover:bg-muted'
              )}
            >
              {key === 'name' && <IconArrowsSort className="w-3.5 h-3.5" />}
              {key === 'size' && <IconRuler className="w-3.5 h-3.5" />}
              {key === 'date' && <IconCalendar className="w-3.5 h-3.5" />}
              {key}
              {sortKey === key && ` ${sortDir === 'asc' ? '↑' : '↓'}`}
            </button>
          ))}
        </div>

        {/* Icon Size Slider */}
        {viewMode === 'grid' && (
          <div className="flex items-center gap-2 min-w-[140px]">
            <span className="text-xs text-muted-foreground whitespace-nowrap">Size</span>
            <Slider
              value={[iconSize]}
              min={24}
              max={96}
              step={8}
              onValueChange={([v]) => setIconSize(v)}
              className="w-24"
            />
          </div>
        )}
      </div>

      {/* Files + Info Split */}
      <div className="flex gap-0 border rounded-lg overflow-hidden bg-background min-h-[200px]">
        {/* File List */}
        <div className="flex-1 min-w-0 p-3">
          {sorted.length === 0 ? (
            <div className="flex flex-col items-center justify-center h-full text-muted-foreground py-12">
              <IconFolder className="w-12 h-12 opacity-20 mb-3" />
              <p className="text-sm">No files yet</p>
              <p className="text-xs opacity-60">Upload files to see them here</p>
            </div>
          ) : viewMode === 'grid' ? (
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-3">
              {sorted.map((file) => {
                const Icon = getFileIcon(file.type);
                const isActive = selectedId === file.id;
                return (
                  <button
                    key={file.id}
                    onClick={() => setSelectedId(isActive ? null : file.id)}
                    onDoubleClick={() => file.type.startsWith('video/') && onPreview?.(file)}
                    className={cn(
                      'flex flex-col items-center gap-2 p-3 rounded-lg border transition-all text-center',
                      isActive
                        ? 'border-primary bg-primary/5 shadow-sm'
                        : 'border-border hover:border-foreground/20 hover:bg-muted/50'
                    )}
                  >
                    <Icon className="text-muted-foreground" style={{ width: iconSize, height: iconSize }} />
                    <span className="text-xs truncate w-full" title={file.name}>
                      {file.name}
                    </span>
                  </button>
                );
              })}
            </div>
          ) : (
            <div className="space-y-1">
              {sorted.map((file) => {
                const Icon = getFileIcon(file.type);
                const isActive = selectedId === file.id;
                return (
                  <button
                    key={file.id}
                    onClick={() => setSelectedId(isActive ? null : file.id)}
                    onDoubleClick={() => file.type.startsWith('video/') && onPreview?.(file)}
                    className={cn(
                      'w-full flex items-center gap-3 px-3 py-2 rounded-md text-left transition-colors',
                      isActive ? 'bg-muted' : 'hover:bg-muted/50'
                    )}
                  >
                    <Icon className="w-5 h-5 text-muted-foreground flex-shrink-0" />
                    <span className="text-sm truncate flex-1 min-w-0" title={file.name}>
                      {file.name}
                    </span>
                    <span className="text-xs text-muted-foreground tabular-nums w-16 text-right">
                      {formatBytes(file.size)}
                    </span>
                    <span className="text-xs text-muted-foreground w-24 text-right hidden sm:block">
                      {new Date(file.uploadedAt).toLocaleDateString()}
                    </span>
                  </button>
                );
              })}
            </div>
          )}
        </div>

        {/* Resizable Info Panel */}
        {selected && (
          <>
            {/* Drag Handle */}
            <div
              className="w-0.5 bg-border cursor-col-resize hover:bg-primary/50 transition-colors relative group"
              onMouseDown={(e) => {
                const startX = e.clientX;
                const startW = infoWidth;
                const handleMouseMove = (ev: MouseEvent) => {
                  const newW = Math.max(200, Math.min(500, startW + (startX - ev.clientX)));
                  setInfoWidth(newW);
                };
                const handleMouseUp = () => {
                  setIsResizing(false);
                  document.removeEventListener('mousemove', handleMouseMove);
                  document.removeEventListener('mouseup', handleMouseUp);
                };
                setIsResizing(true);
                document.addEventListener('mousemove', handleMouseMove);
                document.addEventListener('mouseup', handleMouseUp);
              }}
            >
              <div className="absolute inset-y-0 -left-1 w-2 cursor-col-resize" />
            </div>

            {/* Info Content */}
            <div
              className="p-4 bg-muted/30 border-l space-y-4 select-none"
              style={{ width: infoWidth, minWidth: infoWidth }}
            >
              <div className="flex items-center gap-2 mb-2">
                {(() => {
                  const Icon = getFileIcon(selected.type);
                  return <Icon className="w-8 h-8 text-muted-foreground" />;
                })()}
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate" title={selected.name}>
                    {selected.name}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {selected.type || 'unknown type'}
                  </p>
                </div>
              </div>

              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Size</span>
                  <span>{formatBytes(selected.size)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Uploaded</span>
                  <span>{new Date(selected.uploadedAt).toLocaleString()}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">ID</span>
                  <span className="font-mono text-xs">{selected.id.slice(0, 8)}</span>
                </div>
              </div>

              {selected.type.startsWith('video/') && onPreview && (
                <Button
                  size="sm"
                  variant="secondary"
                  className="w-full mt-2"
                  onClick={() => onPreview(selected)}
                >
                  <IconVideo className="w-4 h-4 mr-1.5" /> Play Video
                </Button>
              )}

              {onDownload && selected.url && (
                <Button
                  size="sm"
                  className="w-full mt-2"
                  onClick={() => onDownload(selected)}
                >
                  Download
                </Button>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
