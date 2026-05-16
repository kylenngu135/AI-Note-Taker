'use client';

import type { ButtonHTMLAttributes, ReactNode } from 'react';

type Variant = 'primary' | 'secondary' | 'danger' | 'ghost';

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
  children: ReactNode;
}

const styles: Record<Variant, React.CSSProperties> = {
  primary: {
    background: 'var(--accent)',
    color: '#fff',
    border: 'none',
  },
  secondary: {
    background: 'transparent',
    color: 'var(--text)',
    border: '0.5px solid var(--border-strong)',
  },
  danger: {
    background: 'var(--danger)',
    color: '#fff',
    border: 'none',
  },
  ghost: {
    background: 'transparent',
    color: 'var(--text-secondary)',
    border: '0.5px solid var(--border)',
  },
};

export default function Button({
  variant = 'primary',
  children,
  style,
  disabled,
  ...props
}: ButtonProps) {
  const base: React.CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: '6px',
    height: '32px',
    padding: '0 12px',
    borderRadius: 'var(--radius-sm)',
    fontSize: '0.8125rem',
    fontWeight: 500,
    lineHeight: 1,
    cursor: disabled ? 'not-allowed' : 'pointer',
    opacity: disabled ? 0.5 : 1,
    transition: 'background 0.1s, opacity 0.1s',
    whiteSpace: 'nowrap',
    ...styles[variant],
    ...style,
  };

  return (
    <button style={base} disabled={disabled} {...props}>
      {children}
    </button>
  );
}
