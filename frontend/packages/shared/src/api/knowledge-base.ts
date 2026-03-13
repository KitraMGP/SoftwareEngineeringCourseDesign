import type {
  CreateKnowledgeBasePayload,
  Document,
  KnowledgeBase,
  PaginatedResult,
  UpdateKnowledgeBasePayload,
  UploadDocumentResult
} from '../types/domain';

import { apiClient, unwrapData } from './http';

export const knowledgeBaseApi = {
  async listKnowledgeBases(params: { page?: number; size?: number; keyword?: string } = {}) {
    return unwrapData<PaginatedResult<KnowledgeBase>>(
      await apiClient.get('/knowledge-bases', { params })
    );
  },

  async createKnowledgeBase(payload: CreateKnowledgeBasePayload) {
    return unwrapData<{ knowledge_base_id: string }>(await apiClient.post('/knowledge-bases', payload));
  },

  async getKnowledgeBase(knowledgeBaseId: string) {
    return unwrapData<KnowledgeBase>(await apiClient.get(`/knowledge-bases/${knowledgeBaseId}`));
  },

  async updateKnowledgeBase(knowledgeBaseId: string, payload: UpdateKnowledgeBasePayload) {
    return unwrapData<KnowledgeBase>(
      await apiClient.put(`/knowledge-bases/${knowledgeBaseId}`, payload)
    );
  },

  async deleteKnowledgeBase(knowledgeBaseId: string): Promise<void> {
    await apiClient.delete(`/knowledge-bases/${knowledgeBaseId}`);
  },

  async reindexKnowledgeBase(knowledgeBaseId: string) {
    return unwrapData<{ task_id: string }>(
      await apiClient.post(`/knowledge-bases/${knowledgeBaseId}/reindex`)
    );
  },

  async listDocuments(
    knowledgeBaseId: string,
    params: { page?: number; size?: number } = {}
  ) {
    return unwrapData<PaginatedResult<Document>>(
      await apiClient.get(`/knowledge-bases/${knowledgeBaseId}/documents`, { params })
    );
  },

  async uploadDocument(knowledgeBaseId: string, file: File) {
    const formData = new FormData();
    formData.append('file', file);

    return unwrapData<UploadDocumentResult>(
      await apiClient.post(`/knowledge-bases/${knowledgeBaseId}/documents`, formData)
    );
  },

  async getDocument(knowledgeBaseId: string, documentId: string) {
    return unwrapData<Document>(
      await apiClient.get(`/knowledge-bases/${knowledgeBaseId}/documents/${documentId}`)
    );
  },

  async deleteDocument(knowledgeBaseId: string, documentId: string): Promise<void> {
    await apiClient.delete(`/knowledge-bases/${knowledgeBaseId}/documents/${documentId}`);
  }
};
