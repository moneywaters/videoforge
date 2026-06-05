"use client";

import { useState, useMemo, useEffect } from 'react';
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
  IconDownload,
  IconFileText,
  IconAlertCircle,
} from '@tabler/icons-react';
import streamSaver from 'streamsaver';
// Configure stream-saver to use local Service Worker files from public/
streamSaver.mitm = '/streamsaver-mitm.html';
if ('serviceWorker' in navigator && 'WritableStream' in window) {
  streamSaver.url = '/streamsaver-sw.js';
}
import { Zip, ZipDeflate, ZipPassThrough } from 'fflate';
import { Button } from '@/components/ui/button';
import { Slider } from '@/components/ui/slider';
import { cn } from '@/lib/utils';

// Text file extensions that can be previewed inline
const TEXT_EXTENSIONS = new Set([
  'txt', 'md', 'json', 'yaml', 'yml', 'csv', 'tsv', 'log', 'xml', 'html',
  'css', 'js', 'ts', 'jsx', 'tsx', 'py', 'sh', 'bash', 'env', 'gitignore', 'config',
]);

function isTextFile(filename: string): boolean {
  const ext = filename.split('.').pop()?.toLowerCase() ?? '';
  return TEXT_EXTENSIONS.has(ext);
}

export interface BriefFile {
  id: string;
  name: string;
  type: string;
  size: number;
  uploadedAt: string;
  url?: string;
  thumbnailUrl?: string;
}

