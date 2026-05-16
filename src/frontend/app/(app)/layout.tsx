'use client';

import { AppProvider } from '@/context/AppContext';
import Header from '@/components/layout/Header';
import Sidebar from '@/components/layout/Sidebar';
import UploadModal from '@/components/upload/UploadModal';
import { useApp } from '@/context/AppContext';

function AppShell({ children }: { children: React.ReactNode }) {
  const { authLoading, isUploadModalOpen } = useApp();

  if (authLoading) {
    return (
      <div
        style={{
          height: '100vh',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          background: 'var(--bg)',
        }}
      >
        <div
          style={{
            width: '20px',
            height: '20px',
            border: '2px solid var(--border)',
            borderTopColor: 'var(--accent)',
            borderRadius: '50%',
            animation: 'spin 0.8s linear infinite',
          }}
        />
      </div>
    );
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100vh' }}>
      <Header />

      <div
        style={{
          display: 'flex',
          flex: 1,
          marginTop: 'var(--header-height)',
          overflow: 'hidden',
        }}
      >
        <Sidebar />
        <main
          style={{
            flex: 1,
            overflow: 'hidden',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          {children}
        </main>
      </div>

      {isUploadModalOpen && <UploadModal />}
    </div>
  );
}

export default function AppLayout({ children }: { children: React.ReactNode }) {
  return (
    <AppProvider>
      <AppShell>{children}</AppShell>
    </AppProvider>
  );
}
