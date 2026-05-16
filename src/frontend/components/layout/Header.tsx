'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { logout } from '@/lib/api';
import { NotesIcon, ChevronDownIcon } from '@/components/ui/Icons';
import { useApp } from '@/context/AppContext';
import { useToast } from '@/components/ui/Toast';

export default function Header() {
  const { user } = useApp();
  const { addToast } = useToast();
  const router = useRouter();
  const [menuOpen, setMenuOpen] = useState(false);

  const handleLogout = async () => {
    try {
      await logout();
    } finally {
      router.replace('/login');
    }
  };

  return (
    <header
      style={{
        height: 'var(--header-height)',
        background: 'var(--surface)',
        borderBottom: '0.5px solid var(--border)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '0 16px',
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        zIndex: 100,
      }}
    >
      {/* Logo */}
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
        <div
          style={{
            width: '28px',
            height: '28px',
            background: 'var(--accent)',
            borderRadius: '6px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
          }}
        >
          <NotesIcon size={16} color="#fff" />
        </div>
        <span
          style={{
            fontSize: '0.9375rem',
            fontWeight: 600,
            color: 'var(--text)',
            letterSpacing: '-0.01em',
          }}
        >
          Note Taker
        </span>
      </div>

      {/* User chip */}
      {user && (
        <div style={{ position: 'relative' }}>
          <button
            onClick={() => setMenuOpen((v) => !v)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '6px',
              padding: '4px 10px',
              background: 'var(--surface2)',
              border: '0.5px solid var(--border)',
              borderRadius: '99px',
              cursor: 'pointer',
              fontSize: '0.8125rem',
              color: 'var(--text)',
              fontWeight: 500,
            }}
          >
            {user.email}
            <ChevronDownIcon color="var(--text-muted)" />
          </button>

          {menuOpen && (
            <>
              <div
                style={{
                  position: 'fixed',
                  inset: 0,
                  zIndex: 10,
                }}
                onClick={() => setMenuOpen(false)}
              />
              <div
                style={{
                  position: 'absolute',
                  top: 'calc(100% + 6px)',
                  right: 0,
                  background: 'var(--surface)',
                  border: '0.5px solid var(--border)',
                  borderRadius: 'var(--radius-md)',
                  boxShadow: '0 1px 3px rgba(0,0,0,0.06)',
                  overflow: 'hidden',
                  zIndex: 20,
                  minWidth: '140px',
                }}
              >
                <button
                  onClick={handleLogout}
                  style={{
                    width: '100%',
                    padding: '9px 14px',
                    textAlign: 'left',
                    background: 'none',
                    border: 'none',
                    fontSize: '0.8125rem',
                    color: 'var(--danger)',
                    cursor: 'pointer',
                  }}
                >
                  Sign out
                </button>
              </div>
            </>
          )}
        </div>
      )}
    </header>
  );
}
