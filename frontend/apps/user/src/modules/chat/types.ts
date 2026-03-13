export interface CitationReference {
  id: string;
  title: string;
  documentId?: string;
  knowledgeBaseId?: string;
  rank?: number;
  sourcePage?: number | null;
}

export interface UiChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt?: string;
  tag?: string;
  tagTone?: 'info' | 'success' | 'warning' | 'danger';
  citations?: CitationReference[];
  grounded?: boolean;
  isPreview?: boolean;
  isStreaming?: boolean;
  canRegenerate?: boolean;
}
