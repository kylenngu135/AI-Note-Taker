'use client';

import { useApp } from '@/context/AppContext';
import UploadListItem from './UploadListItem';
import Button from '@/components/ui/Button';

export default function Sidebar() {
  const { uploads, processingIds, openUploadModal } = useApp();

  return (
    <aside
      style={{
        width: 'var(--sidebar-width)',
        flexShrink: 0,
        borderRight: '0.5px solid var(--border)',
        display: 'flex',
        flexDirection: 'column',
        background: 'var(--surface)',
        height: '100%',
        overflow: 'hidden',
      }}
    >
      {/* New upload button */}
      <div style={{ padding: '12px' }}>
        <button
          onClick={openUploadModal}
          style={{
            width: '100%',
            height: '32px',
            background: 'var(--accent)',
            color: '#fff',
            border: 'none',
            borderRadius: 'var(--radius-sm)',
            fontSize: '0.8125rem',
            fontWeight: 500,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            gap: '6px',
          }}
        >
          <span style={{ fontSize: '16px', lineHeight: 1, marginTop: '-1px' }}>+</span>
          New upload
        </button>
      </div>

      {/* Uploads section */}
      <div
        style={{
          padding: '0 12px 4px',
          fontSize: '0.6875rem',
          fontWeight: 600,
          textTransform: 'uppercase',
          letterSpacing: '0.06em',
          color: 'var(--text-muted)',
        }}
      >
        Uploads
      </div>

      {/* Upload list */}
      <div style={{ flex: 1, overflowY: 'auto' }}>
        {uploads.length === 0 ? (
          <div
            style={{
              padding: '16px 12px',
              fontSize: '0.8125rem',
              color: 'var(--text-muted)',
              textAlign: 'center',
            }}
          >
            No uploads yet
          </div>
        ) : (
          uploads.map((upload) => (
            <UploadListItem
              key={upload.id}
              upload={upload}
              isProcessing={processingIds.has(upload.id)}
            />
          ))
        )}
      </div>
    </aside>
  );
}
