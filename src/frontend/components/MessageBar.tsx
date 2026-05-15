"use client";

import { useRef } from "react";

interface Props {
  onSend: (prompt: string) => void;
}

export default function MessageBar({ onSend }: Props) {
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  const handleSend = () => {
    const prompt = textareaRef.current?.value.trim() ?? "";
    if (!prompt) return;
    if (textareaRef.current) textareaRef.current.value = "";
    resetHeight();
    onSend(prompt);
  };

  const resetHeight = () => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
    }
  };

  const handleInput = () => {
    if (textareaRef.current) {
      textareaRef.current.style.height = "auto";
      textareaRef.current.style.height = `${textareaRef.current.scrollHeight}px`;
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="flex-shrink-0 px-6 pt-3 pb-4 border-t border-border bg-app-bg">
      <div className="max-w-[720px] mx-auto flex items-end gap-2.5 bg-input border border-border-light rounded-[10px] px-3 py-2.5 focus-within:border-accent transition-colors">
        <textarea
          ref={textareaRef}
          rows={1}
          placeholder="Ask a follow-up question or request changes…"
          onInput={handleInput}
          onKeyDown={handleKeyDown}
          className="flex-1 font-mono text-sm bg-transparent border-none outline-none text-primary resize-none leading-[1.5] max-h-40 overflow-y-auto placeholder:text-muted"
        />
        <button
          onClick={handleSend}
          className="w-8 h-8 rounded-[6px] bg-accent text-primary border-none cursor-pointer text-base flex items-center justify-center flex-shrink-0 transition-colors hover:bg-accent-dim"
        >
          ↑
        </button>
      </div>
      <div className="max-w-[720px] mx-auto mt-2 text-[0.68rem] text-muted text-center">
        AI Notes may make mistakes. Always verify important information.
      </div>
    </div>
  );
}
