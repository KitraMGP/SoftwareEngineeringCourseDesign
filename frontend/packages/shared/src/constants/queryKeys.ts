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
    ['document', knowledgeBaseId, documentId] as const,
  adminOverview: ['admin', 'overview'] as const,
  adminUsersRoot: ['admin', 'users'] as const,
  adminUsers: (filters: Record<string, unknown> = {}) => ['admin', 'users', filters] as const,
  adminTasksRoot: ['admin', 'tasks'] as const,
  adminTasks: (filters: Record<string, unknown> = {}) => ['admin', 'tasks', filters] as const,
  adminProviderConfigs: ['admin', 'provider-configs'] as const,
  adminSystemSettings: ['admin', 'system-settings'] as const,
  adminQuotaPolicies: ['admin', 'quota-policies'] as const,
  adminAuditLogsRoot: ['admin', 'audit-logs'] as const,
  adminAuditLogs: (filters: Record<string, unknown> = {}) =>
    ['admin', 'audit-logs', filters] as const
};
