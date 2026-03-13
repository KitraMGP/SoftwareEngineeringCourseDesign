import type {
  ChangePasswordPayload,
  LoginPayload,
  LoginResult,
  RefreshResult,
  RegisterPayload,
  UpdateProfilePayload,
  User
} from '../types/domain';

import { apiClient, rawClient, unwrapData } from './http';

export const authApi = {
  async register(payload: RegisterPayload): Promise<{ user_id: string }> {
    return unwrapData(await rawClient.post('/auth/register', payload));
  },

  async login(payload: LoginPayload): Promise<LoginResult> {
    return unwrapData(await rawClient.post('/auth/login', payload));
  },

  async refresh(): Promise<RefreshResult> {
    return unwrapData(await rawClient.post('/auth/refresh', {}));
  },

  async logout(): Promise<void> {
    await apiClient.post('/auth/logout');
  },

  async getCurrentUser(): Promise<User> {
    return unwrapData(await apiClient.get('/users/me'));
  },

  async updateCurrentUser(payload: UpdateProfilePayload): Promise<void> {
    await apiClient.put('/users/me', payload);
  },

  async changePassword(payload: ChangePasswordPayload): Promise<void> {
    await apiClient.put('/users/me/password', payload);
  }
};
