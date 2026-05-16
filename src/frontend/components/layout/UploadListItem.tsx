'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { UploadListItem as UploadListItemType } from '@/lib/types';
import { getFileCategory, formatDate } from '@/lib/utils';
import { DocumentIcon, AudioIcon, VideoIcon } from '@/components/ui/Icons';
import Spinner from '@/components/ui/Spinner';

interface Props {
  upload: UploadListItemType;
  isProcessing: boolean;
}

function FileIcon({ filename, size = 15 }: { filename: string; size?: number }) {
  const cat = getFileCategory(filename);
  const color = 'var(--text-secondary)';
  if (cat === 'audio') return <AudioIcon size={size} color={color} />;
  if (cat === 'video') return <VideoIcon size={size} color={color} />;
  return <DocumentIcon size={size} color={color} />;
}

function getExt(filename: string): string {
  return filename.split('.').pop()?.toUpperCase() ?? 'FILE';
}

export default function UploadListItem({ upload, isProcessing }: Props) {
  const pathname = usePathname();
  const isActive = pathname === `/uploads/${upload.id}`;

  return (
    <Link
      href={`/uploads/${upload.id}`}
      style={{ textDecoration: 'none', display: 'block' }}
    >
      <div
        className={isActive ? 'sidebar-item-active' : ''}
        style={{
          display: 'flex',
          alignItems: 'flex-start',
          gap: '8px',
          padding: '8px 12px',
          borderLeft: '2px solid transparent',
          cursor: 'pointer',
          transition: 'background 0.1s, border-color 0.1s',
        }}
        onMouseEnter={(e) => {
          if (!isActive) {
            (e.currentTarget as HTMLElement).style.background = 'var(--surface2)';
          }
        }}
        onMouseLeave={(e) => {
          if (!isActive) {
            (e.currentTarget as HTMLElement).style.background = 'transparent';
          }
        }}
      >
        {/* Icon */}
        <div style={{ flexShrink: 0, marginTop: '1px' }}>
          {isProcessing ? (
            <Spinner size={15} color="var(--warning)" />
          ) : (
            <FileIcon filename={upload.filename} />
          )}
        </div>

        {/* Text */}
        <div style={{ flex: 1, minWidth: 0 }}>
          <div
            className="sidebar-item-name"
            style={{
              fontSize: '0.8125rem',
              fontWeight: 500,
              color: isActive ? 'var(--accent)' : 'var(--text)',
              whiteSpace: 'nowrap',
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              lineHeight: 1.3,
              marginBottom: '3px',
            }}
          >
            {upload.filename}
          </div>
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
            }}
          >
            <span
              style={{
                fontSize: '0.6875rem',
                color: 'var(--text-muted)',
              }}
            >
              {formatDate(upload.created_at)}
            </span>
            <span
              style={{
                fontSize: '0.6875rem',
                color: 'var(--text-muted)',
              }}
            >
              ·
            </span>
            <span
              style={{
                fontSize: '0.6875rem',
                color: 'var(--text-muted)',
              }}
            >
              {getExt(upload.filename)}
            </span>
            {isProcessing && (
              <span
                style={{
                  fontSize: '0.625rem',
                  fontWeight: 600,
                  color: 'var(--warning)',
                  background: 'var(--warning-subtle)',
                  padding: '1px 5px',
                  borderRadius: '99px',
                  textTransform: 'uppercase',
                  letterSpacing: '0.04em',
                }}
              >
                Processing
              </span>
            )}
          </div>
        </div>
      </div>
    </Link>
  );
}
