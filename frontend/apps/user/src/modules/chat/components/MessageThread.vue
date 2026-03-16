<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';

import { ASSISTANT_NAME, AppStatusBadge, formatDateTime } from '@private-kb/shared';

import type { UiChatMessage } from '../types';
import { renderMarkdownToHtml } from '../utils/renderMarkdown';

const props = defineProps<{
  messages: UiChatMessage[];
  streaming?: boolean;
  regeneratingMessageId?: string | null;
}>();

const emit = defineEmits<{
  regenerate: [messageId: string];
}>();

const renderedMessages = computed(() =>
  props.messages.map((message) => ({
    ...message,
    renderedContent: renderMarkdownToHtml(message.content)
  }))
);
const containerRef = ref<HTMLElement>();

watch(
  renderedMessages,
  async () => {
    await nextTick();

    if (containerRef.value) {
      containerRef.value.scrollTop = containerRef.value.scrollHeight;
    }
  },
  {
    deep: true,
    immediate: true
  }
);
</script>

<template>
  <div
    ref="containerRef"
    class="soft-scrollbar min-h-0 flex-1 space-y-5 overflow-y-auto px-5 py-5 lg:px-8"
  >
    <article
      v-for="message in renderedMessages"
      :key="message.id"
      class="chat-message-enter group flex"
      :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
    >
      <div class="flex max-w-[min(100%,42rem)] gap-3" :class="message.role === 'user' ? 'flex-row-reverse' : ''">
        <div
          v-if="message.role === 'assistant'"
          class="mt-1 flex h-10 w-10 shrink-0 items-center justify-center rounded-full bg-slate-900 text-[11px] font-semibold uppercase tracking-[0.18em] text-white"
        >
          AI
        </div>

        <div class="space-y-2">
          <div
            class="rounded-[24px] px-5 py-4 shadow-soft"
            :class="
              message.role === 'user'
                ? 'border border-sky-100/80 bg-sky-50/80 text-slate-900'
                : 'border border-white/70 bg-white/90 text-slate-900 backdrop-blur'
            "
          >
            <div v-if="message.role === 'assistant'" class="mb-2 flex items-center gap-2">
              <p class="text-sm font-semibold text-slate-900">{{ ASSISTANT_NAME }}</p>
              <AppStatusBadge
                v-if="message.tag"
                :tone="message.isStreaming ? 'info' : message.tagTone || 'warning'"
              >
                {{ message.tag }}
              </AppStatusBadge>
              <span class="text-[11px] text-slate-400">
                {{ formatDateTime(message.createdAt) }}
              </span>
              <button
                v-if="message.canRegenerate"
                type="button"
                class="ml-auto rounded-full border border-slate-200 bg-white px-2.5 py-1 text-[11px] font-medium text-slate-600 opacity-0 transition hover:border-slate-300 hover:bg-slate-50 group-hover:opacity-100 focus:opacity-100 disabled:cursor-not-allowed disabled:border-slate-200 disabled:bg-slate-100 disabled:text-slate-400 disabled:opacity-100"
                :disabled="streaming"
                @click="emit('regenerate', message.id)"
              >
                {{ regeneratingMessageId === message.id ? '重新生成中...' : '重新生成' }}
              </button>
            </div>
            <!-- eslint-disable vue/no-v-html -->
            <div
              v-if="message.content"
              class="message-markdown text-sm leading-7 md:text-[15px]"
              v-html="message.renderedContent"
            />
            <!-- eslint-enable vue/no-v-html -->
            <p
              v-else
              class="text-sm leading-7 md:text-[15px]"
              :class="message.isStreaming ? 'text-slate-400' : ''"
            >
              {{ message.isStreaming ? '正在生成回答...' : '' }}
            </p>
          </div>

          <div
            v-if="message.citations?.length"
            class="space-y-2 rounded-[18px] border border-slate-200 bg-slate-50/85 px-4 py-3"
          >
            <p class="text-xs uppercase tracking-[0.28em] text-slate-400">引用来源</p>
            <div
              v-for="citation in message.citations"
              :key="citation.id"
              class="rounded-[16px] border border-white bg-white px-4 py-3"
            >
              <p class="text-sm font-semibold text-slate-900">{{ citation.title }}</p>
              <p v-if="citation.sourcePage" class="mt-1 text-sm leading-6 text-slate-500">
                第 {{ citation.sourcePage }} 页
              </p>
            </div>
          </div>
        </div>
      </div>
    </article>
  </div>
</template>

<style scoped>
.message-markdown :deep(h1),
.message-markdown :deep(h2),
.message-markdown :deep(h3),
.message-markdown :deep(h4),
.message-markdown :deep(h5),
.message-markdown :deep(h6) {
  margin: 0;
  color: #0f172a;
  font-family: 'Noto Serif SC', 'Songti SC', 'Source Han Serif SC', serif;
  font-weight: 600;
  line-height: 1.35;
}

.message-markdown :deep(h1) {
  font-size: 1.45rem;
}

.message-markdown :deep(h2) {
  font-size: 1.3rem;
}

.message-markdown :deep(h3) {
  font-size: 1.15rem;
}

.message-markdown :deep(p),
.message-markdown :deep(ul),
.message-markdown :deep(ol),
.message-markdown :deep(blockquote),
.message-markdown :deep(pre) {
  margin: 0;
}

.message-markdown :deep(* + p),
.message-markdown :deep(* + ul),
.message-markdown :deep(* + ol),
.message-markdown :deep(* + blockquote),
.message-markdown :deep(* + pre),
.message-markdown :deep(* + h1),
.message-markdown :deep(* + h2),
.message-markdown :deep(* + h3),
.message-markdown :deep(* + h4),
.message-markdown :deep(* + h5),
.message-markdown :deep(* + h6) {
  margin-top: 0.8rem;
}

.message-markdown :deep(ul),
.message-markdown :deep(ol) {
  padding-left: 1.35rem;
}

.message-markdown :deep(li + li) {
  margin-top: 0.35rem;
}

.message-markdown :deep(blockquote) {
  border-left: 3px solid rgba(148, 163, 184, 0.65);
  padding-left: 0.9rem;
  color: #475569;
}

.message-markdown :deep(a) {
  color: #1d4ed8;
  text-decoration: underline;
  text-underline-offset: 0.16em;
  word-break: break-word;
}

.message-markdown :deep(code) {
  border-radius: 0.55rem;
  background: rgba(15, 23, 42, 0.08);
  padding: 0.15rem 0.4rem;
  font-family: 'JetBrains Mono', 'SFMono-Regular', monospace;
  font-size: 0.92em;
}

.message-markdown :deep(pre) {
  overflow-x: auto;
  border-radius: 1rem;
  background: #0f172a;
  padding: 0.9rem 1rem;
  color: #e2e8f0;
}

.message-markdown :deep(pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

.message-markdown :deep(strong) {
  font-weight: 700;
}

.message-markdown :deep(del) {
  text-decoration-thickness: 1px;
}
</style>
