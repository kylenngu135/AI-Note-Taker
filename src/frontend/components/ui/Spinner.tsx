'use client';

interface SpinnerProps {
  size?: number;
  color?: string;
}

export default function Spinner({ size = 16, color = 'var(--accent)' }: SpinnerProps) {
  return (
    <svg
      className="spin"
      width={size}
      height={size}
      viewBox="0 0 16 16"
      fill="none"
      aria-hidden="true"
    >
      <circle
        cx="8"
        cy="8"
        r="6"
        stroke={color}
        strokeOpacity="0.25"
        strokeWidth="2"
      />
      <path
        d="M8 2a6 6 0 0 1 6 6"
        stroke={color}
        strokeWidth="2"
        strokeLinecap="round"
      />
    </svg>
  );
}
