export function scopeQueryKey(
  baseKey: readonly unknown[],
  scope: string | null | undefined
): readonly unknown[] {
  return [...baseKey, { scope: scope || 'anonymous' }] as const;
}
