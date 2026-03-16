<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { ElMessage } from 'element-plus';
import { computed, watch } from 'vue';
import { useRouter } from 'vue-router';

import {
  SurfaceCard,
  chatApi,
  DEFAULT_CHAT_MODEL,
  PRODUCT_NAME,
  PRODUCT_TAGLINE,
  queryKeys,
  scopeQueryKey,
  useAuthStore
} from '@private-kb/shared';
import { getErrorMessage } from '@private-kb/shared/utils/errors';

import { useChatUiStore } from '../../../stores/useChatUiStore';

const router = useRouter();
const queryClient = useQueryClient();
const chatUiStore = useChatUiStore();
const authStore = useAuthStore();
const userScope = computed(() => authStore.user?.id ?? null);

const sessionsQuery = useQuery({
  queryKey: computed(() => scopeQueryKey(queryKeys.sessions({ page: 1, size: 20 }), userScope.value)),
  queryFn: () => chatApi.listSessions({ page: 1, size: 20 }),
  enabled: computed(() => !!userScope.value)
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
              {{ PRODUCT_TAGLINE }}。从这里开始新的会话，或先选定知识库再继续提问。
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
              <h2 class="font-serif text-2xl text-slate-900 lg:text-[2rem]">从这里开始新的工作流</h2>
            </div>
          </div>

          <div class="mt-6 grid gap-3 sm:grid-cols-3">
            <div class="rounded-[18px] border border-white/75 bg-white/80 px-4 py-4">
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">步骤 1</p>
              <p class="mt-2 text-sm font-medium text-slate-900">创建会话</p>
            </div>
            <div class="rounded-[18px] border border-white/75 bg-white/80 px-4 py-4">
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">步骤 2</p>
              <p class="mt-2 text-sm font-medium text-slate-900">选择知识库</p>
            </div>
            <div class="rounded-[18px] border border-white/75 bg-white/80 px-4 py-4">
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">步骤 3</p>
              <p class="mt-2 text-sm font-medium text-slate-900">持续追问</p>
            </div>
          </div>
        </SurfaceCard>

        <SurfaceCard>
          <p class="text-xs uppercase tracking-[0.3em] text-slate-400">已选知识库</p>
          <p class="mt-3 font-serif text-[1.7rem] leading-tight text-slate-900">
            {{ chatUiStore.preferredKnowledgeBaseName || '未选择知识库' }}
          </p>
        </SurfaceCard>
      </div>
    </div>
  </div>
</template>
