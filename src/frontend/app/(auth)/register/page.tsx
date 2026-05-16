'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { register } from '@/lib/api';
import { NotesIcon } from '@/components/ui/Icons';
import Input from '@/components/ui/Input';
import Spinner from '@/components/ui/Spinner';

export default function RegisterPage() {
  const router = useRouter();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await register(email, password);
      router.replace('/');
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Registration failed. Please try again.'
      );
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        background: 'var(--bg)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: '24px',
      }}
    >
      <div
        style={{
          background: 'var(--surface)',
          border: '0.5px solid var(--border)',
          borderRadius: 'var(--radius-lg)',
          padding: '28px',
          width: '100%',
          maxWidth: '320px',
        }}
      >
        {/* Logo + title */}
        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            gap: '10px',
            marginBottom: '24px',
          }}
        >
          <div
            style={{
              width: '36px',
              height: '36px',
              background: 'var(--accent)',
              borderRadius: '8px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <NotesIcon size={20} color="#fff" />
          </div>
          <div style={{ textAlign: 'center' }}>
            <div
              style={{
                fontSize: '1rem',
                fontWeight: 600,
                color: 'var(--text)',
                marginBottom: '4px',
              }}
            >
              Note Taker
            </div>
            <div style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)' }}>
              Create your account
            </div>
          </div>
        </div>

        <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          <Input
            id="email"
            label="Email"
            type="email"
            placeholder="you@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            autoComplete="email"
            autoFocus
          />
          <Input
            id="password"
            label="Password"
            type="password"
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
            autoComplete="new-password"
          />

          {error && (
            <p style={{ fontSize: '0.8125rem', color: 'var(--danger)' }}>{error}</p>
          )}

          <button
            type="submit"
            disabled={loading}
            style={{
              width: '100%',
              height: '36px',
              background: loading ? 'var(--border)' : 'var(--accent)',
              color: loading ? 'var(--text-muted)' : '#fff',
              border: 'none',
              borderRadius: 'var(--radius-sm)',
              fontSize: '0.875rem',
              fontWeight: 500,
              cursor: loading ? 'not-allowed' : 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              gap: '8px',
              marginTop: '4px',
              transition: 'background 0.15s',
            }}
          >
            {loading && <Spinner size={14} color="var(--text-muted)" />}
            {loading ? 'Creating account…' : 'Create account'}
          </button>
        </form>

        <p
          style={{
            textAlign: 'center',
            marginTop: '16px',
            fontSize: '0.8125rem',
            color: 'var(--text-secondary)',
          }}
        >
          Already have an account?{' '}
          <Link
            href="/login"
            style={{ color: 'var(--accent)', textDecoration: 'none', fontWeight: 500 }}
          >
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
