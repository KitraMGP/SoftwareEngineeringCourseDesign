export type UserRole = 'user' | 'admin';
export type UserStatus = 'active' | 'frozen';
export type MessageStatus = 'completed' | 'failed';
export type DocumentStatus = 'pending' | 'processing' | 'available' | 'failed' | 'deleting';

export interface User {
  id: string;
  username: string;
  email: string;
  nickname?: string | null;
  avatar_url?: string | null;
  role: UserRole;
  status: UserStatus;
  created_at?: string;
  updated_at?: string;
}

export interface LoginPayload {
  account: string;
  password: string;
}

export interface RegisterPayload {
  username: string;
  email: string;
  password: string;
}

export interface LoginResult {
  access_token: string;
  expires_in: number;
  user: User;
}

export interface RefreshResult {
  access_token: string;
  expires_in: number;
}

export interface UpdateProfilePayload {
  nickname?: string | null;
  avatar_url?: string | null;
}

export interface ChangePasswordPayload {
  old_password: string;
  new_password: string;
}

export interface PaginatedResult<T> {
  items: T[];
  total: number;
  page: number;
  size: number;
}

export interface Session {
  id: string;
  user_id: string;
  name?: string | null;
  model: string;
  knowledge_base_id?: string | null;
  knowledge_base_name?: string | null;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: string;
  session_id: string;
  role: 'user' | 'assistant';
  reply_to_message_id?: string | null;
  content: string;
  status: MessageStatus;
  model_used?: string | null;
  grounded: boolean;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  created_at: string;
  updated_at: string;
}

export interface SessionDetail {
  session: Session;
  messages: Message[];
}

export interface CreateSessionPayload {
  name?: string;
  model: string;
  knowledge_base_id?: string;
}

export interface CreateSessionResult {
  session_id: string;
}

export interface SendMessagePayload {
  content: string;
}

export interface ChatStreamMeta {
  message_id: string;
  grounded: boolean;
  model: string;
}

export interface ChatStreamDelta {
  content: string;
}

export interface ChatStreamDone {
  finish_reason: string;
}

export interface ChatStreamError {
  code: number;
  message: string;
}

export interface KnowledgeBase {
  id: string;
  user_id: string;
  name: string;
  description?: string | null;
  embedding_model: string;
  prompt_template?: string | null;
  retrieval_top_k: number;
  similarity_threshold?: number | null;
  last_indexed_at?: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateKnowledgeBasePayload {
  name: string;
  description?: string | null;
  embedding_model: string;
  prompt_template?: string | null;
  retrieval_top_k?: number;
  similarity_threshold?: number | null;
}

export interface UpdateKnowledgeBasePayload {
  name: string;
  description?: string | null;
  prompt_template?: string | null;
  retrieval_top_k?: number;
  similarity_threshold?: number | null;
}

export interface Document {
  id: string;
  knowledge_base_id: string;
  file_id?: string | null;
  title?: string | null;
  original_filename?: string | null;
  status: DocumentStatus;
  error_message?: string | null;
  content_length: number;
  chunk_count: number;
  created_at: string;
  updated_at: string;
}

export interface UploadDocumentResult {
  document_id: string;
  task_id: string;
  status: 'pending';
}
