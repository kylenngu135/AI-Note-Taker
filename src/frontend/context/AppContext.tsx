'use client';

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
  type ReactNode,
} from 'react';
import { getMe, getUploads } from '@/lib/api';
import type { MeResponse, UploadListItem } from '@/lib/types';

interface AppContextValue {
  user: MeResponse | null;
  uploads: UploadListItem[];
  processingIds: Set<string>;
  authLoading: boolean;
  refreshUploads: () => Promise<void>;
  addProcessingId: (id: string) => void;
  removeProcessingId: (id: string) => void;
  isUploadModalOpen: boolean;
  openUploadModal: () => void;
  closeUploadModal: () => void;
}

const AppContext = createContext<AppContextValue | null>(null);

export function useApp(): AppContextValue {
  const ctx = useContext(AppContext);
  if (!ctx) throw new Error('useApp must be used within AppProvider');
  return ctx;
}

export function AppProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<MeResponse | null>(null);
  const [uploads, setUploads] = useState<UploadListItem[]>([]);
  const [processingIds, setProcessingIds] = useState<Set<string>>(new Set());
  const [authLoading, setAuthLoading] = useState(true);
  const [isUploadModalOpen, setIsUploadModalOpen] = useState(false);
  const initialized = useRef(false);

  const refreshUploads = useCallback(async () => {
    try {
      const data = await getUploads();
      setUploads(data);
    } catch {
      // silently ignore
    }
  }, []);

  const addProcessingId = useCallback((id: string) => {
    setProcessingIds((prev) => new Set(prev).add(id));
  }, []);

  const removeProcessingId = useCallback((id: string) => {
    setProcessingIds((prev) => {
      const next = new Set(prev);
      next.delete(id);
      return next;
    });
  }, []);

  const openUploadModal = useCallback(() => setIsUploadModalOpen(true), []);
  const closeUploadModal = useCallback(() => setIsUploadModalOpen(false), []);

  useEffect(() => {
    if (initialized.current) return;
    initialized.current = true;

    const init = async () => {
      try {
        const userData = await getMe();
        setUser(userData);
        await refreshUploads();
      } catch {
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
      } finally {
        setAuthLoading(false);
      }
    };

    init();
  }, [refreshUploads]);

  const value: AppContextValue = {
    user,
    uploads,
    processingIds,
    authLoading,
    refreshUploads,
    addProcessingId,
    removeProcessingId,
    isUploadModalOpen,
    openUploadModal,
    closeUploadModal,
  };

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
}
