<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue';

import { ASSISTANT_NAME, AppStatusBadge, formatDateTime } from '@private-kb/shared';

import type { UiChatMessage } from '../types';

const props = defineProps<{
  messages: UiChatMessage[];
}>();

const renderedMessages = computed(() => props.messages);
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
    class="soft-scrollbar min-h-0 flex-1 space-y-6 overflow-y-auto px-5 py-6 lg:px-8"
  >
    <article
      v-for="message in renderedMessages"
      :key="message.id"
      class="chat-message-enter flex"
      :class="message.role === 'user' ? 'justify-end' : 'justify-start'"
    >
      <div class="flex max-w-3xl gap-4" :class="message.role === 'user' ? 'flex-row-reverse' : ''">
        <div
          v-if="message.role === 'assistant'"
          class="mt-1 flex h-11 w-11 shrink-0 items-center justify-center rounded-full bg-slate-900 text-xs font-semibold uppercase tracking-[0.18em] text-white"
        >
          AI
        </div>

        <div class="space-y-3">
          <div
            class="rounded-[28px] px-5 py-4 shadow-soft"
            :class="
              message.role === 'user'
                ? 'bg-[linear-gradient(135deg,#bfe9f4,#cfeef2)] text-slate-900'
                : 'border border-white/60 bg-white/84 text-slate-900 backdrop-blur'
            "
          >
            <div v-if="message.role === 'assistant'" class="mb-3 flex items-center gap-3">
              <p class="text-sm font-semibold text-slate-900">{{ ASSISTANT_NAME }}</p>
              <AppStatusBadge v-if="message.tag" :tone="message.isStreaming ? 'info' : 'warning'">
                {{ message.tag }}
              </AppStatusBadge>
              <span class="text-xs text-slate-400">
                {{ formatDateTime(message.createdAt) }}
              </span>
            </div>
            <p
              class="whitespace-pre-wrap text-sm leading-7 md:text-[15px]"
              :class="message.isStreaming && !message.content ? 'text-slate-400' : ''"
            >
              {{ message.content || (message.isStreaming ? '正在生成回答...' : '') }}
            </p>
          </div>

          <div
            v-if="message.citations?.length"
            class="space-y-2 rounded-[22px] border border-slate-200 bg-slate-50/80 px-4 py-4"
          >
            <p class="text-xs uppercase tracking-[0.28em] text-slate-400">引用来源</p>
            <div
              v-for="citation in message.citations"
              :key="citation.id"
              class="rounded-[16px] border border-white bg-white px-4 py-3"
            >
              <p class="text-sm font-semibold text-slate-900">{{ citation.title }}</p>
              <p v-if="citation.snippet" class="mt-1 text-sm leading-6 text-slate-500">
                {{ citation.snippet }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </article>
  </div>
</template>
