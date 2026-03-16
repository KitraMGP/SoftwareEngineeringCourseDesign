function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function sanitizeHref(rawHref: string): string | null {
  const href = rawHref.trim();

  if (!href) {
    return null;
  }

  if (/^(https?:|mailto:|tel:)/i.test(href) || /^(\/|#)/.test(href)) {
    return escapeHtml(href);
  }

  return null;
}

function createPlaceholderToken(index: number): string {
  return `@@MD_TOKEN_${index}@@`;
}

function renderInlineMarkdown(content: string): string {
  const placeholders: string[] = [];

  const withTokens = content
    .replace(/`([^`]+)`/g, (_match, code: string) => {
      const token = createPlaceholderToken(placeholders.length);
      placeholders.push(`<code>${escapeHtml(code)}</code>`);
      return token;
    })
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, (_match, label: string, href: string) => {
      const token = createPlaceholderToken(placeholders.length);
      const safeHref = sanitizeHref(href);

      placeholders.push(
        safeHref
          ? `<a href="${safeHref}" target="_blank" rel="noreferrer noopener">${escapeHtml(label)}</a>`
          : escapeHtml(label)
      );

      return token;
    });

  return placeholders
    .reduce((html, placeholder, index) => {
      const token = createPlaceholderToken(index);
      return html.split(token).join(placeholder);
    }, escapeHtml(withTokens))
    .replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>')
    .replace(/__([^_]+)__/g, '<strong>$1</strong>')
    .replace(/~~([^~]+)~~/g, '<del>$1</del>')
    .replace(/(^|[^*])\*([^*]+)\*(?!\*)/g, '$1<em>$2</em>')
    .replace(/(^|[^_])_([^_]+)_(?!_)/g, '$1<em>$2</em>');
}

function renderParagraph(lines: string[]): string {
  return `<p>${lines.map((line) => renderInlineMarkdown(line.trim())).join('<br />')}</p>`;
}

function renderList(lines: string[], ordered: boolean): string {
  const tagName = ordered ? 'ol' : 'ul';
  const itemPattern = ordered ? /^\d+\.\s+(.*)$/ : /^[-*+]\s+(.*)$/;
  const items: string[] = [];

  let index = 0;

  while (index < lines.length) {
    const match = lines[index]?.match(itemPattern);

    if (!match) {
      index += 1;
      continue;
    }

    const itemLines = [match[1]];
    index += 1;

    while (
      index < lines.length &&
      lines[index] &&
      !lines[index].match(/^\d+\.\s+/) &&
      !lines[index].match(/^[-*+]\s+/)
    ) {
      itemLines.push(lines[index].trim());
      index += 1;
    }

    items.push(`<li>${itemLines.map((line) => renderInlineMarkdown(line)).join('<br />')}</li>`);
  }

  return `<${tagName}>${items.join('')}</${tagName}>`;
}

function renderCodeBlock(language: string | undefined, codeLines: string[]): string {
  const safeLanguage = (language || '').trim().toLowerCase().replace(/[^a-z0-9-]/g, '');
  const languageClass = safeLanguage ? ` class="language-${safeLanguage}"` : '';

  return `<pre><code${languageClass}>${escapeHtml(codeLines.join('\n'))}</code></pre>`;
}

function isListStart(line: string): boolean {
  const normalizedLine = line.trimStart();
  return /^[-*+]\s+/.test(normalizedLine) || /^\d+\.\s+/.test(normalizedLine);
}

function isBlockBoundary(line: string): boolean {
  const normalizedLine = line.trimStart();
  return (
    /^#{1,6}\s+/.test(normalizedLine) ||
    /^```/.test(normalizedLine) ||
    /^>\s?/.test(normalizedLine) ||
    isListStart(normalizedLine)
  );
}

export function renderMarkdownToHtml(markdown: string): string {
  const normalized = markdown.replace(/\r\n?/g, '\n').trim();

  if (!normalized) {
    return '';
  }

  const lines = normalized.split('\n');
  const blocks: string[] = [];

  for (let index = 0; index < lines.length; ) {
    const line = lines[index].trimStart();

    if (!line || !line.trim()) {
      index += 1;
      continue;
    }

    const codeFenceMatch = line.match(/^```([\w-]+)?\s*$/);

    if (codeFenceMatch) {
      const codeLines: string[] = [];
      index += 1;

      while (index < lines.length && !lines[index].match(/^```\s*$/)) {
        codeLines.push(lines[index]);
        index += 1;
      }

      if (index < lines.length) {
        index += 1;
      }

      blocks.push(renderCodeBlock(codeFenceMatch[1], codeLines));
      continue;
    }

    const headingMatch = line.match(/^(#{1,6})\s+(.*)$/);

    if (headingMatch) {
      const level = headingMatch[1].length;
      blocks.push(`<h${level}>${renderInlineMarkdown(headingMatch[2].trim())}</h${level}>`);
      index += 1;
      continue;
    }

    if (/^>\s?/.test(line)) {
      const quotedLines: string[] = [];

      while (index < lines.length && /^>\s?/.test(lines[index].trimStart())) {
        quotedLines.push(lines[index].trimStart().replace(/^>\s?/, ''));
        index += 1;
      }

      blocks.push(`<blockquote>${renderMarkdownToHtml(quotedLines.join('\n'))}</blockquote>`);
      continue;
    }

    if (/^[-*+]\s+/.test(line)) {
      const listLines: string[] = [];

      while (index < lines.length && lines[index] && /^[-*+]\s+/.test(lines[index].trimStart())) {
        listLines.push(lines[index].trimStart());
        index += 1;
      }

      blocks.push(renderList(listLines, false));
      continue;
    }

    if (/^\d+\.\s+/.test(line)) {
      const listLines: string[] = [];

      while (index < lines.length && lines[index] && /^\d+\.\s+/.test(lines[index].trimStart())) {
        listLines.push(lines[index].trimStart());
        index += 1;
      }

      blocks.push(renderList(listLines, true));
      continue;
    }

    const paragraphLines: string[] = [];

    while (
      index < lines.length &&
      lines[index] &&
      lines[index].trim() &&
      !isBlockBoundary(lines[index])
    ) {
      paragraphLines.push(lines[index]);
      index += 1;
    }

    blocks.push(renderParagraph(paragraphLines));
  }

  return blocks.join('');
}
