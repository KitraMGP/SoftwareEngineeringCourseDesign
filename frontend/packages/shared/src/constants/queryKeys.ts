export const queryKeys = {
  me: ['me'] as const,
  sessionsRoot: ['sessions'] as const,
  sessions: (filters: Record<string, unknown> = {}) => ['sessions', filters] as const,
  session: (sessionId: string) => ['session', sessionId] as const,
  knowledgeBasesRoot: ['knowledge-bases'] as const,
  knowledgeBases: (filters: Record<string, unknown> = {}) =>
    ['knowledge-bases', filters] as const,
  knowledgeBase: (knowledgeBaseId: string) => ['knowledge-base', knowledgeBaseId] as const,
  documentsRoot: (knowledgeBaseId: string) => ['documents', knowledgeBaseId] as const,
  documents: (knowledgeBaseId: string, filters: Record<string, unknown> = {}) =>
    ['documents', knowledgeBaseId, filters] as const,
  document: (knowledgeBaseId: string, documentId: string) =>
    ['document', knowledgeBaseId, documentId] as const
};
