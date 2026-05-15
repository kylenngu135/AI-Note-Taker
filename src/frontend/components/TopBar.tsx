"use client";

interface Props {
  title: string;
  showActions: boolean;
  onExport: () => void;
  onDelete: () => void;
}

export default function TopBar({
  title,
  showActions,
  onExport,
  onDelete,
}: Props) {
  return (
    <div className="h-14 flex items-center justify-between px-7 border-b border-border flex-shrink-0">
      <div className="font-display text-[1.1rem] text-secondary tracking-[-0.01em]">
        {title}
      </div>
      {showActions && (
        <div className="flex gap-2">
          <button
            onClick={onExport}
            className="font-mono text-[0.75rem] px-3.5 py-[7px] rounded-[6px] border border-border-light bg-transparent text-secondary cursor-pointer transition-all flex items-center gap-1.5 hover:border-accent hover:text-accent"
          >
            ↑ Export
          </button>
          <button
            onClick={onDelete}
            className="font-mono text-[0.75rem] px-3.5 py-[7px] rounded-[6px] border border-border-light bg-transparent text-secondary cursor-pointer transition-all flex items-center gap-1.5 hover:border-danger hover:text-danger"
          >
            ✕ Delete
          </button>
        </div>
      )}
    </div>
  );
}
