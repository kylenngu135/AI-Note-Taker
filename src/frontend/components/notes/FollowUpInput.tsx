'use client';

import { useRef, useState } from 'react';
import { getExportUrl } from '@/lib/api';
import { SendIcon } from '@/components/ui/Icons';
import Spinner from '@/components/ui/Spinner';

interface Props {
  uploadId: string;
  onSend: (prompt: string) => Promise<void>;
  disabled?: boolean;
}

export default function FollowUpInput({ uploadId, onSend, disabled }: Props) {
  const [text, setText] = useState('');
  const [sending, setSending] = useState(false);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const canSend = text.trim().length > 0 && !sending && !disabled;

  const handleSend = async () => {
    if (!canSend) return;
    const prompt = text.trim();
    setText('');
    setSending(true);
    try {
      await onSend(prompt);
    } finally {
      setSending(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleExport = (format: 'txt' | 'pdf' | 'docx') => {
    window.location.href = getExportUrl(uploadId, format);
  };

  return (
    <div
      style={{
        borderTop: '0.5px solid var(--border)',
        background: 'var(--surface)',
        padding: '12px 24px 16px',
      }}
    >
      {/* Textarea */}
      <textarea
        ref={textareaRef}
        value={text}
        onChange={(e) => setText(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder="Ask a follow-up question… (⌘+Enter to send)"
        rows={3}
        style={{
          width: '100%',
          padding: '10px 12px',
          border: '0.5px solid var(--border-strong)',
          borderRadius: 'var(--radius-md)',
          background: 'var(--bg)',
          color: 'var(--text)',
          fontSize: '0.875rem',
          lineHeight: 1.5,
          resize: 'none',
          outline: 'none',
          fontFamily: 'var(--font-dm-sans, "DM Sans"), sans-serif',
          transition: 'border-color 0.15s',
        }}
        onFocus={(e) => {
          e.currentTarget.style.borderColor = 'var(--accent)';
        }}
        onBlur={(e) => {
          e.currentTarget.style.borderColor = 'var(--border-strong)';
        }}
        disabled={disabled || sending}
      />

      {/* Actions row */}
      <div
        style={{
          marginTop: '8px',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          gap: '8px',
        }}
      >
        {/* Export buttons */}
        <div style={{ display: 'flex', gap: '6px' }}>
          {(['txt', 'pdf', 'docx'] as const).map((fmt) => (
            <button
              key={fmt}
              onClick={() => handleExport(fmt)}
              style={{
                padding: '5px 10px',
                background: 'transparent',
                border: '0.5px solid var(--border)',
                borderRadius: 'var(--radius-sm)',
                fontSize: '0.75rem',
                fontWeight: 500,
                color: 'var(--text-secondary)',
                cursor: 'pointer',
                textTransform: 'uppercase',
                letterSpacing: '0.04em',
              }}
            >
              {fmt}
            </button>
          ))}
        </div>

        {/* Send button */}
        <button
          onClick={handleSend}
          disabled={!canSend}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '6px',
            padding: '0 14px',
            height: '32px',
            background: canSend ? 'var(--accent)' : 'var(--border)',
            color: canSend ? '#fff' : 'var(--text-muted)',
            border: 'none',
            borderRadius: 'var(--radius-sm)',
            fontSize: '0.8125rem',
            fontWeight: 500,
            cursor: canSend ? 'pointer' : 'not-allowed',
            transition: 'background 0.15s, color 0.15s',
          }}
        >
          {sending ? (
            <>
              <Spinner size={13} color="currentColor" />
              Sending
            </>
          ) : (
            <>
              <SendIcon size={13} color="currentColor" />
              Send
            </>
          )}
        </button>
      </div>
    </div>
  );
}
