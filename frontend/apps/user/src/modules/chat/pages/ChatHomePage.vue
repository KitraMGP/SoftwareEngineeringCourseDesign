<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { ElMessage } from 'element-plus';
import { watch } from 'vue';
import { useRouter } from 'vue-router';

import {
  AppStatusBadge,
  SurfaceCard,
  chatApi,
  DEFAULT_CHAT_MODEL,
  PRODUCT_NAME,
  PRODUCT_TAGLINE,
  queryKeys
} from '@private-kb/shared';
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
  <div class="flex h-full flex-1 flex-col px-5 py-6 lg:px-8">
    <div class="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-4">
      <section class="rounded-[28px] border border-white/68 bg-white/76 px-6 py-6 shadow-soft backdrop-blur">
        <div class="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div class="max-w-3xl space-y-3">
            <p class="text-xs uppercase tracking-[0.36em] text-slate-400">Conversation workspace</p>
            <h1 class="font-serif text-3xl text-slate-900 lg:text-[2.6rem]">{{ PRODUCT_NAME }}</h1>
            <p class="text-sm leading-7 text-slate-600 md:text-base">
              {{ PRODUCT_TAGLINE }}。从这里开始新的普通会话，或先选定知识库，把资料整理和提问动作放在同一个工作区完成。
            </p>
          </div>

          <div class="flex flex-wrap gap-3">
            <button
              type="button"
              class="rounded-full bg-slate-900 px-5 py-3 text-sm font-semibold text-white transition hover:bg-slate-800"
              @click="createSessionMutation.mutate()"
            >
              {{ createSessionMutation.isPending.value ? '创建中...' : '开始空会话' }}
            </button>
            <button
              type="button"
              class="rounded-full border border-slate-200 bg-white px-5 py-3 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
              @click="chatUiStore.openKnowledgeBaseDrawer()"
            >
              选择知识库
            </button>
          </div>
        </div>
      </section>

      <div class="grid flex-1 gap-4 xl:grid-cols-[1.2fr_0.8fr]">
        <SurfaceCard tone="soft" class="flex flex-col justify-between">
          <div class="space-y-4">
            <p class="text-xs uppercase tracking-[0.32em] text-slate-400">Start here</p>
            <div class="space-y-3">
              <h2 class="font-serif text-2xl text-slate-900 lg:text-[2rem]">先把下一步动作说清楚</h2>
              <p class="max-w-2xl text-sm leading-7 text-slate-600">
                首页只保留开始动作，不再堆叠大段状态说明。创建会话后直接进入消息区，若要结合资料提问，再为下一次会话指定知识库。
              </p>
            </div>
          </div>

          <div class="mt-6 grid gap-3 sm:grid-cols-3">
            <div class="rounded-[18px] border border-white/75 bg-white/80 px-4 py-4">
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">步骤 1</p>
              <p class="mt-2 text-sm font-medium text-slate-900">创建会话</p>
              <p class="mt-1 text-sm leading-6 text-slate-500">直接进入聊天区，避免首页承载过多说明。</p>
            </div>
            <div class="rounded-[18px] border border-white/75 bg-white/80 px-4 py-4">
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">步骤 2</p>
              <p class="mt-2 text-sm font-medium text-slate-900">准备知识库</p>
              <p class="mt-1 text-sm leading-6 text-slate-500">资料整理和问答切换保持在同一套工作流里。</p>
            </div>
            <div class="rounded-[18px] border border-white/75 bg-white/80 px-4 py-4">
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">步骤 3</p>
              <p class="mt-2 text-sm font-medium text-slate-900">持续追问</p>
              <p class="mt-1 text-sm leading-6 text-slate-500">消息历史会保留在左侧会话列表，便于回看。</p>
            </div>
          </div>
        </SurfaceCard>

        <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-1">
          <SurfaceCard>
            <p class="text-xs uppercase tracking-[0.3em] text-slate-400">当前准备</p>
            <p class="mt-3 font-serif text-[1.7rem] leading-tight text-slate-900">
              {{ chatUiStore.preferredKnowledgeBaseName || '尚未选择知识库' }}
            </p>
            <p class="mt-3 text-sm leading-6 text-slate-500">
              {{ chatUiStore.preferredKnowledgeBaseName ? '下一次新会话会直接绑定到当前知识库。' : '如果你准备围绕资料提问，先从右侧抽屉选择一个知识库。' }}
            </p>
          </SurfaceCard>

          <SurfaceCard>
            <p class="text-xs uppercase tracking-[0.3em] text-slate-400">当前能力</p>
            <div class="mt-4 flex flex-wrap gap-2">
              <AppStatusBadge tone="success">会话管理</AppStatusBadge>
              <AppStatusBadge tone="success">知识库维护</AppStatusBadge>
              <AppStatusBadge tone="info">实时问答</AppStatusBadge>
              <AppStatusBadge tone="info">文档状态跟踪</AppStatusBadge>
            </div>
            <p class="mt-4 text-sm leading-6 text-slate-500">
              首页只保留关键状态，细节说明下沉到具体页面，避免首屏信息过厚。
            </p>
          </SurfaceCard>
        </div>
      </div>
    </div>
  </div>
</template>
