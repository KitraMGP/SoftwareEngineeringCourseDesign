import { defineStore } from 'pinia';

import { authApi } from '../api/auth';
import type { LoginPayload, User } from '../types/domain';
import { normalizeApiError } from '../utils/errors';
import { clearStoredAccessToken, readStoredAccessToken, writeStoredAccessToken } from './token';

let bootstrapPromise: Promise<void> | null = null;

interface AuthState {
  accessToken: string | null;
  user: User | null;
  isInitialized: boolean;
  isBootstrapping: boolean;
}

export const useAuthStore = defineStore('shared-auth', {
  state: (): AuthState => ({
    accessToken: readStoredAccessToken(),
    user: null,
    isInitialized: false,
    isBootstrapping: false
  }),

  getters: {
    isAuthenticated: (state) => !!state.accessToken && !!state.user,
    isAdmin: (state) => state.user?.role === 'admin',
    displayName: (state) => state.user?.nickname || state.user?.username || '访客'
  },

  actions: {
    setAccessToken(token: string | null) {
      this.accessToken = token;

      if (token) {
        writeStoredAccessToken(token);
      } else {
        clearStoredAccessToken();
      }
    },

    setUser(user: User | null) {
      this.user = user;
    },

    clearAuth() {
      this.setAccessToken(null);
      this.user = null;
    },

    async bootstrap() {
      if (this.isInitialized) {
        return;
      }

      if (bootstrapPromise) {
        return bootstrapPromise;
      }

      this.isBootstrapping = true;
      bootstrapPromise = (async () => {
        try {
          if (await this.tryRestoreFromStoredAccessToken()) {
            return;
          }

          if (await this.tryRestoreFromRefreshToken()) {
            return;
          }

          this.clearAuth();
        } catch {
          this.clearAuth();
        } finally {
          this.isInitialized = true;
          this.isBootstrapping = false;
        }
      })();

      try {
        await bootstrapPromise;
      } finally {
        bootstrapPromise = null;
      }
    },

    async tryRestoreFromStoredAccessToken(): Promise<boolean> {
      if (!this.accessToken) {
        return false;
      }

      try {
        await this.fetchCurrentUser();
        return true;
      } catch (error) {
        this.user = null;
        const normalized = normalizeApiError(error);

        if (normalized.status && normalized.status !== 401) {
          throw normalized;
        }

        return false;
      }
    },

    async tryRestoreFromRefreshToken(): Promise<boolean> {
      try {
        await this.refreshAccessToken();
        await this.fetchCurrentUser();
        return true;
      } catch (error) {
        const normalized = normalizeApiError(error);

        if (normalized.status && normalized.status !== 401) {
          throw normalized;
        }

        return false;
      }
    },

    async login(payload: LoginPayload) {
      const result = await authApi.login(payload);
      this.setAccessToken(result.access_token);
      this.user = result.user;
      this.isInitialized = true;
      return result.user;
    },

    async refreshAccessToken(): Promise<string | null> {
      const result = await authApi.refresh();
      this.setAccessToken(result.access_token);
      return result.access_token;
    },

    async fetchCurrentUser() {
      const user = await authApi.getCurrentUser();
      this.user = user;
      return user;
    },

    async logout() {
      try {
        if (this.accessToken) {
          await authApi.logout();
        }
      } finally {
        this.clearAuth();
        this.isInitialized = true;
      }
    }
  }
});
