<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { ElMessage } from 'element-plus';
import { watch } from 'vue';
import { useRouter } from 'vue-router';

import { chatApi, DEFAULT_CHAT_MODEL, EmptyStatePanel, PRODUCT_NAME, PRODUCT_TAGLINE, queryKeys } from '@private-kb/shared';
import { getErrorMessage } from '@private-kb/shared/utils/errors';

import { useChatUiStore } from '../../../stores/useChatUiStore';

const router = useRouter();
const queryClient = useQueryClient();
const chatUiStore = useChatUiStore();

const sessionsQuery = useQuery({
  queryKey: queryKeys.sessions({ page: 1, size: 20 }),
  queryFn: () => chatApi.listSessions({ page: 1, size: 20 })
});

const createSessionMutation = useMutation({
  mutationFn: () =>
    chatApi.createSession({
      model: DEFAULT_CHAT_MODEL,
      knowledge_base_id: chatUiStore.preferredKnowledgeBaseId || undefined
    }),
  onSuccess: async ({ session_id }) => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.sessionsRoot });
    router.push({ name: 'session-detail', params: { sessionId: session_id } });
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

watch(
  () => sessionsQuery.data.value?.items,
  (items) => {
    if (items?.length) {
      router.replace({ name: 'session-detail', params: { sessionId: items[0].id } });
    }
  },
  { immediate: true }
);
</script>

<template>
  <div class="flex h-full flex-1 flex-col items-center justify-center px-5 py-10 lg:px-12">
    <div class="w-full max-w-5xl space-y-8">
      <div class="space-y-4 text-center">
        <p class="text-xs uppercase tracking-[0.36em] text-slate-400">Conversation workspace</p>
        <h1 class="font-serif text-4xl text-slate-900 lg:text-5xl">{{ PRODUCT_NAME }}</h1>
        <p class="mx-auto max-w-2xl text-sm leading-7 text-slate-500 md:text-base">
          {{ PRODUCT_TAGLINE }}。当前版本已接通会话、知识库管理与空会话实时问答，绑定知识库后的检索式问答仍在联调中。
        </p>
      </div>

      <div class="grid gap-5 lg:grid-cols-[1.2fr_0.8fr]">
        <EmptyStatePanel
          title="开始你的第一轮整理"
          description="如果你刚进入系统，可以先创建空会话，或者先挑一个知识库作为下一轮对话上下文。"
          action-label="创建空会话"
          @action="createSessionMutation.mutate()"
        />

        <div class="space-y-4">
          <div class="rounded-[28px] border border-white/60 bg-white/82 p-6 shadow-soft backdrop-blur">
            <p class="text-xs uppercase tracking-[0.3em] text-slate-400">当前准备</p>
            <p class="mt-4 font-serif text-2xl text-slate-900">
              {{ chatUiStore.preferredKnowledgeBaseName || '尚未选择知识库' }}
            </p>
            <p class="mt-3 text-sm leading-6 text-slate-500">
              {{ chatUiStore.preferredKnowledgeBaseName ? '下一次会话将基于所选知识库创建。' : '若先选择知识库，下一次会话会自动带上对应上下文。' }}
            </p>
            <button
              type="button"
              class="mt-6 rounded-full border border-slate-200 px-4 py-2.5 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
              @click="chatUiStore.openKnowledgeBaseDrawer()"
            >
              打开知识库抽屉
            </button>
          </div>

          <div class="rounded-[28px] border border-white/60 bg-slate-900 px-6 py-6 text-white shadow-soft">
            <p class="text-xs uppercase tracking-[0.32em] text-white/55">当前范围</p>
            <ul class="mt-4 space-y-3 text-sm leading-6 text-white/78">
              <li>会话创建、列表和删除已联调</li>
              <li>知识库与文档管理已联调</li>
              <li>未绑定知识库的空会话已支持流式问答</li>
              <li>知识库问答、停止生成与重生成仍待接入</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
