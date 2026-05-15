"use client";

interface Props {
  onClose: () => void;
  onExport: (format: "txt" | "pdf" | "docx") => void;
}

const formats: { label: string; value: "txt" | "pdf" | "docx" }[] = [
  { label: ".txt", value: "txt" },
  { label: ".pdf", value: "pdf" },
  { label: ".docx", value: "docx" },
];

export default function ExportModal({ onClose, onExport }: Props) {
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
          <span className="font-display text-[1.1rem] text-primary">Export As</span>
          <button
            onClick={onClose}
            className="bg-transparent border-none text-muted text-base cursor-pointer p-1 hover:text-primary transition-colors"
          >
            ✕
          </button>
        </div>
        <div className="flex gap-2.5">
          {formats.map(({ label, value }) => (
            <button
              key={value}
              onClick={() => onExport(value)}
              className="flex-1 font-mono text-[0.8rem] py-3 px-4 border border-border-light rounded-[6px] bg-transparent text-secondary cursor-pointer transition-all hover:border-accent hover:text-accent"
            >
              {label}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