interface BriefFileExplorerProps {
  files: BriefFile[];
  briefId?: string;
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

export function BriefFileExplorer({ files, briefId, onDownload, onPreview }: BriefFileExplorerProps) {
  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [iconSize, setIconSize] = useState(48);
  const [sortKey, setSortKey] = useState<SortKey>('date');
  const [sortDir, setSortDir] = useState<SortDir>('desc');
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [infoWidth, setInfoWidth] = useState(280);
  const [isResizing, setIsResizing] = useState(false);
  const [isZipping, setIsZipping] = useState(false);
  const [zipProgress, setZipProgress] = useState<{
    current: number;
    total: number;
    bytesWritten: number;
    currentFile: string;
  } | null>(null);
  const [textContent, setTextContent] = useState<string | null>(null);
  const [textLoading, setTextLoading] = useState(false);
  const [textError, setTextError] = useState<string | null>(null);
  const [imageError, setImageError] = useState(false);
  const [videoError, setVideoError] = useState(false);

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

  const handleDownloadAll = async () => {
    if (files.length === 0) return;

    setIsZipping(true);
    setZipProgress({ current: 0, total: files.length, bytesWritten: 0, currentFile: 'Preparing...' });

    // Create streaming download — this immediately prompts the browser download
    const fileName = briefId ? `${briefId}-files.zip` : 'brief-files.zip';
    const fileStream = streamSaver.createWriteStream(fileName, { size: 0 });
    const writer = fileStream.getWriter();

    // Streaming ZIP — pumps output chunks to the writer
    const zip = new Zip((err, data, final) => {
      if (err) {
        console.error('ZIP compression error:', err);
        writer.abort(err);
        return;
      }
      writer.write(new Uint8Array(data)).catch(() => {});
      if (final) writer.close();
    });

    const failed: string[] = [];
    let added = 0;
    let totalBytes = 0;

    try {
      for (const file of files) {
        setZipProgress((p) => p && { ...p, currentFile: file.name });

        if (!file.url) {
          failed.push(file.name);
          setZipProgress((p) => p && { ...p, current: p.current + 1 });
          continue;
        }

        try {
          const resp = await fetch(file.url);
          if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
          const blob = await resp.blob();
          if (blob.size === 0) throw new Error('empty blob');

          // Choose PassThrough for already-compressed media, Deflate for text
          const ext = file.name.split('.').pop()?.toLowerCase() ?? '';
          const compressedExts = [
            'mp4', 'mov', 'webm', 'mkv', 'avi', 'mp3', 'wav',
            'jpg', 'jpeg', 'png', 'gif', 'webp', 'avif',
            'zip', 'rar', '7z', 'gz', 'bz2',
          ];
          const shouldCompress = !compressedExts.includes(ext);

          const entry = shouldCompress
            ? new ZipDeflate(file.name, { level: 6 })
            : new ZipPassThrough(file.name);

          zip.add(entry);

          // Stream the blob through the entry in chunks (64KB)
          const CHUNK = 65536;
          let offset = 0;
          while (offset < blob.size) {
            const slice = blob.slice(offset, offset + CHUNK);
            const buf = await slice.arrayBuffer();
            offset += CHUNK;
            const isFinal = offset >= blob.size;
            entry.push(new Uint8Array(buf), isFinal);
            totalBytes += buf.byteLength;
            setZipProgress({
              current: added + failed.length + 1,
              total: files.length,
              bytesWritten: totalBytes,
              currentFile: file.name,
            });
          }

          added++;
        } catch (err) {
          console.warn(`Failed to stream ${file.name}:`, err);
          failed.push(file.name);
        } finally {
          setZipProgress((p) =>
            p
              ? {
                  ...p,
                  current: added + failed.length,
                  total: files.length,
                }
              : p
          );
        }
      }

      zip.end();
    } catch (err) {
      console.error('ZIP stream failed:', err);
      try {
        writer.abort(err);
      } catch {
        // ignore
      }
      alert('Download failed: ' + (err instanceof Error ? err.message : 'unknown error'));
    } finally {
      setZipProgress(null);
      setIsZipping(false);
      if (failed.length > 0) {
        alert(`Downloaded ${added} of ${files.length} files. Skipped: ${failed.join(', ')}`);
      }
    }
  };

  // Load text content when a text file is selected
  useEffect(() => {
    setTextContent(null);
    setTextError(null);
    if (!selected || !selected.url) return;
    if (!isTextFile(selected.name)) return;
    if (selected.size > 1024 * 1024) {
      setTextError('File too large to preview inline, please download');
      return;
    }
    setTextLoading(true);
    fetch(selected.url)
      .then((r) => r.text())
      .then((text) => {
        // Only show first 50KB
        setTextContent(text.slice(0, 50 * 1024));
        setTextError(text.length > 50 * 1024 ? 'File truncated at 50KB' : null);
      })
      .catch(() => setTextError('Failed to load file content'))
      .finally(() => setTextLoading(false));
  }, [selected]);

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

        {/* Download All Button */}
        <Button
          size="sm"
          variant="outline"
          onClick={handleDownloadAll}
          disabled={files.length === 0 || isZipping}
          className="ml-auto"
        >
          {isZipping && zipProgress ? (
            <div className="flex items-center gap-2">
              <span className="text-xs">
                {zipProgress.current}/{zipProgress.total}
              </span>
            </div>
          ) : (
            <>
              <IconDownload className="w-4 h-4 mr-1.5" />
              Download All
            </>
          )}
        </Button>
        {isZipping && zipProgress && (
          <div className="absolute top-full right-0 mt-1 bg-background border rounded-md shadow-lg p-2 text-xs space-y-1 z-50 min-w-[160px]">
            <div className="flex justify-between gap-2">
              <span className="text-muted-foreground">Files:</span>
              <span>
                {zipProgress.current} / {zipProgress.total}
              </span>
            </div>
            <div className="flex justify-between gap-2">
              <span className="text-muted-foreground">Size:</span>
              <span>{(zipProgress.bytesWritten / 1024 / 1024).toFixed(1)} MB</span>
            </div>
            <div className="truncate max-w-[140px]">
              <span className="text-muted-foreground">Current:</span> {zipProgress.currentFile}
            </div>
          </div>
        )}
      </div>

      {/* Files + Info Split */}
      <div className="relative border rounded-lg overflow-hidden bg-background min-h-[200px]">
        {/* File List — full width, info pane overlays on top */}
        <div className="p-3">
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
                const isMedia = file.type.startsWith('image/') || file.type.startsWith('video/');
                return (
                  <div
                    role="button"
                    tabIndex={0}
                    key={file.id}
                    onClick={() => setSelectedId(isActive ? null : file.id)}
                    onDoubleClick={() => (file.type.startsWith('video/') || file.type.startsWith('image/')) && onPreview?.(file)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        setSelectedId(isActive ? null : file.id);
                      }
                    }}
                    className={cn(
                      'flex flex-col items-center gap-2 p-3 rounded-lg border transition-all text-center cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2',
                      isActive
                        ? 'border-primary bg-primary/5 shadow-sm'
                        : 'border-border hover:border-foreground/20 hover:bg-muted/50'
                    )}
                  >
                    {isMedia && file.thumbnailUrl ? (
                      <img
                        src={file.thumbnailUrl}
                        alt={file.name}
                        className="object-cover rounded-md"
                        style={{ width: iconSize, height: iconSize }}
                      />
                    ) : (
                      <Icon className="text-muted-foreground" style={{ width: iconSize, height: iconSize }} />
                    )}
                    <span className="text-xs truncate w-full" title={file.name}>
                      {file.name}
                    </span>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="space-y-1">
              {sorted.map((file) => {
                const Icon = getFileIcon(file.type);
                const isActive = selectedId === file.id;
                return (
                  <div
                    role="button"
                    tabIndex={0}
                    key={file.id}
                    onClick={() => setSelectedId(isActive ? null : file.id)}
                    onDoubleClick={() => (file.type.startsWith('video/') || file.type.startsWith('image/')) && onPreview?.(file)}
                    onKeyDown={(e) => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        setSelectedId(isActive ? null : file.id);
                      }
                    }}
                    className={cn(
                      'w-full flex items-center gap-3 px-3 py-2 rounded-md text-left transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2',
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
                  </div>
                );
              })}
            </div>
          )}
        </div>

        {/* Info Panel — absolute overlay, does NOT push file list */}
        {selected && (
          <>
            {/* Click-away backdrop */}
            <div
              className="absolute inset-0 bg-transparent"
              onClick={() => setSelectedId(null)}
              aria-hidden="true"
            />

            {/* Drag Handle */}
            <div
              className="absolute top-0 bottom-0 w-1 bg-border cursor-col-resize hover:bg-primary/50 transition-colors z-20"
              style={{ right: infoWidth }}
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
              <div className="absolute inset-y-0 -left-1 -right-1 cursor-col-resize" />
            </div>

            {/* Info Content */}
            <div
              className="absolute top-0 right-0 bottom-0 p-4 bg-background/95 backdrop-blur-sm border-l space-y-4 select-none overflow-y-auto z-30 shadow-lg"
              style={{ width: infoWidth }}
            >
              <div className="flex items-center gap-2 mb-2">
                {(() => {
                  const Icon = getFileIcon(selected.type);
                  return <Icon className="w-8 h-8 text-muted-foreground shrink-0" />;
                })()}
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium truncate" title={selected.name}>
                    {selected.name}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {selected.type || 'unknown type'}
                  </p>
                </div>
              </div>

              {/* Image Preview */}
              {selected.type.startsWith('image/') && selected.url && !imageError && (
                <div className="rounded-md overflow-hidden border bg-muted/30">
                  <img
                    src={selected.url}
                    alt={selected.name}
                    className="w-full h-auto max-h-[200px] object-contain"
                    onError={() => setImageError(true)}
                  />
                </div>
              )}
              {selected.type.startsWith('image/') && imageError && (
                <div className="rounded-md border bg-muted/30 p-8 flex flex-col items-center justify-center text-muted-foreground">
                  <IconAlertCircle className="w-8 h-8 mb-2" />
                  <p className="text-xs text-center">Preview unavailable</p>
                </div>
              )}

              {/* Video Preview */}
              {selected.type.startsWith('video/') && selected.url && !videoError && (
                <div className="rounded-md overflow-hidden border bg-black">
                  <video
                    src={selected.url}
                    controls
                    className="w-full max-h-[300px]"
                    onError={() => setVideoError(true)}
                  />
                </div>
              )}
              {selected.type.startsWith('video/') && videoError && (
                <button
                  type="button"
                  className="rounded-md border bg-muted/30 p-8 flex flex-col items-center justify-center text-muted-foreground w-full cursor-pointer hover:bg-muted/50 transition-colors"
                  onClick={() => selected.url && window.open(selected.url, '_blank', 'noopener,noreferrer')}
                >
                  <IconVideo className="w-8 h-8 mb-2" />
                  <p className="text-xs text-center">Video preview unavailable (click to download)</p>
                </button>
              )}

              {/* Text Preview */}
              {isTextFile(selected.name) && selected.url && (
                <div className="space-y-2">
                  <p className="text-xs text-muted-foreground flex items-center gap-1">
                    <IconFileText className="w-3.5 h-3.5" />
                    Preview
                  </p>
                  {textLoading ? (
                    <div className="rounded-md bg-muted/30 p-4 text-center text-xs text-muted-foreground">
                      Loading...
                    </div>
                  ) : textError ? (
                    <div className="rounded-md bg-muted/30 p-3 text-xs text-muted-foreground">
                      {textError}
                    </div>
                  ) : textContent ? (
                    <pre className="rounded-md bg-gray-900 text-gray-100 p-3 text-xs overflow-y-auto max-h-[40vh] whitespace-pre-wrap">
                      {textContent.split('\n').map((line, i) => (
                        <div key={i} className="flex">
                          <span className="w-8 text-gray-500 select-none text-right pr-2 shrink-0">
                            {i + 1}
                          </span>
                          <span className="flex-1">{line || ' '}</span>
                        </div>
                      ))}
                    </pre>
                  ) : null}
                </div>
              )}

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
