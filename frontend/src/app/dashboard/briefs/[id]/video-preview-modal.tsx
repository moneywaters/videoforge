"use client";

import { useEffect, useRef } from 'react';
import { IconX } from '@tabler/icons-react';

interface VideoPreviewModalProps {
  url: string | null;
  title?: string;
  onClose: () => void;
}

export function VideoPreviewModal({ url, title, onClose }: VideoPreviewModalProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (url) {
      document.body.style.overflow = 'hidden';
      videoRef.current?.play().catch(() => {});
    } else {
      document.body.style.overflow = '';
    }
    return () => {
      document.body.style.overflow = '';
    };
  }, [url]);

  if (!url) return null;

  return (
    <div
      className="fixed inset-0 bg-black/80 z-50 flex items-center justify-center p-4"
      onClick={(e) => {
        if (e.target === containerRef.current) onClose();
      }}
    >
      <div
        ref={containerRef}
        className="relative w-full max-w-5xl bg-black rounded-lg overflow-hidden shadow-2xl"
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-2 bg-black/60 backdrop-blur-sm">
          <h3 className="text-sm font-medium text-white truncate max-w-[80%]">
            {title || 'Video Preview'}
          </h3>
          <button
            onClick={onClose}
            className="p-1.5 rounded-md hover:bg-white/10 transition-colors"
            aria-label="Close preview"
          >
            <IconX className="w-5 h-5 text-white" />
          </button>
        </div>

        {/* Video */}
        <video
          ref={videoRef}
          src={url}
          controls
          className="w-full max-h-[80vh] block"
          playsInline
          preload="metadata"
          autoPlay
          muted
        >
          Your browser does not support the video tag.
        </video>

        {/* Controls hint */}
        <div className="px-4 py-2 bg-black/60 backdrop-blur-sm text-xs text-muted-foreground text-center">
          Click outside the player to close • Native HTML5 video controls
        </div>
      </div>
    </div>
  );
}
