'use client';

import NoteContent from './NoteContent';
import Spinner from '@/components/ui/Spinner';

interface Props {
  question: string;
  answer: string | null;
  loading?: boolean;
}

export default function FollowUpBlock({ question, answer, loading }: Props) {
  return (
    <div style={{ marginBottom: '24px' }}>
      {/* Question card */}
      <div
        style={{
          background: 'var(--surface2)',
          borderLeft: '3px solid var(--accent)',
          borderRadius: 'var(--radius-md)',
          padding: '12px 14px',
          marginBottom: '12px',
        }}
      >
        <div
          style={{
            fontSize: '0.625rem',
            fontWeight: 700,
            textTransform: 'uppercase',
            letterSpacing: '0.08em',
            color: 'var(--accent)',
            marginBottom: '6px',
          }}
        >
          You asked
        </div>
        <p
          style={{
            fontSize: '0.9375rem',
            color: 'var(--text)',
            lineHeight: 1.5,
            fontFamily: 'var(--font-dm-sans, "DM Sans"), sans-serif',
          }}
        >
          {question}
        </p>
      </div>

      {/* Answer */}
      {loading ? (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px', padding: '4px 0' }}>
          <Spinner size={15} />
          <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>
            Generating response…
          </span>
        </div>
      ) : answer ? (
        <NoteContent content={answer} />
      ) : null}
    </div>
  );
}
