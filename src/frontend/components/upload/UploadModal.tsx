'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { uploadFile } from '@/lib/api';
import { useApp } from '@/context/AppContext';
import { useToast } from '@/components/ui/Toast';
import DropZone from './DropZone';
import Button from '@/components/ui/Button';
import Spinner from '@/components/ui/Spinner';

export default function UploadModal() {
  const { closeUploadModal, addProcessingId, refreshUploads } = useApp();
  const { addToast } = useToast();
  const router = useRouter();

  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [uploading, setUploading] = useState(false);

  const handleUpload = async () => {
    if (!selectedFile || uploading) return;

    setUploading(true);
    try {
      const { upload_id } = await uploadFile(selectedFile);
      addProcessingId(upload_id);
      await refreshUploads();
      closeUploadModal();
      router.push(`/uploads/${upload_id}`);
      addToast('File uploaded — generating study sheet…', 'info');
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : 'Upload failed. Please try again.',
        'error'
      );
      setUploading(false);
    }
  };

  return (
    <>
      {/* Backdrop */}
      <div
        style={{
          position: 'fixed',
          inset: 0,
          background: 'rgba(0,0,0,0.32)',
          zIndex: 200,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: '16px',
        }}
        onClick={(e) => {
          if (e.target === e.currentTarget && !uploading) closeUploadModal();
        }}
      >
        {/* Modal */}
        <div
          style={{
            background: 'var(--surface)',
            borderRadius: 'var(--radius-lg)',
            padding: '24px',
            width: '100%',
            maxWidth: '400px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
          }}
          onClick={(e) => e.stopPropagation()}
        >
          <h2
            style={{
              fontSize: '1rem',
              fontWeight: 600,
              color: 'var(--text)',
              marginBottom: '4px',
            }}
          >
            Upload a file
          </h2>
          <p
            style={{
              fontSize: '0.8125rem',
              color: 'var(--text-secondary)',
              marginBottom: '16px',
            }}
          >
            PDF, DOCX, TXT, MP4, or MP3 — up to 100 MB
          </p>

          <DropZone
            onFileSelected={setSelectedFile}
            disabled={uploading}
          />

          {selectedFile && (
            <div
              style={{
                marginTop: '12px',
                padding: '8px 10px',
                background: 'var(--surface2)',
                borderRadius: 'var(--radius-sm)',
                fontSize: '0.8125rem',
                color: 'var(--text)',
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
              }}
            >
              <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {selectedFile.name}
              </span>
              <button
                onClick={() => !uploading && setSelectedFile(null)}
                style={{
                  background: 'none',
                  border: 'none',
                  color: 'var(--text-muted)',
                  cursor: 'pointer',
                  fontSize: '14px',
                  padding: '0 2px',
                  flexShrink: 0,
                }}
              >
                ×
              </button>
            </div>
          )}

          {/* Footer */}
          <div
            style={{
              marginTop: '20px',
              display: 'flex',
              justifyContent: 'flex-end',
              gap: '8px',
            }}
          >
            <Button
              variant="secondary"
              onClick={closeUploadModal}
              disabled={uploading}
            >
              Cancel
            </Button>
            <Button
              variant="primary"
              onClick={handleUpload}
              disabled={!selectedFile || uploading}
              style={{
                minWidth: '80px',
                gap: '6px',
              }}
            >
              {uploading ? (
                <>
                  <Spinner size={13} color="#fff" />
                  Uploading…
                </>
              ) : (
                'Upload'
              )}
            </Button>
          </div>
        </div>
      </div>
    </>
  );
}
