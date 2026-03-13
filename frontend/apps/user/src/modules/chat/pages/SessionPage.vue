<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query';
import { useQueryClient } from '@tanstack/vue-query';
import { ArrowRight } from '@element-plus/icons-vue';
import { ElMessage } from 'element-plus';
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import {
  AppStatusBadge,
  DEFAULT_CHAT_MODEL,
  EmptyStatePanel,
  SurfaceCard,
  chatApi,
  formatDateTime,
  queryKeys,
  useAuthStore
} from '@private-kb/shared';
import { getErrorMessage } from '@private-kb/shared/utils/errors';

import MessageComposer from '../components/MessageComposer.vue';
import MessageThread from '../components/MessageThread.vue';
import { buildPreviewConversation } from '../mock';
import type { UiChatMessage } from '../types';
import { useChatUiStore } from '../../../stores/useChatUiStore';

const route = useRoute();
const queryClient = useQueryClient();
const authStore = useAuthStore();
const chatUiStore = useChatUiStore();

const draftInput = ref('');
const transientMessages = ref<UiChatMessage[]>([]);
const isStreaming = ref(false);
const streamAbortController = ref<AbortController | null>(null);

const sessionId = computed(() => String(route.params.sessionId || ''));

const sessionQuery = useQuery({
  queryKey: computed(() => queryKeys.session(sessionId.value)),
  queryFn: () => chatApi.getSessionDetail(sessionId.value),
  enabled: computed(() => !!sessionId.value)
});

const session = computed(() => sessionQuery.data.value?.session || null);
const canSendMessage = computed(() => !!session.value && !session.value.knowledge_base_id);

const persistedMessages = computed<UiChatMessage[]>(() =>
  (sessionQuery.data.value?.messages ?? []).map((item) => ({
    id: item.id,
    role: item.role,
    content: item.content,
    grounded: item.grounded,
    createdAt: item.created_at,
    tag:
      item.role === 'assistant' && session.value?.knowledge_base_id && !item.grounded
        ? '未命中知识库'
        : undefined
  }))
);

const displayMessages = computed<UiChatMessage[]>(() => {
  if (persistedMessages.value.length) {
    return [...persistedMessages.value, ...transientMessages.value];
  }

  if (transientMessages.value.length) {
    return transientMessages.value;
  }

  return buildPreviewConversation(session.value || undefined);
});

const capabilityTone = computed<'success' | 'warning'>(() =>
  canSendMessage.value ? 'success' : 'warning'
);

const capabilityLabel = computed(() =>
  canSendMessage.value ? '实时问答已接通' : '知识库问答待开放'
);

const composerHint = computed(() => {
  if (!session.value) {
    return '';
  }

  if (session.value.knowledge_base_id) {
    return '当前绑定知识库的会话仍等待检索联调。若要体验实时问答，请新建一个未绑定知识库的空会话。';
  }

  if (isStreaming.value) {
    return 'Assistant 正在返回内容。当前阶段尚未接入停止生成，请等待这一轮回答完成。';
  }

  return '当前会话已接入 DeepSeek SSE，可直接提问并查看消息持久化结果。';
});

const composerPlaceholder = computed(() =>
  canSendMessage.value
    ? '输入你的问题，按 Ctrl/Command + Enter 发送。'
    : '当前绑定知识库的会话暂未开放发送，可改为新建空会话体验实时对话。'
);

const composerSubmitLabel = computed(() => (isStreaming.value ? '生成中...' : '发送消息'));

async function refetchSessionState() {
  await queryClient.invalidateQueries({ queryKey: queryKeys.sessionsRoot });
  await sessionQuery.refetch();
}

