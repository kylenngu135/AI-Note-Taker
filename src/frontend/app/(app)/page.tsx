'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useApp } from '@/context/AppContext';
import { UploadFolderIcon } from '@/components/ui/Icons';

export default function HomePage() {
  const { uploads, authLoading, openUploadModal } = useApp();
  const router = useRouter();

  useEffect(() => {
    if (!authLoading && uploads.length > 0) {
      router.replace(`/uploads/${uploads[0].id}`);
    }
  }, [uploads, authLoading, router]);

  if (authLoading) return null;
  if (uploads.length > 0) return null;

  return (
    <div
      style={{
        flex: 1,
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '48px 24px',
      }}
    >
      <div
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          gap: '16px',
          textAlign: 'center',
          maxWidth: '320px',
        }}
      >
        <div
          style={{
            width: '56px',
            height: '56px',
            background: 'var(--surface2)',
            borderRadius: '12px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <UploadFolderIcon size={32} color="var(--text-muted)" />
        </div>

        <div>
          <h1
            style={{
              fontSize: '1rem',
              fontWeight: 500,
              color: 'var(--text)',
              marginBottom: '6px',
            }}
          >
            Upload your first document
          </h1>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', lineHeight: 1.5 }}>
            Upload a PDF, Word doc, text file, MP4 video, or MP3 audio file to get
            AI-generated study notes.
          </p>
        </div>

        <button
          onClick={openUploadModal}
          style={{
            padding: '0 18px',
            height: '36px',
            background: 'var(--accent)',
            color: '#fff',
            border: 'none',
            borderRadius: 'var(--radius-sm)',
            fontSize: '0.875rem',
            fontWeight: 500,
            cursor: 'pointer',
            marginTop: '4px',
          }}
        >
          Upload a file
        </button>
      </div>
    </div>
  );
}
