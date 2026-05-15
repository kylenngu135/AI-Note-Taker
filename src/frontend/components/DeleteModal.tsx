"use client";

interface Props {
  onClose: () => void;
  onConfirm: () => void;
}

export default function DeleteModal({ onClose, onConfirm }: Props) {
  const handleOverlayClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (e.target === e.currentTarget) onClose();
  };

  return (
    <div
      className="fixed inset-0 bg-black/60 flex items-center justify-center z-[1000]"
      onClick={handleOverlayClick}
    >
      <div className="bg-panel border border-border-light rounded-[10px] p-6 min-w-[280px] animate-fade-up">
        <div className="flex items-center justify-between mb-5">
          <span className="font-display text-[1.1rem] text-primary">
            Delete Document?
          </span>
          <button
            onClick={onClose}
            className="bg-transparent border-none text-muted text-base cursor-pointer p-1 hover:text-primary transition-colors"
          >
            ✕
          </button>
        </div>
        <p className="text-[0.85rem] text-secondary mb-5 leading-relaxed">
          This will delete the document and all its conversation history. This
          action cannot be undone.
        </p>
        <div className="flex gap-2.5">
          <button
            onClick={onConfirm}
            className="flex-1 font-mono text-[0.8rem] py-3 px-4 border border-danger rounded-[6px] bg-transparent text-danger cursor-pointer transition-all hover:bg-danger hover:text-primary"
          >
            Yes, Delete
          </button>
          <button
            onClick={onClose}
            className="flex-1 font-mono text-[0.8rem] py-3 px-4 border border-border-light rounded-[6px] bg-transparent text-secondary cursor-pointer transition-all hover:border-accent hover:text-accent"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}
