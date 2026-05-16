'use client';

import {
  createContext,
  useCallback,
  useContext,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import type { ToastItem, ToastType } from '@/lib/types';

interface ToastContextValue {
  addToast: (message: string, type?: ToastType) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

export function useToast(): ToastContextValue {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used within ToastProvider');
  return ctx;
}

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const counter = useRef(0);

  const addToast = useCallback((message: string, type: ToastType = 'info') => {
    const id = String(++counter.current);
    setToasts((prev) => [...prev, { id, message, type }]);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 3000);
  }, []);

  const dismiss = useCallback((id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const borderColor: Record<ToastType, string> = {
    success: 'var(--success)',
    error: 'var(--danger)',
    info: 'var(--accent)',
  };

  return (
    <ToastContext.Provider value={{ addToast }}>
      {children}
      <div
        style={{
          position: 'fixed',
          top: '16px',
          right: '16px',
          zIndex: 9999,
          display: 'flex',
          flexDirection: 'column',
          gap: '8px',
          pointerEvents: 'none',
        }}
        aria-live="polite"
      >
        {toasts.map((t) => (
          <div
            key={t.id}
            className="toast-enter"
            style={{
              background: 'var(--surface)',
              border: '0.5px solid var(--border)',
              borderLeft: `3px solid ${borderColor[t.type]}`,
              borderRadius: 'var(--radius-md)',
              padding: '10px 14px',
              minWidth: '240px',
              maxWidth: '320px',
              boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
              display: 'flex',
              alignItems: 'flex-start',
              gap: '10px',
              pointerEvents: 'all',
            }}
          >
            <span
              style={{
                flex: 1,
                fontSize: '0.8125rem',
                color: 'var(--text)',
                lineHeight: 1.4,
              }}
            >
              {t.message}
            </span>
            <button
              onClick={() => dismiss(t.id)}
              style={{
                background: 'none',
                border: 'none',
                color: 'var(--text-muted)',
                fontSize: '14px',
                lineHeight: 1,
                padding: 0,
                cursor: 'pointer',
                flexShrink: 0,
              }}
              aria-label="Dismiss"
            >
              ×
            </button>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}
