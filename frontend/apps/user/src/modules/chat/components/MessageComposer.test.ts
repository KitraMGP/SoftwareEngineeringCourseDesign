import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';

import MessageComposer from './MessageComposer.vue';

describe('MessageComposer', () => {
  it('submits on Enter and preserves Shift+Enter for newline', async () => {
    const wrapper = mount(MessageComposer, {
      props: {
        modelValue: 'hello'
      }
    });

    const textarea = wrapper.get('textarea');

    await textarea.trigger('keydown', { key: 'Enter' });
    expect(wrapper.emitted('submit')).toHaveLength(1);

    await textarea.trigger('keydown', { key: 'Enter', shiftKey: true });
    expect(wrapper.emitted('submit')).toHaveLength(1);
  });

  it('uses a compact single-line textarea by default', () => {
    const wrapper = mount(MessageComposer, {
      props: {
        modelValue: ''
      }
    });

    const textarea = wrapper.get('textarea');

    expect(textarea.attributes('rows')).toBe('1');
    expect(textarea.attributes('placeholder')).toContain('Shift+Enter 换行');
  });

  it('does not submit while composing', async () => {
    const wrapper = mount(MessageComposer, {
      props: {
        modelValue: 'hello'
      }
    });

    const textarea = wrapper.get('textarea');

    await textarea.trigger('keydown', { key: 'Enter', isComposing: true });

    expect(wrapper.emitted('submit')).toBeUndefined();
  });

  it('does not submit when stop action is shown', async () => {
    const wrapper = mount(MessageComposer, {
      props: {
        modelValue: 'hello',
        showStopAction: true
      }
    });

    const textarea = wrapper.get('textarea');

    await textarea.trigger('keydown', { key: 'Enter' });

    expect(wrapper.emitted('submit')).toBeUndefined();
  });
});
