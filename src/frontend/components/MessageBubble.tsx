"use client";

import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";

interface Props {
  role: "user" | "assistant";
  content: string;
  label?: string;
  isLoading?: boolean;
}

export default function MessageBubble({
  role,
  content,
  label,
  isLoading = false,
}: Props) {
  const isUser = role === "user";
  const displayLabel = label ?? (isUser ? "You" : "AI Study Assistant");

  return (
    <div
      className={[
        "rounded-[10px] border p-5 text-sm leading-[1.8] text-primary animate-fade-up",
        isUser
          ? "bg-hover border-border-light self-end max-w-[80%]"
          : "bg-panel border-border",
      ].join(" ")}
    >
      <div className="text-[0.68rem] tracking-[0.1em] uppercase text-accent mb-2.5">
        {displayLabel}
      </div>
      {isLoading ? (
        <span className="text-secondary">Generating…</span>
      ) : isUser ? (
        <span>{content}</span>
      ) : (
        <div className="message-prose">
          <ReactMarkdown remarkPlugins={[remarkGfm]}>{content}</ReactMarkdown>
        </div>
      )}
    </div>
  );
}
