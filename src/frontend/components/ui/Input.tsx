'use client';

import { forwardRef, type InputHTMLAttributes } from 'react';

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

const Input = forwardRef<HTMLInputElement, InputProps>(function Input(
  { label, error, id, style, ...props },
  ref
) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
      {label && (
        <label
          htmlFor={id}
          style={{
            fontSize: '0.8125rem',
            fontWeight: 500,
            color: 'var(--text)',
          }}
        >
          {label}
        </label>
      )}
      <input
        ref={ref}
        id={id}
        style={{
          height: '36px',
          padding: '0 10px',
          border: '0.5px solid var(--border-strong)',
          borderRadius: 'var(--radius-sm)',
          background: 'var(--surface)',
          color: 'var(--text)',
          fontSize: '0.875rem',
          outline: 'none',
          transition: 'border-color 0.15s',
          width: '100%',
          ...style,
        }}
        onFocus={(e) => {
          e.currentTarget.style.borderColor = 'var(--accent)';
        }}
        onBlur={(e) => {
          e.currentTarget.style.borderColor = 'var(--border-strong)';
        }}
        {...props}
      />
      {error && (
        <span style={{ fontSize: '0.75rem', color: 'var(--danger)' }}>
          {error}
        </span>
      )}
    </div>
  );
});

export default Input;
