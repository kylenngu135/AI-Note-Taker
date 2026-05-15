"use client";

export default function ProcessingOverlay() {
  return (
    <div className="fixed inset-0 bg-black/65 flex items-center justify-center z-[1000]">
      <div className="flex flex-col items-center justify-center gap-6 bg-[#1a3456] border border-[#2a4f7a] rounded-[10px] px-16 py-12 animate-fade-up shadow-[0_8px_32px_rgba(0,0,0,0.4)]">
        <div className="w-10 h-10 rounded-full border-[3px] border-accent/20 border-t-accent animate-spin-custom" />
        <p className="text-secondary text-sm">
          Transcribing and generating your study sheet…
        </p>
      </div>
    </div>
  );
}
