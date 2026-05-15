"use client";

import { useEffect, useRef } from "react";
import MessageBubble from "./MessageBubble";
import type { MessageBubbleItem } from "@/lib/types";

interface Props {
  messages: MessageBubbleItem[];
}

export default function NotesView({ messages }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [messages]);

  return (
    <div
      ref={containerRef}
      className="flex-1 overflow-y-auto scrollbar-notes"
    >
      <div className="max-w-[720px] mx-auto px-6 py-8 flex flex-col gap-6">
        {messages.map((msg) => (
          <MessageBubble
            key={msg.id}
            role={msg.role}
            content={msg.content}
            label={msg.label}
            isLoading={msg.isLoading}
          />
        ))}
      </div>
    </div>
  );
}
