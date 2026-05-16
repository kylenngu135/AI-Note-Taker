'use client';

interface IconProps {
  size?: number;
  color?: string;
  className?: string;
}

export function DocumentIcon({ size = 16, color = 'currentColor' }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" aria-hidden="true">
      <path
        d="M4 2h6l3 3v9a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1V3a1 1 0 0 1 1-1z"
        stroke={color}
        strokeWidth="1.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M10 2v3h3"
        stroke={color}
        strokeWidth="1.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M6 8h4M6 11h4"
        stroke={color}
        strokeWidth="1.2"
        strokeLinecap="round"
      />
    </svg>
  );
}

export function AudioIcon({ size = 16, color = 'currentColor' }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" aria-hidden="true">
      <path
        d="M10 3 6 6H3a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h3l4 3V3z"
        stroke={color}
        strokeWidth="1.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M12.5 5.5a3 3 0 0 1 0 5"
        stroke={color}
        strokeWidth="1.2"
        strokeLinecap="round"
      />
    </svg>
  );
}

export function VideoIcon({ size = 16, color = 'currentColor' }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" aria-hidden="true">
      <rect
        x="1"
        y="3"
        width="10"
        height="10"
        rx="1"
        stroke={color}
        strokeWidth="1.2"
      />
      <path
        d="M11 6l4-2v8l-4-2V6z"
        stroke={color}
        strokeWidth="1.2"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function UploadCloudIcon({ size = 24, color = 'currentColor' }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path
        d="M4 16.5A4.5 4.5 0 0 1 7.5 8a6 6 0 0 1 11.8 1.5A3.5 3.5 0 0 1 20.5 16H17"
        stroke={color}
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <path
        d="M12 12v9M9 18l3 3 3-3"
        stroke={color}
        strokeWidth="1.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  );
}

export function NotesIcon({ size = 20, color = 'currentColor' }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 20 20" fill="none" aria-hidden="true">
      <rect x="3" y="2" width="14" height="16" rx="2" stroke={color} strokeWidth="1.5" />
      <path d="M7 7h6M7 10h6M7 13h4" stroke={color} strokeWidth="1.5" strokeLinecap="round" />
    </svg>
  );
}

export function ChevronDownIcon({ size = 14, color = 'currentColor' }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 14 14" fill="none" aria-hidden="true">
      <path d="M3 5l4 4 4-4" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export function SendIcon({ size = 16, color = 'currentColor' }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" aria-hidden="true">
      <path d="M14 2 7 9M14 2 9 14l-2-5-5-2 12-5z" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export function TrashIcon({ size = 14, color = 'currentColor' }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 14 14" fill="none" aria-hidden="true">
      <path d="M2 3.5h10M5.5 3.5V2.5a.5.5 0 0 1 .5-.5h2a.5.5 0 0 1 .5.5v1M11.5 3.5 11 11a1 1 0 0 1-1 1H4a1 1 0 0 1-1-1l-.5-7.5" stroke={color} strokeWidth="1.2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}

export function UploadFolderIcon({ size = 40, color = 'currentColor' }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 40 40" fill="none" aria-hidden="true">
      <path d="M8 12a2 2 0 0 1 2-2h6l2 2h12a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H10a2 2 0 0 1-2-2V12z" stroke={color} strokeWidth="1.5" />
      <path d="M20 17v8M17 22l3 3 3-3" stroke={color} strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
