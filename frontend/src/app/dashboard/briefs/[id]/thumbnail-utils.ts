"use client";

/**
 * Generate a thumbnail Blob from a File (image or video).
 * For images: resizes to max 256px using canvas.
 * For videos: captures the first frame (at 1s or 0.1s) using canvas.
 * Returns null for unsupported types or on error.
 */
export async function generateThumbnail(file: File): Promise<string | null> {
  if (file.type.startsWith("image/")) {
    return generateImageThumbnail(file);
  }
  if (file.type.startsWith("video/")) {
    return generateVideoThumbnail(file);
  }
  return null;
}

function generateImageThumbnail(file: File): Promise<string | null> {
  return new Promise((resolve) => {
    const img = new Image();
    const url = URL.createObjectURL(file);

    img.onload = () => {
      URL.revokeObjectURL(url);
      const canvas = document.createElement("canvas");
      const ctx = canvas.getContext("2d");
      if (!ctx) {
        resolve(null);
        return;
      }

      const maxSize = 256;
      let { width, height } = img;
      if (width > height) {
        if (width > maxSize) {
          height = Math.round((height * maxSize) / width);
          width = maxSize;
        }
      } else {
        if (height > maxSize) {
          width = Math.round((width * maxSize) / height);
          height = maxSize;
        }
      }

      canvas.width = width;
      canvas.height = height;
      ctx.drawImage(img, 0, 0, width, height);
      resolve(canvas.toDataURL("image/jpeg", 0.82));
    };

    img.onerror = () => {
      URL.revokeObjectURL(url);
      resolve(null);
    };

    img.src = url;
  });
}

function generateVideoThumbnail(file: File): Promise<string | null> {
  return new Promise((resolve) => {
    const video = document.createElement("video");
    const url = URL.createObjectURL(file);

    video.src = url;
    video.crossOrigin = "anonymous";
    video.muted = true;
    video.playsInline = true;

    video.onloadeddata = () => {
      const seekTime = Math.min(1, video.duration * 0.1 || 0.1);
      video.currentTime = seekTime;
    };

    video.onseeked = () => {
      const canvas = document.createElement("canvas");
      const ctx = canvas.getContext("2d");
      if (!ctx) {
        URL.revokeObjectURL(url);
        video.src = "";
        resolve(null);
        return;
      }

      const maxSize = 256;
      let { videoWidth: width, videoHeight: height } = video;
      if (width > height) {
        if (width > maxSize) {
          height = Math.round((height * maxSize) / width);
          width = maxSize;
        }
      } else {
        if (height > maxSize) {
          width = Math.round((width * maxSize) / height);
          height = maxSize;
        }
      }

      canvas.width = width;
      canvas.height = height;
      ctx.drawImage(video, 0, 0, width, height);
      URL.revokeObjectURL(url);
      video.src = "";
      resolve(canvas.toDataURL("image/jpeg", 0.82));
    };

    video.onerror = () => {
      URL.revokeObjectURL(url);
      resolve(null);
    };

    video.load();
  });
}
