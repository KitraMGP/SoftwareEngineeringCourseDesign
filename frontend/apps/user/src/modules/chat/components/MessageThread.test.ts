import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

import MessageThread from './MessageThread.vue';

describe('MessageThread', () => {
  it('renders markdown content inside chat bubbles', () => {
    const wrapper = mount(MessageThread, {
      props: {
        messages: [
          {
            id: 'assistant-1',
            role: 'assistant',
            content: '# 标题\n\n这里有 **加粗** 和 `code`。\n\n- 第一项\n- 第二项',
            createdAt: '2026-03-16T00:00:00.000Z'
          }
        ]
      }
    });

    const html = wrapper.html();

    expect(html).toContain('<h1>标题</h1>');
    expect(html).toContain('<strong>加粗</strong>');
    expect(html).toContain('<code>code</code>');
    const items = wrapper.findAll('li').map((item) => item.text());

    expect(items).toEqual(['第一项', '第二项']);
  });
});
