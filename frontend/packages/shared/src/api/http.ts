import axios, { type AxiosInstance, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';

import { ApiRequestError, type ApiSuccessEnvelope } from '../types/api';
import { normalizeApiError } from '../utils/errors';

interface ApiAuthController {
  getAccessToken: () => string | null;
  refreshAccessToken: () => Promise<string | null>;
  clearAuth: () => void;
  onUnauthorized?: () => void;
}

interface RetriableConfig extends InternalAxiosRequestConfig {
  _retry?: boolean;
}

export const apiBaseURL = (import.meta.env.VITE_API_BASE_URL ?? '/api/v1').replace(/\/$/, '');

export const rawClient = axios.create({
  baseURL: apiBaseURL,
  withCredentials: true
});

export const apiClient = axios.create({
  baseURL: apiBaseURL,
  withCredentials: true
});

let authController: ApiAuthController | null = null;
let refreshPromise: Promise<string | null> | null = null;

function isRefreshRelated(url?: string): boolean {
  return !!url && url.includes('/auth/refresh');
}

function shouldRetry401(config?: RetriableConfig): boolean {
  if (!config || config._retry) {
    return false;
  }

  if (!authController) {
    return false;
  }

  return !isRefreshRelated(config.url);
}

async function refreshAccessTokenOnce(): Promise<string | null> {
  if (!authController) {
    return null;
  }

  if (!refreshPromise) {
    refreshPromise = authController
      .refreshAccessToken()
      .catch(() => null)
      .finally(() => {
        refreshPromise = null;
      });
  }

  return refreshPromise;
}

apiClient.interceptors.request.use((config) => {
  const token = authController?.getAccessToken();

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const normalized = normalizeApiError(error);
    const config = error.config as RetriableConfig | undefined;

    if (error.response?.status !== 401 || !shouldRetry401(config)) {
      return Promise.reject(normalized);
    }

    config!._retry = true;
    const nextToken = await refreshAccessTokenOnce();

    if (!nextToken) {
      authController?.clearAuth();
      authController?.onUnauthorized?.();
      return Promise.reject(normalized);
    }

    config!.headers.Authorization = `Bearer ${nextToken}`;
    return apiClient.request(config!);
  }
);

export function configureApiClient(controller: ApiAuthController): void {
  authController = controller;
}

export function resolveApiUrl(path: string): string {
  if (/^https?:\/\//.test(path)) {
    return path;
  }

  const normalizedPath = path.startsWith('/') ? path : `/${path}`;
  return `${apiBaseURL}${normalizedPath}`;
}

export function unwrapData<T>(response: AxiosResponse<ApiSuccessEnvelope<T>>): T {
  return response.data.data;
}

export function ensureApiError(error: unknown): ApiRequestError {
  return normalizeApiError(error);
}

export function createQueryClientDefaults() {
  return {
    queries: {
      retry: 1,
      refetchOnWindowFocus: false
    },
    mutations: {
      retry: 0
    }
  };
}

export type { AxiosInstance };
