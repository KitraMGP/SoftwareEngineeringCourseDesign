import { createPinia, setActivePinia } from 'pinia';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiRequestError } from '../types/api';
import { useAuthStore } from './useAuthStore';

const authApiMock = vi.hoisted(() => ({
  login: vi.fn(),
  refresh: vi.fn(),
  logout: vi.fn(),
  getCurrentUser: vi.fn(),
  updateCurrentUser: vi.fn(),
  changePassword: vi.fn(),
}));

let storedAccessToken: string | null = null;

vi.mock('../api/auth', () => ({
  authApi: authApiMock
}));

vi.mock('./token', () => ({
  readStoredAccessToken: () => storedAccessToken,
  writeStoredAccessToken: (token: string) => {
    storedAccessToken = token;
  },
  clearStoredAccessToken: () => {
    storedAccessToken = null;
  }
}));

describe('useAuthStore bootstrap', () => {
  const user = {
    id: 'user-1',
    username: 'tester',
    email: 'tester@example.com',
    role: 'user' as const,
    status: 'active' as const
  };

  beforeEach(() => {
    setActivePinia(createPinia());
    storedAccessToken = null;
    vi.clearAllMocks();
  });

  it('restores session from stored access token without forcing refresh', async () => {
    storedAccessToken = 'persisted-access-token';
    authApiMock.getCurrentUser.mockResolvedValue(user);

    const store = useAuthStore();
    await store.bootstrap();

    expect(authApiMock.getCurrentUser).toHaveBeenCalledTimes(1);
    expect(authApiMock.refresh).not.toHaveBeenCalled();
    expect(store.accessToken).toBe('persisted-access-token');
    expect(store.user).toEqual(user);
    expect(store.isAuthenticated).toBe(true);
  });

  it('falls back to refresh token when stored access token is no longer valid', async () => {
    storedAccessToken = 'expired-access-token';
    authApiMock.getCurrentUser
      .mockRejectedValueOnce(new ApiRequestError('unauthorized', { status: 401 }))
      .mockResolvedValueOnce(user);
    authApiMock.refresh.mockResolvedValue({
      access_token: 'fresh-access-token',
      expires_in: 1800
    });

    const store = useAuthStore();
    await store.bootstrap();

    expect(authApiMock.refresh).toHaveBeenCalledTimes(1);
    expect(authApiMock.getCurrentUser).toHaveBeenCalledTimes(2);
    expect(store.accessToken).toBe('fresh-access-token');
    expect(store.user).toEqual(user);
    expect(store.isAuthenticated).toBe(true);
  });
});
