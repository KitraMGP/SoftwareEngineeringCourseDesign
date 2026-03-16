import { describe, expect, it } from 'vitest';

import { renderMarkdownToHtml } from './renderMarkdown';

describe('renderMarkdownToHtml', () => {
  it('renders common markdown blocks and inline formatting', () => {
    const html = renderMarkdownToHtml(`# 标题

这里有 **加粗**、*强调* 和 [链接](https://example.com)。

- 第一项
- 第二项

> 引用内容

\`\`\`ts
const value = 1;
\`\`\``);

    expect(html).toContain('<h1>标题</h1>');
    expect(html).toContain('<strong>加粗</strong>');
    expect(html).toContain('<em>强调</em>');
    expect(html).toContain('<a href="https://example.com"');
    expect(html).toContain('<ul><li>第一项</li><li>第二项</li></ul>');
    expect(html).toContain('<blockquote><p>引用内容</p></blockquote>');
    expect(html).toContain('<pre><code class="language-ts">const value = 1;</code></pre>');
  });

  it('escapes raw html and rejects unsafe links', () => {
    const html = renderMarkdownToHtml(
      '<script>alert(1)</script> [bad](javascript:alert(1)) and `const a = "<b>";`'
    );

    expect(html).toContain('&lt;script&gt;alert(1)&lt;/script&gt;');
    expect(html).not.toContain('javascript:alert(1)');
    expect(html).toContain('<code>const a = &quot;&lt;b&gt;&quot;;</code>');
  });

  it('renders indented list items like model replies in the chat stream', () => {
    const html = renderMarkdownToHtml(`服务层

  * **技术实现**：主要由一系列使用 Java 和 Spring Boot 框架编写的可扩展微服务组成。
  * **核心功能**：
  * 提供 RESTful API 供展示层调用。`);

    expect(html).toContain('<p>服务层</p>');
    expect(html).toContain('<ul>');
    expect(html).toContain('<li><strong>技术实现</strong>：主要由一系列使用 Java 和 Spring Boot 框架编写的可扩展微服务组成。</li>');
    expect(html).toContain('<li><strong>核心功能</strong>：</li>');
    expect(html).toContain('<li>提供 RESTful API 供展示层调用。</li>');
  });
});
