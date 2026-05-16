'use client';

interface SkeletonProps {
  width?: string | number;
  height?: number;
  className?: string;
}

export default function Skeleton({ width = '100%', height = 14, className = '' }: SkeletonProps) {
  return (
    <div
      className={`skeleton-bar ${className}`}
      style={{ width, height }}
      aria-hidden="true"
    />
  );
}
