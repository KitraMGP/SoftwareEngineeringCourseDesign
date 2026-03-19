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
  citations?: MessageCitation[];
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

export interface StreamStopResult {
  stopped: boolean;
}

export interface MessageCitation {
  id: string;
  document_chunk_id: string;
  document_id: string;
  knowledge_base_id: string;
  document_name: string;
  rank_no: number;
  source_page?: number | null;
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

export interface AdminOverview {
  user_count: number;
  active_user_count: number;
  knowledge_base_count: number;
  document_count: number;
  pending_task_count: number;
  failed_task_count: number;
}

export type AdminUser = User;

export interface AdminTask {
  id: string;
  task_type: string;
  resource_type: string;
  resource_id?: string | null;
  user_id?: string | null;
  user_label?: string | null;
  status: 'pending' | 'running' | 'succeeded' | 'failed' | 'cancelled';
  attempt_count: number;
  max_attempts: number;
  next_run_at?: string | null;
  started_at?: string | null;
  finished_at?: string | null;
  error_code?: string | null;
  error_message?: string | null;
  created_at: string;
  updated_at: string;
}

export interface ProviderConfig {
  id?: string | null;
  provider: string;
  base_url: string;
  default_chat_model: string;
  default_embedding_model: string;
  is_enabled: boolean;
  has_api_key: boolean;
  api_key_source: 'database' | 'environment' | 'missing';
  created_at?: string | null;
  updated_at?: string | null;
}

export interface SystemSetting {
  key: string;
  category: string;
  value: unknown;
  description?: string | null;
  updated_at: string;
}

export interface QuotaPolicy {
  id: string;
  scope_type: 'system_default' | 'user';
  scope_id?: string | null;
  daily_total_tokens_limit: number;
  storage_bytes_limit: number;
  document_count_limit: number;
  warn_ratio: string;
  created_at: string;
  updated_at: string;
}

export interface AuditLog {
  id: number;
  actor_user_id?: string | null;
  actor_label?: string | null;
  actor_role: 'user' | 'admin' | 'system';
  action: string;
  resource_type?: string | null;
  resource_id?: string | null;
  target_user_id?: string | null;
  target_label?: string | null;
  result: 'success' | 'failure';
  metadata: unknown;
  created_at: string;
}
