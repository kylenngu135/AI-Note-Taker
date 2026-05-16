import type {
  AuthResponse,
  MeResponse,
  Note,
  NoteWithHistory,
  UploadListItem,
  UploadResponse,
} from './types';

const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

function redirectToLogin(): never {
  if (typeof window !== 'undefined') {
    window.location.href = '/login';
  }
  throw new Error('Unauthorized');
}

async function apiFetch<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    credentials: 'include',
    ...options,
  });

  if (res.status === 401) {
    redirectToLogin();
  }

  if (!res.ok) {
    const text = await res.text().catch(() => `HTTP ${res.status}`);
    throw new Error(text || `HTTP ${res.status}`);
  }

  return res.json() as Promise<T>;
}

// ── Auth ────────────────────────────────────────────────────────────────────

export async function getMe(): Promise<MeResponse> {
  return apiFetch<MeResponse>('/api/auth/me');
}

export async function login(
  email: string,
  password: string
): Promise<AuthResponse> {
  return apiFetch<AuthResponse>('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
}

export async function register(
  email: string,
  password: string
): Promise<AuthResponse> {
  return apiFetch<AuthResponse>('/api/auth/register', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  });
}

export async function logout(): Promise<void> {
  await fetch(`${API_BASE}/api/auth/logout`, {
    method: 'POST',
    credentials: 'include',
  });
}

// ── Uploads ──────────────────────────────────────────────────────────────────

export async function getUploads(): Promise<UploadListItem[]> {
  return apiFetch<UploadListItem[]>('/api/uploads');
}

export async function uploadFile(file: File): Promise<UploadResponse> {
  const formData = new FormData();
  formData.append('file', file);

  const res = await fetch(`${API_BASE}/api/uploads`, {
    method: 'POST',
    credentials: 'include',
    body: formData,
  });

  if (res.status === 401) {
    redirectToLogin();
  }

  if (!res.ok) {
    const text = await res.text().catch(() => `HTTP ${res.status}`);
    throw new Error(text || `HTTP ${res.status}`);
  }

  return res.json() as Promise<UploadResponse>;
}

export async function deleteUpload(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/api/uploads/${id}`, {
    method: 'DELETE',
    credentials: 'include',
  });

  if (res.status === 401) {
    redirectToLogin();
  }

  if (!res.ok) {
    const text = await res.text().catch(() => `HTTP ${res.status}`);
    throw new Error(text || `HTTP ${res.status}`);
  }
}

// ── Notes ────────────────────────────────────────────────────────────────────

export async function getNotes(uploadId: string): Promise<NoteWithHistory> {
  return apiFetch<NoteWithHistory>(`/api/uploads/${uploadId}/notes`);
}

export async function regenerateNote(
  uploadId: string,
  prompt: string
): Promise<Note> {
  return apiFetch<Note>(`/api/uploads/${uploadId}/notes/regenerate`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ prompt }),
  });
}

export function getExportUrl(
  uploadId: string,
  format: 'txt' | 'pdf' | 'docx'
): string {
  return `${API_BASE}/api/uploads/${uploadId}/notes?format=${format}`;
}
