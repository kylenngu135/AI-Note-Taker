import API_BASE_URL from "./config";
import type {
  Tag,
  UploadListItem,
  NoteWithHistory,
  Note,
  User,
  UploadResponse,
  AuthResponse,
} from "./types";

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function fetchJson<T>(url: string, options?: RequestInit): Promise<T> {
  const response = await fetch(url, {
    ...options,
    credentials: "include",
  });

  if (!response.ok) {
    const text = await response.text().catch(() => "Unknown error");
    throw new ApiError(response.status, text);
  }

  return response.json() as Promise<T>;
}

// ── Auth ──────────────────────────────────────────────────────────────────────

export async function getMe(): Promise<User> {
  return fetchJson<User>(`${API_BASE_URL}/api/auth/me`);
}

export async function register(
  email: string,
  password: string,
): Promise<AuthResponse> {
  return fetchJson<AuthResponse>(`${API_BASE_URL}/api/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
}

export async function login(
  email: string,
  password: string,
): Promise<AuthResponse> {
  return fetchJson<AuthResponse>(`${API_BASE_URL}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });
}

export async function logout(): Promise<void> {
  await fetch(`${API_BASE_URL}/api/auth/logout`, {
    method: "POST",
    credentials: "include",
  });
}

// ── Uploads ───────────────────────────────────────────────────────────────────

export async function getUploads(tagFilter?: string): Promise<UploadListItem[]> {
  const url = tagFilter
    ? `${API_BASE_URL}/api/uploads?tag=${encodeURIComponent(tagFilter)}`
    : `${API_BASE_URL}/api/uploads`;
  return fetchJson<UploadListItem[]>(url);
}

export async function uploadFile(file: File): Promise<UploadResponse> {
  const formData = new FormData();
  formData.append("file", file);

  const response = await fetch(`${API_BASE_URL}/api/uploads`, {
    method: "POST",
    body: formData,
    credentials: "include",
  });

  if (!response.ok) {
    const text = await response.text().catch(() => "Upload failed");
    throw new ApiError(response.status, text);
  }

  return response.json() as Promise<UploadResponse>;
}

export async function deleteUpload(id: string): Promise<void> {
  const response = await fetch(`${API_BASE_URL}/api/uploads/${id}`, {
    method: "DELETE",
    credentials: "include",
  });

  if (response.status !== 204) {
    throw new ApiError(response.status, "Failed to delete upload");
  }
}

export async function getNotesByUploadId(
  id: string,
): Promise<NoteWithHistory> {
  return fetchJson<NoteWithHistory>(`${API_BASE_URL}/api/uploads/${id}/notes`);
}

export async function exportNotes(
  id: string,
  format: "txt" | "pdf" | "docx",
): Promise<Blob> {
  const response = await fetch(
    `${API_BASE_URL}/api/uploads/${id}/notes?format=${format}`,
    { credentials: "include" },
  );

  if (!response.ok) {
    throw new ApiError(response.status, "Failed to export notes");
  }

  return response.blob();
}

export async function regenerateNote(
  id: string,
  prompt: string,
): Promise<Note> {
  return fetchJson<Note>(
    `${API_BASE_URL}/api/uploads/${id}/notes/regenerate`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ prompt }),
    },
  );
}

// ── Tags ──────────────────────────────────────────────────────────────────────

export async function addTagToUpload(
  uploadId: string,
  name: string,
  color: string,
): Promise<Tag> {
  return fetchJson<Tag>(`${API_BASE_URL}/api/uploads/${uploadId}/tags`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, color }),
  });
}

export async function removeTagFromUpload(
  uploadId: string,
  tagId: string,
): Promise<void> {
  const response = await fetch(
    `${API_BASE_URL}/api/uploads/${uploadId}/tags/${tagId}`,
    { method: "DELETE", credentials: "include" },
  );

  if (!response.ok) {
    throw new ApiError(response.status, "Failed to remove tag");
  }
}
