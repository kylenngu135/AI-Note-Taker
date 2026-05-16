'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { deleteUpload, getNotes, regenerateNote } from '@/lib/api';
import { useApp } from '@/context/AppContext';
import { useToast } from '@/components/ui/Toast';
import NoteContent from '@/components/notes/NoteContent';
import FollowUpBlock from '@/components/notes/FollowUpBlock';
import FollowUpInput from '@/components/notes/FollowUpInput';
import Skeleton from '@/components/ui/Skeleton';
import Spinner from '@/components/ui/Spinner';
import { TrashIcon } from '@/components/ui/Icons';
import { formatDate, getFileCategory } from '@/lib/utils';
import type { NoteHistory, NoteWithHistory } from '@/lib/types';

const POLL_INTERVAL_MS = 3000;
const MAX_POLL_MS = 5 * 60 * 1000;

interface OptimisticFollowUp {
  question: string;
  loading: boolean;
  answer: string | null;
}

function ProcessingSkeleton() {
  return (
    <div style={{ padding: '24px' }}>
      <div style={{ maxWidth: 'var(--content-max)', margin: '0 auto' }}>
        {/* Title skeleton */}
        <div style={{ marginBottom: '8px' }}>
          <Skeleton width="200px" height={22} />
        </div>
        <Skeleton width="140px" height={13} />

        <div style={{ marginTop: '32px', display: 'flex', flexDirection: 'column', gap: '10px' }}>
          <Skeleton width="100%" height={14} />
          <Skeleton width="90%" height={14} />
          <Skeleton width="75%" height={14} />
          <Skeleton width="100%" height={14} />
          <Skeleton width="70%" height={14} />
          <div style={{ marginTop: '8px' }} />
          <Skeleton width="95%" height={14} />
          <Skeleton width="88%" height={14} />
          <Skeleton width="60%" height={14} />
          <Skeleton width="100%" height={14} />
          <Skeleton width="80%" height={14} />
        </div>

        <div
          style={{
            marginTop: '24px',
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
          }}
        >
          <Spinner size={14} color="var(--text-muted)" />
          <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>
            Transcribing and generating your study sheet…
          </span>
        </div>
      </div>
    </div>
  );
}

function getFollowUpPairs(history: NoteHistory[]): Array<{ question: string; answer: string }> {
  const pairs: Array<{ question: string; answer: string }> = [];

  // history[0] = user (raw transcription) → skip
  // history[1] = assistant (initial note) → already shown
  // history[2..] = alternating user/assistant pairs

  let i = 0;
  // Skip raw transcription entry (first user entry with no content)
  while (i < history.length && !(history[i].role === 'user' && history[i].content === '')) {
    i++;
  }
  i++; // skip the transcription

  // Skip the first assistant entry (shown as main note)
  if (i < history.length && history[i].role === 'assistant') {
    i++;
  }

  // Remaining are follow-up pairs
  while (i < history.length) {
    const userEntry = history[i];
    const assistantEntry = history[i + 1];
    if (userEntry?.role === 'user' && userEntry.prompt) {
      pairs.push({
        question: userEntry.prompt,
        answer: assistantEntry?.content ?? '',
      });
    }
    i += 2;
  }

  return pairs;
}

