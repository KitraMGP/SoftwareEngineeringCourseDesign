import type {
  AdminOverview,
  AdminTask,
  AdminUser,
  AuditLog,
  PaginatedResult,
  ProviderConfig,
  QuotaPolicy,
  SystemSetting
} from '../types/domain';

import { apiClient, unwrapData } from './http';

export const adminApi = {
  async getOverview() {
    return unwrapData<AdminOverview>(await apiClient.get('/admin/overview'));
  },

  async listUsers(
    params: {
      page?: number;
      size?: number;
      keyword?: string;
      status?: string;
      role?: string;
    } = {}
  ) {
    return unwrapData<PaginatedResult<AdminUser>>(
      await apiClient.get('/admin/users', { params })
    );
  },

  async freezeUser(userId: string) {
    return unwrapData<AdminUser>(await apiClient.post(`/admin/users/${userId}/freeze`));
  },

  async unfreezeUser(userId: string) {
    return unwrapData<AdminUser>(await apiClient.post(`/admin/users/${userId}/unfreeze`));
  },

  async listTasks(
    params: { page?: number; size?: number; status?: string; task_type?: string } = {}
  ) {
    return unwrapData<PaginatedResult<AdminTask>>(
      await apiClient.get('/admin/tasks', { params })
    );
  },

  async retryTask(taskId: string) {
    return unwrapData<AdminTask>(await apiClient.post(`/admin/tasks/${taskId}/retry`));
  },

  async listProviderConfigs() {
    return unwrapData<{ items: ProviderConfig[] }>(
      await apiClient.get('/admin/provider-configs')
    ).items;
  },

  async updateProviderApiKey(provider: string, payload: { api_key: string }) {
    return unwrapData<ProviderConfig>(
      await apiClient.put(`/admin/provider-configs/${provider}`, payload)
    );
  },

  async listSystemSettings() {
    return unwrapData<{ items: SystemSetting[] }>(await apiClient.get('/admin/settings')).items;
  },

  async listQuotaPolicies() {
    return unwrapData<{ items: QuotaPolicy[] }>(await apiClient.get('/admin/quota-policies'))
      .items;
  },

  async listAuditLogs(params: { page?: number; size?: number } = {}) {
    return unwrapData<PaginatedResult<AuditLog>>(
      await apiClient.get('/admin/audit-logs', { params })
    );
  }
};
