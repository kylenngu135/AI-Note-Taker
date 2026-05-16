import type { FileCategory } from './types';

export function getFileCategory(filename: string): FileCategory {
  const ext = filename.split('.').pop()?.toLowerCase() ?? '';
  if (ext === 'mp4') return 'video';
  if (ext === 'mp3') return 'audio';
  return 'document';
}

export function formatDate(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

export function truncate(str: string, max: number): string {
  if (str.length <= max) return str;
  return str.slice(0, max - 1) + '…';
}