export default function UploadPage() {
  const params = useParams();
  const id = params.id as string;
  const router = useRouter();
  const { uploads, processingIds, removeProcessingId, refreshUploads } = useApp();
  const { addToast } = useToast();

  const [noteData, setNoteData] = useState<NoteWithHistory | null>(null);
  const [loading, setLoading] = useState(true);
  const [processing, setProcessing] = useState(false);
  const [optimistic, setOptimistic] = useState<OptimisticFollowUp[]>([]);
  const [deleteConfirm, setDeleteConfirm] = useState(false);
  const [deleting, setDeleting] = useState(false);

  const pollTimer = useRef<ReturnType<typeof setInterval> | null>(null);
  const pollStarted = useRef<number>(0);

  const stopPolling = useCallback(() => {
    if (pollTimer.current) {
      clearInterval(pollTimer.current);
      pollTimer.current = null;
    }
  }, []);

  const fetchNotes = useCallback(async (): Promise<boolean> => {
    try {
      const data = await getNotes(id);
      if (data.note.content) {
        setNoteData(data);
        setLoading(false);
        setProcessing(false);
        removeProcessingId(id);
        return true;
      }
    } catch {
      // still processing — continue polling
    }
    return false;
  }, [id, removeProcessingId]);

  const startPolling = useCallback(() => {
    pollStarted.current = Date.now();
    pollTimer.current = setInterval(async () => {
      if (Date.now() - pollStarted.current > MAX_POLL_MS) {
        stopPolling();
        setLoading(false);
        setProcessing(false);
        addToast('Processing is taking longer than expected. Please refresh.', 'error');
        return;
      }
      const done = await fetchNotes();
      if (done) stopPolling();
    }, POLL_INTERVAL_MS);
  }, [fetchNotes, stopPolling, addToast]);

  // Initial load
  useEffect(() => {
    stopPolling();
    setLoading(true);
    setNoteData(null);
    setOptimistic([]);
    setDeleteConfirm(false);

    const isProcessingNow = processingIds.has(id);

    if (isProcessingNow) {
      setProcessing(true);
      setLoading(false);
      startPolling();
      return;
    }

    fetchNotes().then((done) => {
      if (!done) {
        setProcessing(true);
        setLoading(false);
        startPolling();
      }
    });

    return () => stopPolling();
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [id]);

  const handleSend = async (prompt: string) => {
    const idx = optimistic.length;
    setOptimistic((prev) => [
      ...prev,
      { question: prompt, loading: true, answer: null },
    ]);

    try {
      await regenerateNote(id, prompt);
      const updated = await getNotes(id);
      setNoteData(updated);
      setOptimistic((prev) => prev.filter((_, i) => i !== idx));
    } catch (err) {
      setOptimistic((prev) => prev.filter((_, i) => i !== idx));
      addToast(
        err instanceof Error ? err.message : 'Failed to send follow-up.',
        'error'
      );
    }
  };

  const handleDelete = async () => {
    setDeleting(true);
    try {
      await deleteUpload(id);
      await refreshUploads();
      addToast('Upload deleted.', 'success');

      // Navigate to next upload or empty state
      const remaining = uploads.filter((u) => u.id !== id);
      if (remaining.length > 0) {
        router.replace(`/uploads/${remaining[0].id}`);
      } else {
        router.replace('/');
      }
    } catch (err) {
      setDeleting(false);
      setDeleteConfirm(false);
      addToast(
        err instanceof Error ? err.message : 'Failed to delete upload.',
        'error'
      );
    }
  };

  // Find the current upload metadata from the uploads list
  const uploadMeta = uploads.find((u) => u.id === id);
  const filename = uploadMeta?.filename ?? (noteData ? 'Document' : '');
  const createdAt = uploadMeta?.created_at;
  const fileExt = filename.split('.').pop()?.toUpperCase() ?? '';

  if (loading) {
    return <ProcessingSkeleton />;
  }

  if (processing && !noteData) {
    return <ProcessingSkeleton />;
  }

  if (!noteData) {
    return (
      <div
        style={{
          flex: 1,
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: 'var(--text-muted)',
          fontSize: '0.875rem',
        }}
      >
        Note not found.
      </div>
    );
  }

  const followUpPairs = getFollowUpPairs(noteData.history);

  return (
    <div
      style={{
        flex: 1,
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
        background: 'var(--bg)',
      }}
    >
      {/* Scrollable content */}
      <div style={{ flex: 1, overflowY: 'auto' }}>
        <div
          style={{
            maxWidth: 'var(--content-max)',
            margin: '0 auto',
            padding: '24px',
          }}
        >
          {/* Title row */}
          <div
            style={{
              display: 'flex',
              alignItems: 'flex-start',
              justifyContent: 'space-between',
              gap: '16px',
              marginBottom: '16px',
            }}
          >
            <div>
              <h1
                style={{
                  fontFamily: 'var(--font-dm-serif-display, "DM Serif Display"), serif',
                  fontSize: '1.375rem',
                  fontWeight: 400,
                  color: 'var(--text)',
                  lineHeight: 1.3,
                  marginBottom: '6px',
                  wordBreak: 'break-word',
                }}
              >
                {filename}
              </h1>
              <div
                style={{
                  display: 'flex',
                  gap: '8px',
                  alignItems: 'center',
                  fontSize: '0.75rem',
                  color: 'var(--text-muted)',
                }}
              >
                {createdAt && <span>{formatDate(createdAt)}</span>}
                {createdAt && fileExt && <span>·</span>}
                {fileExt && <span>{fileExt}</span>}
              </div>
            </div>

            {/* Delete button / confirmation */}
            <div style={{ flexShrink: 0, display: 'flex', gap: '6px', alignItems: 'center' }}>
              {deleteConfirm ? (
                <>
                  <span style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)' }}>
                    Are you sure?
                  </span>
                  <button
                    onClick={() => setDeleteConfirm(false)}
                    disabled={deleting}
                    style={{
                      padding: '4px 10px',
                      background: 'transparent',
                      border: '0.5px solid var(--border)',
                      borderRadius: 'var(--radius-sm)',
                      fontSize: '0.8125rem',
                      color: 'var(--text-secondary)',
                      cursor: 'pointer',
                    }}
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleDelete}
                    disabled={deleting}
                    style={{
                      padding: '4px 10px',
                      background: 'var(--danger)',
                      border: 'none',
                      borderRadius: 'var(--radius-sm)',
                      fontSize: '0.8125rem',
                      color: '#fff',
                      cursor: deleting ? 'not-allowed' : 'pointer',
                      opacity: deleting ? 0.7 : 1,
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                    }}
                  >
                    {deleting ? <Spinner size={12} color="#fff" /> : null}
                    {deleting ? 'Deleting…' : 'Delete'}
                  </button>
                </>
              ) : (
                <button
                  onClick={() => setDeleteConfirm(true)}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: '5px',
                    padding: '4px 10px',
                    background: 'transparent',
                    border: '0.5px solid var(--border)',
                    borderRadius: 'var(--radius-sm)',
                    fontSize: '0.8125rem',
                    color: 'var(--text-secondary)',
                    cursor: 'pointer',
                  }}
                >
                  <TrashIcon size={12} />
                  Delete
                </button>
              )}
            </div>
          </div>

          {/* Study sheet label */}
          <div
            style={{
              fontSize: '0.6875rem',
              fontWeight: 700,
              textTransform: 'uppercase',
              letterSpacing: '0.08em',
              color: 'var(--text-muted)',
              marginBottom: '12px',
            }}
          >
            Study sheet
          </div>

          {/* Main note content */}
          <NoteContent content={noteData.note.content} />

          {/* Divider + Follow-ups */}
          {(followUpPairs.length > 0 || optimistic.length > 0) && (
            <>
              <hr
                style={{
                  border: 'none',
                  borderTop: '1px solid var(--border)',
                  margin: '24px 0',
                }}
              />
              <div
                style={{
                  fontSize: '0.6875rem',
                  fontWeight: 700,
                  textTransform: 'uppercase',
                  letterSpacing: '0.08em',
                  color: 'var(--text-muted)',
                  marginBottom: '20px',
                }}
              >
                Follow-ups
              </div>

              {followUpPairs.map((pair, i) => (
                <FollowUpBlock
                  key={i}
                  question={pair.question}
                  answer={pair.answer}
                />
              ))}

              {optimistic.map((opt, i) => (
                <FollowUpBlock
                  key={`opt-${i}`}
                  question={opt.question}
                  answer={opt.answer}
                  loading={opt.loading}
                />
              ))}
            </>
          )}

          {/* Bottom padding for sticky input */}
          <div style={{ height: '16px' }} />
        </div>
      </div>

      {/* Sticky follow-up input */}
      <FollowUpInput
        uploadId={id}
        onSend={handleSend}
        disabled={processing}
      />
    </div>
  );
}
