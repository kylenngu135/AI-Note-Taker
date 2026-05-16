'use client';

import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import type { Components } from 'react-markdown';

interface Props {
  content: string;
}

const components: Components = {
  h1: ({ children }) => (
    <h1
      style={{
        fontFamily: 'var(--font-dm-sans, "DM Sans"), sans-serif',
        fontSize: '0.6875rem',
        fontWeight: 600,
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
        color: 'var(--text-secondary)',
        marginTop: '1.75rem',
        marginBottom: '0.5rem',
      }}
    >
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2
      style={{
        fontFamily: 'var(--font-dm-sans, "DM Sans"), sans-serif',
        fontSize: '0.6875rem',
        fontWeight: 600,
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
        color: 'var(--text-secondary)',
        marginTop: '1.5rem',
        marginBottom: '0.5rem',
      }}
    >
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3
      style={{
        fontFamily: 'var(--font-dm-sans, "DM Sans"), sans-serif',
        fontSize: '0.6875rem',
        fontWeight: 600,
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
        color: 'var(--text-secondary)',
        marginTop: '1.25rem',
        marginBottom: '0.375rem',
      }}
    >
      {children}
    </h3>
  ),
  p: ({ children }) => (
    <p
      style={{
        fontFamily: 'var(--font-dm-serif-display, "DM Serif Display"), serif',
        fontSize: '1rem',
        lineHeight: '1.75',
        color: 'var(--text)',
        marginBottom: '1rem',
      }}
    >
      {children}
    </p>
  ),
  li: ({ children }) => (
    <li
      style={{
        fontFamily: 'var(--font-dm-serif-display, "DM Serif Display"), serif',
        fontSize: '1rem',
        lineHeight: '1.7',
        color: 'var(--text)',
        marginBottom: '0.25rem',
      }}
    >
      {children}
    </li>
  ),
  ul: ({ children }) => (
    <ul style={{ paddingLeft: '1.5rem', marginBottom: '1rem' }}>{children}</ul>
  ),
  ol: ({ children }) => (
    <ol style={{ paddingLeft: '1.5rem', marginBottom: '1rem' }}>{children}</ol>
  ),
  pre: ({ children }) => (
    <pre
      style={{
        background: 'var(--surface2)',
        border: '1px solid var(--border)',
        borderRadius: 'var(--radius-sm)',
        padding: '1rem',
        overflowX: 'auto',
        marginBottom: '1rem',
        fontFamily: 'var(--font-jetbrains-mono, "JetBrains Mono"), monospace',
        fontSize: '0.8125rem',
        lineHeight: 1.6,
      }}
    >
      {children}
    </pre>
  ),
  code: ({ children, className }) => {
    const isBlock = Boolean(className);
    return (
      <code
        style={{
          fontFamily: 'var(--font-jetbrains-mono, "JetBrains Mono"), monospace',
          fontSize: '0.8125rem',
          background: isBlock ? 'transparent' : 'var(--surface2)',
          border: isBlock ? 'none' : '1px solid var(--border)',
          padding: isBlock ? '0' : '0.125rem 0.3125rem',
          borderRadius: isBlock ? '0' : '3px',
        }}
      >
        {children}
      </code>
    );
  },
  blockquote: ({ children }) => (
    <blockquote
      style={{
        borderLeft: '3px solid var(--border-strong)',
        paddingLeft: '1rem',
        color: 'var(--text-secondary)',
        marginBottom: '1rem',
        fontFamily: 'var(--font-dm-serif-display, "DM Serif Display"), serif',
        fontStyle: 'italic',
      }}
    >
      {children}
    </blockquote>
  ),
  hr: () => (
    <hr
      style={{
        border: 'none',
        borderTop: '1px solid var(--border)',
        margin: '1.5rem 0',
      }}
    />
  ),
  strong: ({ children }) => (
    <strong style={{ fontWeight: 700 }}>{children}</strong>
  ),
  a: ({ children, href }) => (
    <a
      href={href}
      target="_blank"
      rel="noreferrer"
      style={{ color: 'var(--accent)', textDecoration: 'underline' }}
    >
      {children}
    </a>
  ),
};

export default function NoteContent({ content }: Props) {
  return (
    <div>
      <ReactMarkdown remarkPlugins={[remarkGfm]} components={components}>
        {content}
      </ReactMarkdown>
    </div>
  );
}
