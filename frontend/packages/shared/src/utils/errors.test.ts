import { describe, expect, it } from 'vitest';

import { ApiRequestError } from '../types/api';
import { getErrorMessage, toFieldErrorMap } from './errors';

describe('error helpers', () => {
  it('maps field errors to first message', () => {
    const error = new ApiRequestError('validation failed', {
      details: [
        { field: 'email', message: 'email is invalid' },
        { field: 'email', message: 'email is required' },
        { field: 'password', message: 'password is required' }
      ]
    });

    expect(toFieldErrorMap(error)).toEqual({
      email: 'email is invalid',
      password: 'password is required'
    });
  });

  it('returns fallback for unknown errors', () => {
    expect(getErrorMessage(new Error('boom'))).toBe('boom');
    expect(getErrorMessage(null, 'fallback')).toBe('请求失败');
  });
});
