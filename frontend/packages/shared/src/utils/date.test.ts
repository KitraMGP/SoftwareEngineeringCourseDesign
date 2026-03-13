import { describe, expect, it, vi } from 'vitest';

import { formatRelativeTime } from './date';

describe('date helpers', () => {
  it('formats recent relative time', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-03-13T10:00:00.000Z'));

    expect(formatRelativeTime('2026-03-13T09:30:00.000Z')).toBe('30 分钟前');

    vi.useRealTimers();
  });
});
