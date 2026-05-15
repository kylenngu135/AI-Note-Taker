"use client";

import Link from "next/link";
import { useAuth } from "@/context/AuthContext";
import type { UploadListItem } from "@/lib/types";

interface Props {
  uploads: UploadListItem[];
  currentUploadId: string | null;
  activeTagFilter: string | null;
  onUploadClick: (id: string, name: string) => void;
  onNewUpload: () => void;
  onTagFilter: (tagName: string) => void;
  onTagFilterClear: () => void;
}

export default function Sidebar({
  uploads,
  currentUploadId,
  activeTagFilter,
  onUploadClick,
  onNewUpload,
  onTagFilter,
  onTagFilterClear,
}: Props) {
  const { user, logout } = useAuth();

  const handleSignOut = async () => {
    await logout();
    window.location.reload();
  };

  return (
    <aside className="w-[260px] flex-shrink-0 bg-sidebar border-r border-border flex flex-col h-screen overflow-hidden">
      {/* Top section */}
      <div className="px-3.5 pt-5 pb-3 flex flex-col gap-2.5 border-b border-border">
        <div className="font-display text-[1.2rem] text-accent tracking-[-0.02em] px-1">
          ⌁ AI Notes
        </div>
        <button
          onClick={onNewUpload}
          className="flex items-center gap-2 px-3 py-[9px] bg-accent text-primary rounded-[6px] cursor-pointer border-none font-mono text-[0.8rem] transition-colors hover:bg-accent-dim"
        >
          <span className="text-base leading-none">＋</span>
          <span>New Upload</span>
        </button>
      </div>

      {/* Recents */}
      <div className="flex-1 overflow-y-auto px-2.5 py-4 flex flex-col gap-1 scrollbar-sidebar">
        <div className="text-[0.68rem] tracking-[0.15em] uppercase text-muted px-1.5 mb-1.5">
          Recents
        </div>

        {activeTagFilter && (
          <div className="flex items-center justify-between px-2.5 py-1 mb-1 bg-accent/[0.08] border border-accent/25 rounded-[6px] text-[0.65rem] text-accent">
            <span>Filtering: #{activeTagFilter}</span>
            <button
              onClick={onTagFilterClear}
              className="bg-transparent border-none text-accent cursor-pointer text-[0.65rem] p-0 opacity-70 hover:opacity-100 transition-opacity"
            >
              ✕
            </button>
          </div>
        )}

        {uploads.length === 0 ? (
          <div className="text-[0.78rem] text-muted px-1.5">
            {activeTagFilter
              ? `No uploads tagged #${activeTagFilter}.`
              : "No uploads yet."}
          </div>
        ) : (
          uploads.map((upload) => {
            const ext = upload.filename.split(".").pop()?.toLowerCase() ?? "";
            const date = new Date(upload.created_at).toLocaleDateString(
              "en-US",
              { month: "short", day: "numeric", year: "numeric" },
            );
            const isActive = upload.id === currentUploadId;

            return (
              <div
                key={upload.id}
                onClick={() => onUploadClick(upload.id, upload.filename)}
                className={[
                  "flex flex-col gap-1 px-2.5 py-2 rounded-[6px] cursor-pointer transition-colors border text-secondary text-[0.8rem]",
                  isActive
                    ? "bg-hover border-border-light text-primary"
                    : "border-transparent hover:bg-hover hover:text-primary",
                ].join(" ")}
              >
                <div className="flex items-center gap-2 whitespace-nowrap overflow-hidden">
                  <span className="text-[0.65rem] text-muted flex-shrink-0">
                    {date}
                  </span>
                  <span className="flex-1 overflow-hidden text-ellipsis">
                    {upload.filename}
                  </span>
                  <span className="text-[0.6rem] px-[5px] py-0.5 rounded-[3px] bg-hover border border-border text-muted flex-shrink-0 uppercase">
                    {ext}
                  </span>
                </div>

                {upload.tags && upload.tags.length > 0 && (
                  <div className="flex flex-wrap gap-[3px] pl-0.5">
                    {upload.tags.map((tag) => (
                      <span
                        key={tag.id}
                        onClick={(e) => {
                          e.stopPropagation();
                          onTagFilter(tag.name);
                        }}
                        className="inline-flex items-center text-[0.62rem] px-1.5 py-0.5 rounded-[8px] font-mono tracking-[0.02em] border whitespace-nowrap leading-[1.4] opacity-65 cursor-pointer hover:opacity-100 transition-opacity"
                        style={{
                          borderColor: tag.color || "#2a4f7a",
                          color: tag.color || "#94a3b8",
                        }}
                      >
                        {tag.name}
                      </span>
                    ))}
                  </div>
                )}
              </div>
            );
          })
        )}
      </div>

      {/* Bottom: Sign in/out */}
      <div className="px-2.5 py-3 border-t border-border">
        {user ? (
          <button
            onClick={handleSignOut}
            className="flex items-center gap-2.5 px-3 py-2.5 bg-transparent border border-border-light rounded-[6px] cursor-pointer font-mono text-[0.8rem] text-secondary w-full transition-all hover:bg-hover hover:text-primary"
          >
            <span className="text-base">◎</span>
            <span>Sign Out</span>
          </button>
        ) : (
          <Link
            href="/auth"
            className="flex items-center gap-2.5 px-3 py-2.5 bg-transparent border border-border-light rounded-[6px] font-mono text-[0.8rem] text-secondary w-full transition-all hover:bg-hover hover:text-primary no-underline"
          >
            <span className="text-base">◎</span>
            <span>Sign In</span>
          </Link>
        )}
      </div>
    </aside>
  );
}
