export interface Tag {
  id: string;
  name: string;
  type: string;
  color: string;
}

export interface UploadListItem {
  id: string;
  filename: string;
  created_at: string;
  tags: Tag[];
}

export interface Note {
  id: string;
  upload_id: string;
  content: string;
  created_at: string;
  last_updated_at: string;
}

export interface NoteHistory {
  id: string;
  note_id: string;
  upload_id: string;
  role: 'user' | 'assistant';
  prompt: string;
  content: string;
  created_at: string;
}

export interface NoteWithHistory {
  note: Note;
  history: NoteHistory[];
  tags: Tag[];
}

export interface MeResponse {
  user_id: string;
  email: string;
}

export interface AuthResponse {
  token: string;
}

export interface UploadResponse {
  job_id: string;
  upload_id: string;
}

export type ExportFormat = 'txt' | 'pdf' | 'docx';

export type ToastType = 'success' | 'error' | 'info';

export interface ToastItem {
  id: string;
  message: string;
  type: ToastType;
}

export type FileCategory = 'document' | 'audio' | 'video';
