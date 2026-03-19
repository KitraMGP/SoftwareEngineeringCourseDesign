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

const workflowSteps = [
  {
    step: '步骤 1',
    title: '创建会话',
    description: '先开启一个新的工作台，让本轮问答有清晰的上下文起点。'
  },
  {
    step: '步骤 2',
    title: '选择知识库',
    description: '按当前任务绑定合适资料，让后续回答优先参考你的私有内容。'
  },
  {
    step: '步骤 3',
    title: '开始问答',
    description: '进入会话后直接提问，围绕同一主题持续追问和展开。'
  }
] as const;

const sessionsQuery = useQuery({
  queryKey: computed(() =>
    scopeQueryKey(queryKeys.sessions({ page: 1, size: 20 }), userScope.value)
  ),
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
      <section
        class="rounded-[28px] border border-white/68 bg-white/76 px-6 py-6 shadow-soft backdrop-blur"
      >
        <div class="flex flex-col gap-5 lg:flex-row lg:items-end lg:justify-between">
          <div class="max-w-3xl space-y-3">
            <p class="text-xs uppercase tracking-[0.36em] text-slate-400">
              Conversation workspace
            </p>
            <h1 class="font-serif text-3xl text-slate-900 lg:text-[2.6rem]">
              {{ PRODUCT_NAME }}
            </h1>
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
        <SurfaceCard tone="soft" class="flex flex-col">
          <div class="space-y-4">
            <p class="text-xs uppercase tracking-[0.32em] text-slate-400">Start here</p>
            <div class="space-y-3">
              <h2 class="font-serif text-2xl text-slate-900 lg:text-[2rem]">
                从这里开始新的工作流
              </h2>
              <p class="max-w-2xl text-sm leading-7 text-slate-600">
                按照下面三个步骤组织一次新的问答流程，先搭好会话与知识上下文，再开始正式提问。
              </p>
            </div>
          </div>

          <div class="mt-8 space-y-4">
            <article
              v-for="item in workflowSteps"
              :key="item.step"
              class="rounded-[22px] border border-white/80 bg-white/84 px-5 py-4 shadow-[0_18px_40px_rgba(148,163,184,0.12)]"
            >
              <div class="space-y-2">
                <p class="text-xs uppercase tracking-[0.28em] text-slate-400">
                  {{ item.step }}
                </p>
                <h3 class="font-serif text-xl text-slate-900">{{ item.title }}</h3>
                <p class="max-w-xl text-sm leading-7 text-slate-600">
                  {{ item.description }}
                </p>
              </div>
            </article>
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
