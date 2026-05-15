"use client";

import { useState } from "react";
import type { Tag } from "@/lib/types";

const TAG_COLORS = [
  "#ef4444",
  "#f97316",
  "#eab308",
  "#22c55e",
  "#3b82f6",
  "#8b5cf6",
  "#ec4899",
  "#6b7280",
];

interface Props {
  tags: Tag[];
  uploadId: string;
  onAddTag: (name: string, color: string) => void;
  onRemoveTag: (tagId: string) => void;
}

export default function NotesTagsBar({
  tags,
  uploadId,
  onAddTag,
  onRemoveTag,
}: Props) {
  const [showForm, setShowForm] = useState(false);
  const [inputValue, setInputValue] = useState("");
  const [selectedColor, setSelectedColor] = useState(TAG_COLORS[0]);

  if (!uploadId) return null;

  const handleSubmit = () => {
    const rawName = inputValue
      .trim()
      .toLowerCase()
      .replace(/\s+/g, "-")
      .replace(/[^a-z0-9-]/g, "");

    if (!rawName) return;

    onAddTag(rawName, selectedColor);
    setInputValue("");
    setShowForm(false);
    setSelectedColor(TAG_COLORS[0]);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      e.preventDefault();
      handleSubmit();
    }
    if (e.key === "Escape") {
      setInputValue("");
      setShowForm(false);
    }
  };

  const handleCancel = () => {
    setInputValue("");
    setShowForm(false);
  };

  return (
    <div className="flex-shrink-0 px-7 py-[7px] border-b border-border flex items-center flex-wrap gap-1.5 bg-app-bg min-h-9">
      {/* Existing tags */}
      <div className="flex flex-wrap gap-[5px] items-center">
        {tags.map((tag) => (
          <span
            key={tag.id}
            className="inline-flex items-center gap-[3px] text-[0.62rem] px-1.5 py-0.5 rounded-[8px] font-mono tracking-[0.02em] border whitespace-nowrap leading-[1.4] opacity-90 cursor-pointer hover:opacity-100 transition-opacity"
            style={{
              borderColor: tag.color || "#2a4f7a",
              color: tag.color || "#94a3b8",
            }}
          >
            #{tag.name}
            {tag.type === "user" && (
              <button
                onClick={() => onRemoveTag(tag.id)}
                className="bg-transparent border-none cursor-pointer text-[0.55rem] p-0 opacity-60 leading-none inline-flex items-center hover:opacity-100 transition-opacity"
                style={{ color: "inherit" }}
                aria-label="Remove tag"
              >
                ✕
              </button>
            )}
          </span>
        ))}
      </div>

      {/* Add tag */}
      {!showForm && (
        <button
          onClick={() => setShowForm(true)}
          className="text-[0.62rem] px-[7px] py-0.5 border border-dashed border-border-light rounded-[8px] bg-transparent text-muted cursor-pointer font-mono transition-all hover:border-accent hover:text-accent"
        >
          + tag
        </button>
      )}

      {showForm && (
        <div className="flex items-center gap-[5px]">
          <input
            autoFocus
            type="text"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="tag-name…"
            maxLength={30}
            autoComplete="off"
            className="font-mono text-[0.75rem] bg-input border border-border-light rounded-[6px] text-primary px-2 py-[3px] outline-none w-[120px] transition-colors focus:border-accent placeholder:text-muted"
          />
          {/* Color picker */}
          <div className="flex gap-1 items-center">
            {TAG_COLORS.map((color) => (
              <button
                key={color}
                type="button"
                aria-label={color}
                onClick={() => setSelectedColor(color)}
                className="w-[13px] h-[13px] rounded-full border-2 cursor-pointer p-0 flex-shrink-0 transition-transform hover:scale-125"
                style={{
                  background: color,
                  borderColor:
                    selectedColor === color ? "#f1f5f9" : "transparent",
                  transform:
                    selectedColor === color ? "scale(1.1)" : undefined,
                }}
              />
            ))}
          </div>
          <button
            onClick={handleSubmit}
            className="font-mono text-[0.68rem] px-[7px] py-[3px] border border-border-light rounded-[6px] bg-transparent text-secondary cursor-pointer transition-all hover:border-accent hover:text-accent"
          >
            Add
          </button>
          <button
            onClick={handleCancel}
            className="font-mono text-[0.68rem] px-[7px] py-[3px] border border-border-light rounded-[6px] bg-transparent text-secondary cursor-pointer transition-all hover:border-danger hover:text-danger"
          >
            ✕
          </button>
        </div>
      )}
    </div>
  );
}
