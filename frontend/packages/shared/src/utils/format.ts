export function getInitials(label?: string | null): string {
  const text = (label ?? '').trim();
  if (!text) {
    return 'AI';
  }

  return Array.from(text).slice(0, 2).join('').toUpperCase();
}

export function formatBytes(size?: number | null): string {
  if (!size || size <= 0) {
    return '0 B';
  }

  const units = ['B', 'KB', 'MB', 'GB'];
  let value = size;
  let index = 0;

  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }

  return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`;
}

export function shortenText(value?: string | null, limit = 120): string {
  const text = (value ?? '').trim();
  if (!text) {
    return '';
  }

  return text.length > limit ? `${text.slice(0, limit).trim()}...` : text;
}
