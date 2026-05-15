"use client";

import { useRef, useState } from "react";

interface Props {
  onUpload: (file: File) => void;
}

export default function WelcomeView({ onUpload }: Props) {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0] ?? null;
    setSelectedFile(file);
  };

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(true);
  };

  const handleDragLeave = () => setIsDragging(false);

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    const file = e.dataTransfer.files?.[0] ?? null;
    setSelectedFile(file);
  };

  const handleUpload = () => {
    if (!selectedFile) {
      alert("Please select a file to upload.");
      return;
    }
    onUpload(selectedFile);
  };

  return (
    <div className="flex-1 overflow-y-auto flex">
      <div className="max-w-[640px] mx-auto px-6 pt-[60px] pb-6 w-full animate-fade-up">
        <h1 className="font-display text-[2.2rem] font-normal text-primary tracking-[-0.03em] mb-2">
          What are we studying today?
        </h1>
        <p className="text-secondary text-sm mb-8">
          Upload a document, video, or audio file to generate notes.
        </p>

        <div className="bg-panel border border-border-light rounded-[10px] p-6 flex flex-col gap-4">
          {/* Drop zone */}
          <div
            className="relative"
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
          >
            <input
              ref={fileInputRef}
              type="file"
              accept=".pdf,.docx,.txt,.mp4,.mov,.mp3,.wav,.m4a"
              onChange={handleFileChange}
              className="absolute inset-0 opacity-0 cursor-pointer w-full h-full"
            />
            <label
              className={[
                "flex flex-col items-center justify-center gap-1.5 px-6 py-7 border-[1.5px] border-dashed rounded-[6px] cursor-pointer transition-all text-center",
                isDragging
                  ? "border-accent bg-white/[0.04]"
                  : "border-border-light hover:border-accent hover:bg-white/[0.04]",
              ].join(" ")}
            >
              <span className="text-[1.6rem] text-accent">⊕</span>
              <span className="text-primary text-sm">
                {selectedFile ? selectedFile.name : "Drop a file or click to browse"}
              </span>
              <span className="text-muted text-[0.7rem] tracking-[0.05em]">
                PDF · DOCX · TXT · MP4 · MOV · MP3 · WAV · M4A
              </span>
            </label>
          </div>

          {/* Custom prompt (UI placeholder — not sent to backend) */}
          <div className="flex">
            <input
              type="text"
              placeholder="Custom prompt (optional) — leave blank to use default…"
              className="font-mono text-[0.8rem] bg-input border border-border-light rounded-[6px] text-primary px-3.5 py-2.5 w-full outline-none transition-colors focus:border-accent placeholder:text-muted"
            />
          </div>

          {/* Upload button */}
          <div className="flex gap-2 flex-wrap">
            <button
              onClick={handleUpload}
              className="font-mono text-[0.75rem] tracking-[0.03em] px-[18px] py-[9px] rounded-[6px] border border-accent bg-accent text-primary font-medium cursor-pointer transition-all hover:bg-accent-dim hover:border-accent-dim"
            >
              Upload
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