async function handleSubmit() {
  const content = draftInput.value.trim();

  if (!content || !session.value || !sessionId.value || isStreaming.value) {
    return;
  }

  if (!canSendMessage.value) {
    ElMessage.warning('当前只有未绑定知识库的空会话支持实时问答。');
    return;
  }

  if (!authStore.accessToken) {
    ElMessage.error('登录状态已失效，请重新登录后重试。');
    return;
  }

  const createdAt = new Date().toISOString();
  const localUserMessageId = `local-user-${Date.now()}`;
  let activeAssistantMessageId = `local-assistant-${Date.now()}`;

  transientMessages.value = [
    ...transientMessages.value,
    {
      id: localUserMessageId,
      role: 'user',
      content,
      createdAt
    },
    {
      id: activeAssistantMessageId,
      role: 'assistant',
      content: '',
      createdAt,
      tag: '生成中',
      isStreaming: true
    }
  ];

  draftInput.value = '';
  isStreaming.value = true;

  streamAbortController.value?.abort();
  const controller = new AbortController();
  streamAbortController.value = controller;

  const patchAssistantMessage = (updater: (message: UiChatMessage) => UiChatMessage) => {
    transientMessages.value = transientMessages.value.map((message) =>
      message.id === activeAssistantMessageId ? updater(message) : message
    );
  };

  try {
    await chatApi.streamSessionMessage(
      sessionId.value,
      { content },
      {
        accessToken: authStore.accessToken,
        signal: controller.signal,
        refreshAccessToken: () => authStore.refreshAccessToken(),
        onUnauthorized: () => authStore.clearAuth(),
        onMeta: (payload) => {
          transientMessages.value = transientMessages.value.map((message) =>
            message.id === activeAssistantMessageId
              ? {
                  ...message,
                  id: payload.message_id,
                  grounded: payload.grounded
                }
              : message
          );

          activeAssistantMessageId = payload.message_id;
        },
        onDelta: (payload) => {
          patchAssistantMessage((message) => ({
            ...message,
            content: `${message.content}${payload.content}`
          }));
        }
      }
    );

    await refetchSessionState();
    transientMessages.value = [];
  } catch (error) {
    if (controller.signal.aborted) {
      return;
    }

    await refetchSessionState();
    transientMessages.value = [];
    ElMessage.error(getErrorMessage(error, '消息发送失败，请稍后重试。'));
  } finally {
    if (streamAbortController.value === controller) {
      streamAbortController.value = null;
    }
    isStreaming.value = false;
  }
}

watch(
  sessionId,
  (nextSessionId, previousSessionId) => {
    if (previousSessionId && nextSessionId !== previousSessionId) {
      streamAbortController.value?.abort();
      streamAbortController.value = null;
      transientMessages.value = [];
      isStreaming.value = false;
      draftInput.value = '';
    }
  }
);

onBeforeUnmount(() => {
  streamAbortController.value?.abort();
});
</script>

<template>
  <div class="flex h-full min-h-0 flex-1 flex-col">
    <template v-if="sessionQuery.isError.value">
      <div class="flex h-full items-center justify-center px-5 py-10 lg:px-8">
        <div class="w-full max-w-3xl">
          <EmptyStatePanel
            title="没有找到这条会话"
            description="会话可能已经被删除，或者当前链接已失效。你可以从左侧重新选择其他会话。"
          />
        </div>
      </div>
    </template>

    <template v-else>
      <div class="shrink-0 border-b border-slate-200/70 px-5 py-5 lg:px-8">
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p class="text-xs uppercase tracking-[0.32em] text-slate-400">Session overview</p>
            <h1 class="mt-2 font-serif text-3xl text-slate-900">
              {{ session?.name || '未命名会话' }}
            </h1>
            <p class="mt-2 text-sm leading-6 text-slate-500">
              创建于 {{ formatDateTime(session?.created_at) }}，模型标识为 {{ session?.model || DEFAULT_CHAT_MODEL }}。
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <AppStatusBadge :tone="session?.knowledge_base_id ? 'success' : 'info'">
              {{ session?.knowledge_base_name || '未绑定知识库' }}
            </AppStatusBadge>
            <AppStatusBadge :tone="capabilityTone">
              {{ capabilityLabel }}
            </AppStatusBadge>
            <button
              type="button"
              class="rounded-full border border-slate-200 bg-white px-4 py-2.5 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
              @click="chatUiStore.openKnowledgeBaseDrawer()"
            >
              {{ session?.knowledge_base_id ? '换一个知识库开始新会话' : '为下一次会话选择知识库' }}
            </button>
          </div>
        </div>

        <SurfaceCard class="mt-5" :padded="false">
          <div class="flex flex-wrap items-center gap-3 px-5 py-4 text-sm text-slate-500">
            <span>当前会话会保留结构和历史记录。</span>
            <span class="hidden text-slate-300 md:inline">/</span>
            <span>切换知识库将从新会话开始，避免上下文混杂。</span>
            <span class="hidden text-slate-300 md:inline">/</span>
            <router-link
              v-if="session?.knowledge_base_id"
              :to="{ name: 'knowledge-base-detail', params: { kbId: session.knowledge_base_id } }"
              class="inline-flex items-center gap-1 text-slate-900 hover:text-brand-night"
            >
              查看当前知识库
              <el-icon><ArrowRight /></el-icon>
            </router-link>
          </div>
        </SurfaceCard>
      </div>

      <MessageThread :messages="displayMessages" />

      <MessageComposer
        v-model="draftInput"
        :disabled="!canSendMessage || isStreaming"
        :hint="composerHint"
        :placeholder="composerPlaceholder"
        :submit-label="composerSubmitLabel"
        @submit="handleSubmit"
      />
    </template>
  </div>
</template>
