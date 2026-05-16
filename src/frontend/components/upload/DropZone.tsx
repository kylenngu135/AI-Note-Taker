'use client';

import { useCallback, useRef, useState } from 'react';
import { UploadCloudIcon } from '@/components/ui/Icons';

interface Props {
  onFileSelected: (file: File) => void;
  disabled?: boolean;
}

const ACCEPTED = [
  'application/pdf',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  'text/plain',
  'text/plain; charset=utf-8',
  'video/mp4',
  'audio/mpeg',
];

export default function DropZone({ onFileSelected, disabled }: Props) {
  const [isDragOver, setIsDragOver] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFile = useCallback(
    (file: File) => {
      if (!disabled) onFileSelected(file);
    },
    [onFileSelected, disabled]
  );

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault();
      setIsDragOver(false);
      const file = e.dataTransfer.files[0];
      if (file) handleFile(file);
    },
    [handleFile]
  );

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(true);
  };

  const handleDragLeave = () => setIsDragOver(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) handleFile(file);
    e.target.value = '';
  };

  return (
    <div
      onDrop={handleDrop}
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onClick={() => !disabled && inputRef.current?.click()}
      style={{
        border: `1.5px dashed ${isDragOver ? 'var(--accent)' : 'var(--border-strong)'}`,
        borderRadius: 'var(--radius-md)',
        background: isDragOver ? 'var(--accent-subtle)' : 'var(--bg)',
        padding: '32px 24px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '10px',
        cursor: disabled ? 'default' : 'pointer',
        transition: 'border-color 0.15s, background 0.15s',
      }}
    >
      <UploadCloudIcon
        size={32}
        color={isDragOver ? 'var(--accent)' : 'var(--text-muted)'}
      />
      <div style={{ textAlign: 'center' }}>
        <p style={{ fontSize: '0.875rem', color: 'var(--text)', marginBottom: '4px' }}>
          Drag and drop your file here
        </p>
        <p style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)' }}>
          or{' '}
          <span style={{ color: 'var(--accent)', textDecoration: 'underline' }}>
            browse to upload
          </span>
        </p>
      </div>
      <input
        ref={inputRef}
        type="file"
        accept=".pdf,.docx,.txt,.mp4,.mp3"
        style={{ display: 'none' }}
        onChange={handleChange}
        disabled={disabled}
      />
    </div>
  );
}
