export interface Tag {
  id: string;
  user_id?: string;
  name: string;
  type: string;
  color: string;
  created_at: string;
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
  role: string;
  prompt: string;
  content: string;
  created_at: string;
}

export interface NoteWithHistory {
  note: Note;
  history: NoteHistory[];
  tags: Tag[];
}

export interface User {
  user_id: string;
  email: string;
}

export interface UploadResponse {
  job_id: string;
  upload_id: string;
}

export interface AuthResponse {
  token: string;
}

export interface MessageBubbleItem {
  id: string;
  role: "user" | "assistant";
  content: string;
  label?: string;
  isLoading?: boolean;
}
