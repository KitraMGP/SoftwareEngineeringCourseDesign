export interface CitationReference {
  id: string;
  title: string;
  snippet?: string;
}

export interface UiChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
  createdAt?: string;
  tag?: string;
  citations?: CitationReference[];
  grounded?: boolean;
  isPreview?: boolean;
  isStreaming?: boolean;
}
