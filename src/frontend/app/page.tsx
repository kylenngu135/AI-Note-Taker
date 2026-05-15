"use client";

import { useCallback, useEffect, useState } from "react";
import Sidebar from "@/components/Sidebar";
import TopBar from "@/components/TopBar";
import NotesTagsBar from "@/components/NotesTagsBar";
import WelcomeView from "@/components/WelcomeView";
import NotesView from "@/components/NotesView";
import MessageBar from "@/components/MessageBar";
import ProcessingOverlay from "@/components/ProcessingOverlay";
import ExportModal from "@/components/ExportModal";
import DeleteModal from "@/components/DeleteModal";
import {
  getUploads,
  uploadFile,
  deleteUpload,
  getNotesByUploadId,
  exportNotes,
  regenerateNote,
  addTagToUpload,
  removeTagFromUpload,
} from "@/lib/api";
import type {
  UploadListItem,
  Tag,
  MessageBubbleItem,
  NoteWithHistory,
} from "@/lib/types";

type AppView = "welcome" | "notes" | "processing";

let msgId = 0;
function nextId() {
  return String(++msgId);
}

export default function Home() {
  const [view, setView] = useState<AppView>("welcome");
  const [currentUpload, setCurrentUpload] = useState<UploadListItem | null>(
    null,
  );
  const [messages, setMessages] = useState<MessageBubbleItem[]>([]);
  const [currentTags, setCurrentTags] = useState<Tag[]>([]);
  const [uploads, setUploads] = useState<UploadListItem[]>([]);
  const [activeTagFilter, setActiveTagFilter] = useState<string | null>(null);
  const [showExportModal, setShowExportModal] = useState(false);
  const [showDeleteModal, setShowDeleteModal] = useState(false);

  // ── Load uploads list ──────────────────────────────────────────────────────

  const loadUploads = useCallback(
    async (tagFilter?: string | null): Promise<UploadListItem[]> => {
      try {
        const list = await getUploads(tagFilter ?? undefined);
        setUploads(list ?? []);
        return list ?? [];
      } catch {
        setUploads([]);
        return [];
      }
    },
    [],
  );

  useEffect(() => {
    loadUploads(activeTagFilter);
  }, [activeTagFilter, loadUploads]);

  // ── Build messages from note history ──────────────────────────────────────

  function buildMessages(
    data: NoteWithHistory,
    filename: string,
  ): MessageBubbleItem[] {
    const result: MessageBubbleItem[] = [];
    const baseName = filename.replace(/\.[^.]+$/, "");

    if (data.history && data.history.length > 0) {
      let firstUserRendered = false;
      for (const item of data.history) {
        if (item.role === "user" && item.prompt) {
          if (!firstUserRendered) {
            result.push({
              id: nextId(),
              role: "user",
              content: `${baseName}_transcription.txt`,
              label: "You",
            });
            firstUserRendered = true;
          } else {
            result.push({
              id: nextId(),
              role: "user",
              content: item.prompt,
              label: "You",
            });
          }
        } else if (item.role === "assistant" && item.content) {
          result.push({
            id: nextId(),
            role: "assistant",
            content: item.content,
            label: "AI Study Assistant",
          });
        }
      }
    } else if (data.note?.content) {
      result.push({
        id: nextId(),
        role: "assistant",
        content: data.note.content,
        label: "AI Study Assistant",
      });
    } else {
      result.push({
        id: nextId(),
        role: "assistant",
        content: "No study sheet found.",
        label: "AI Study Assistant",
      });
    }

    return result;
  }

  // ── Load notes for a specific upload ─────────────────────────────────────

  const loadNotesById = useCallback(
    async (id: string, name: string) => {
      setView("notes");
      setCurrentUpload({ id, filename: name, created_at: "", tags: [] });
      setCurrentTags([]);
      setMessages([
        {
          id: nextId(),
          role: "assistant",
          content: "Loading…",
          label: "Loading…",
          isLoading: true,
        },
      ]);

      try {
        const data = await getNotesByUploadId(id);
        setCurrentTags(data.tags ?? []);
        setMessages(buildMessages(data, name));
        await loadUploads(activeTagFilter);
      } catch {
        setMessages([
          {
            id: nextId(),
            role: "assistant",
            content: "Failed to load study sheet.",
            label: "AI Study Assistant",
          },
        ]);
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [activeTagFilter, loadUploads],
  );

  // ── Upload a file ─────────────────────────────────────────────────────────

  const handleUpload = useCallback(
    async (file: File) => {
      setView("processing");

      try {
        const { upload_id } = await uploadFile(file);

        // Poll for the note (backend processes uploads asynchronously)
        const deadline = Date.now() + 120_000;
        while (Date.now() < deadline) {
          await new Promise((r) => setTimeout(r, 2000));
          try {
            const data = await getNotesByUploadId(upload_id);
            if (data?.note?.content) {
              await loadUploads(activeTagFilter);
              setCurrentUpload({
                id: upload_id,
                filename: file.name,
                created_at: new Date().toISOString(),
                tags: [],
              });
              setCurrentTags(data.tags ?? []);
              setMessages(buildMessages(data, file.name));
              setView("notes");
              return;
            }
          } catch {
            // Note not ready yet — keep polling
          }
        }

        // Timed out — show whatever is available
        await loadUploads(activeTagFilter);
        await loadNotesById(upload_id, file.name);
      } catch (err) {
        alert(
          "Upload failed: " +
            (err instanceof Error ? err.message : "Unknown error"),
        );
        setView("welcome");
      }
    },
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [activeTagFilter, loadUploads, loadNotesById],
  );

  // ── Delete current upload ─────────────────────────────────────────────────

  const handleDelete = useCallback(async () => {
    if (!currentUpload) return;
    try {
      await deleteUpload(currentUpload.id);
      window.location.reload();
    } catch {
      alert("Failed to delete document");
    }
  }, [currentUpload]);

  // ── Export notes ──────────────────────────────────────────────────────────

  const handleExport = useCallback(
    async (format: "txt" | "pdf" | "docx") => {
      if (!currentUpload) return;
      try {
        const blob = await exportNotes(currentUpload.id, format);
        const url = URL.createObjectURL(blob);
        const a = document.createElement("a");
        a.href = url;
        a.download = `notes-${currentUpload.filename}.${format}`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
      } catch (err) {
        alert(
          "Failed to export notes: " +
            (err instanceof Error ? err.message : "Unknown error"),
        );
      }
    },
    [currentUpload],
  );

  // ── Send follow-up prompt ─────────────────────────────────────────────────

  const handleSend = useCallback(
    async (prompt: string) => {
      if (!currentUpload) return;

      const loadingId = nextId();
      setMessages((prev) => [
        ...prev,
        { id: nextId(), role: "user", content: prompt, label: "You" },
        {
          id: loadingId,
          role: "assistant",
          content: "",
          label: "Generating…",
          isLoading: true,
        },
      ]);

      try {
        const note = await regenerateNote(currentUpload.id, prompt);
        setMessages((prev) =>
          prev
            .filter((m) => m.id !== loadingId)
            .concat({
              id: nextId(),
              role: "assistant",
              content: note.content,
              label: "AI Study Assistant",
            }),
        );
      } catch (err) {
        setMessages((prev) =>
          prev
            .filter((m) => m.id !== loadingId)
            .concat({
              id: nextId(),
              role: "assistant",
              content:
                "Error: " +
                (err instanceof Error ? err.message : "Failed to regenerate"),
              label: "Error",
            }),
        );
      }
    },
    [currentUpload],
  );

  // ── Tag management ────────────────────────────────────────────────────────

  const handleAddTag = useCallback(
    async (name: string, color: string) => {
      if (!currentUpload) return;
      try {
        const tag = await addTagToUpload(currentUpload.id, name, color);
        setCurrentTags((prev) => {
          const without = prev.filter((t) => t.id !== tag.id);
          return [...without, tag];
        });
        setUploads((prev) =>
          prev.map((u) =>
            u.id === currentUpload.id
              ? {
                  ...u,
                  tags: [
                    ...(u.tags ?? []).filter((t) => t.id !== tag.id),
                    tag,
                  ],
                }
              : u,
          ),
        );
      } catch (err) {
        console.error("Failed to add tag:", err);
      }
    },
    [currentUpload],
  );

  const handleRemoveTag = useCallback(
    async (tagId: string) => {
      if (!currentUpload) return;
      try {
        await removeTagFromUpload(currentUpload.id, tagId);
        setCurrentTags((prev) => prev.filter((t) => t.id !== tagId));
        setUploads((prev) =>
          prev.map((u) =>
            u.id === currentUpload.id
              ? {
                  ...u,
                  tags: (u.tags ?? []).filter((t) => t.id !== tagId),
                }
              : u,
          ),
        );
      } catch (err) {
        console.error("Failed to remove tag:", err);
      }
    },
    [currentUpload],
  );

  // ── Tag filter ────────────────────────────────────────────────────────────

  const handleTagFilter = useCallback((tagName: string) => {
    setActiveTagFilter((prev) => (prev === tagName ? null : tagName));
  }, []);

  const handleTagFilterClear = useCallback(() => {
    setActiveTagFilter(null);
  }, []);

  // ── Welcome view ──────────────────────────────────────────────────────────

  const showWelcomeView = useCallback(() => {
    setView("welcome");
    setCurrentUpload(null);
    setCurrentTags([]);
    setMessages([]);
  }, []);

  // ─────────────────────────────────────────────────────────────────────────

  const topbarTitle =
    view === "notes" ? (currentUpload?.filename ?? "AI Notes") : "AI Notes";

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-app-bg text-primary font-mono text-sm">
      <div className="grain" />

      <Sidebar
        uploads={uploads}
        currentUploadId={currentUpload?.id ?? null}
        activeTagFilter={activeTagFilter}
        onUploadClick={loadNotesById}
        onNewUpload={showWelcomeView}
        onTagFilter={handleTagFilter}
        onTagFilterClear={handleTagFilterClear}
      />

      <main className="flex flex-col flex-1 h-screen overflow-hidden bg-app-bg">
        <TopBar
          title={topbarTitle}
          showActions={view === "notes"}
          onExport={() => setShowExportModal(true)}
          onDelete={() => setShowDeleteModal(true)}
        />

        {view === "notes" && (
          <NotesTagsBar
            tags={currentTags}
            uploadId={currentUpload?.id ?? ""}
            onAddTag={handleAddTag}
            onRemoveTag={handleRemoveTag}
          />
        )}

        {view === "welcome" && <WelcomeView onUpload={handleUpload} />}
        {view === "notes" && <NotesView messages={messages} />}
        {view === "notes" && <MessageBar onSend={handleSend} />}
      </main>

      {view === "processing" && <ProcessingOverlay />}

      {showExportModal && (
        <ExportModal
          onClose={() => setShowExportModal(false)}
          onExport={(fmt) => {
            handleExport(fmt);
            setShowExportModal(false);
          }}
        />
      )}

      {showDeleteModal && (
        <DeleteModal
          onClose={() => setShowDeleteModal(false)}
          onConfirm={() => {
            setShowDeleteModal(false);
            handleDelete();
          }}
        />
      )}
    </div>
  );
}
