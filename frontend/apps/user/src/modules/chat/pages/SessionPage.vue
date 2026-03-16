<script setup lang="ts">
import { useQuery, useQueryClient } from '@tanstack/vue-query';
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
  scopeQueryKey,
  useAuthStore
} from '@private-kb/shared';
import type { Message } from '@private-kb/shared';
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
const userScope = computed(() => authStore.user?.id ?? null);

interface ActiveStreamState {
  kind: 'send' | 'regenerate';
  userMessage?: UiChatMessage;
  assistantMessage: UiChatMessage;
}

type StreamRequestOptions = NonNullable<Parameters<typeof chatApi.streamSessionMessage>[2]>;

const draftInput = ref('');
const activeStream = ref<ActiveStreamState | null>(null);
const isStopPending = ref(false);
const streamAbortController = ref<AbortController | null>(null);

const sessionId = computed(() => String(route.params.sessionId || ''));

const sessionQuery = useQuery({
  queryKey: computed(() => scopeQueryKey(queryKeys.session(sessionId.value), userScope.value)),
  queryFn: () => chatApi.getSessionDetail(sessionId.value),
  enabled: computed(() => !!userScope.value && !!sessionId.value)
});

const session = computed(() => sessionQuery.data.value?.session || null);
const canSendMessage = computed(() => !!session.value);
const isKnowledgeBoundSession = computed(() => !!session.value?.knowledge_base_id);
const isStreaming = computed(() => !!activeStream.value);
const regeneratingMessageId = computed(() =>
  activeStream.value?.kind === 'regenerate' ? activeStream.value.assistantMessage.id : null
);

function getStreamingTag(kind: ActiveStreamState['kind'], stopPending = false) {
  if (stopPending) {
    return '正在停止';
  }

  return kind === 'regenerate' ? '重新生成中' : '生成中';
}

function mapCitation(message: Message): UiChatMessage['citations'] {
  if (message.role !== 'assistant' || !message.citations?.length) {
    return [];
  }

  return message.citations.map((citation) => ({
    id: citation.id,
    title: citation.document_name,
    documentId: citation.document_id,
    knowledgeBaseId: citation.knowledge_base_id,
    rank: citation.rank_no,
    sourcePage: citation.source_page ?? null
  }));
}

function getAssistantTag(message: Message): Pick<UiChatMessage, 'tag' | 'tagTone'> {
  if (message.role !== 'assistant' || !session.value?.knowledge_base_id) {
    return {};
  }

  if (message.grounded) {
    return {
      tag: '命中知识库',
      tagTone: 'success'
    };
  }

  return {
    tag: '未命中知识库',
    tagTone: 'warning'
  };
}

function toUiMessage(message: Message): UiChatMessage {
  return {
    id: message.id,
    role: message.role,
    content: message.content,
    grounded: message.grounded,
    createdAt: message.created_at,
    citations: mapCitation(message),
    canRegenerate: message.role === 'assistant' && message.status === 'completed',
    ...getAssistantTag(message)
  };
}

function patchActiveAssistantMessage(updater: (message: UiChatMessage) => UiChatMessage) {
  if (!activeStream.value) {
    return;
  }

  activeStream.value = {
    ...activeStream.value,
    assistantMessage: updater(activeStream.value.assistantMessage)
  };
}

const persistedMessages = computed<UiChatMessage[]>(() =>
  (sessionQuery.data.value?.messages ?? []).map((item) => toUiMessage(item))
);

const displayMessages = computed<UiChatMessage[]>(() => {
  if (persistedMessages.value.length) {
    if (activeStream.value?.kind === 'send' && activeStream.value.userMessage) {
      return [
        ...persistedMessages.value,
        activeStream.value.userMessage,
        activeStream.value.assistantMessage
      ];
    }

    if (activeStream.value?.kind === 'regenerate') {
      return persistedMessages.value.map((message) =>
        message.id === activeStream.value?.assistantMessage.id
          ? activeStream.value.assistantMessage
          : message
      );
    }

    return persistedMessages.value;
  }

  if (activeStream.value?.kind === 'send' && activeStream.value.userMessage) {
    return [activeStream.value.userMessage, activeStream.value.assistantMessage];
  }

  return buildPreviewConversation(session.value || undefined);
});

const composerHint = computed(() => {
  if (!session.value) {
    return '';
  }

  if (isStreaming.value) {
    return isStopPending.value
      ? '正在请求停止本轮生成，请稍候。'
      : 'Assistant 正在返回内容，你可以点击“停止生成”中断本轮回答。';
  }

  return '';
});

const composerPlaceholder = computed(() => '输入问题。Enter 发送，Shift+Enter 换行。');

const composerSubmitLabel = computed(() => '发送');

async function refetchSessionState() {
  await queryClient.invalidateQueries({ queryKey: queryKeys.sessionsRoot });
  await sessionQuery.refetch();
}

async function executeStream(
  request: (streamOptions: StreamRequestOptions) => Promise<void>,
  fallbackMessage: string
) {
  streamAbortController.value?.abort();
  const controller = new AbortController();
  streamAbortController.value = controller;
  isStopPending.value = false;

  let finishReason = '';

  const streamOptions: StreamRequestOptions = {
    accessToken: authStore.accessToken,
    signal: controller.signal,
    refreshAccessToken: () => authStore.refreshAccessToken(),
    onUnauthorized: () => authStore.clearAuth(),
    onMeta: (payload) => {
      patchActiveAssistantMessage((message) => ({
        ...message,
        id: payload.message_id,
        grounded: payload.grounded
      }));
    },
    onDelta: (payload) => {
      patchActiveAssistantMessage((message) => ({
        ...message,
        content: `${message.content}${payload.content}`
      }));
    },
    onDone: (payload) => {
      finishReason = payload.finish_reason;
    }
  };

  try {
    await request(streamOptions);
    await refetchSessionState();

    if (finishReason === 'cancelled') {
      ElMessage.info('已停止本轮生成。');
    }
  } catch (error) {
    if (controller.signal.aborted) {
      return;
    }

    await refetchSessionState();
    ElMessage.error(getErrorMessage(error, fallbackMessage));
  } finally {
    if (streamAbortController.value === controller) {
      streamAbortController.value = null;
    }
    activeStream.value = null;
    isStopPending.value = false;
  }
}

async function handleSubmit() {
  const content = draftInput.value.trim();

  if (!content || !session.value || !sessionId.value || isStreaming.value) {
    return;
  }

  if (!authStore.accessToken) {
    ElMessage.error('登录状态已失效，请重新登录后重试。');
    return;
  }

  const createdAt = new Date().toISOString();
  activeStream.value = {
    kind: 'send',
    userMessage: {
      id: `local-user-${Date.now()}`,
      role: 'user',
      content,
      createdAt
    },
    assistantMessage: {
      id: `local-assistant-${Date.now()}`,
      role: 'assistant',
      content: '',
      createdAt,
      tag: getStreamingTag('send'),
      tagTone: 'info',
      citations: [],
      isStreaming: true,
      canRegenerate: false
    }
  };

  draftInput.value = '';
  await executeStream(
    (streamOptions) =>
      chatApi.streamSessionMessage(sessionId.value, { content }, streamOptions),
    '消息发送失败，请稍后重试。'
  );
}

async function handleRegenerate(messageId: string) {
  if (!session.value || !sessionId.value || isStreaming.value) {
    return;
  }

  if (!authStore.accessToken) {
    ElMessage.error('登录状态已失效，请重新登录后重试。');
    return;
  }

  const targetMessage = persistedMessages.value.find(
    (message) => message.id === messageId && message.role === 'assistant'
  );

  if (!targetMessage) {
    ElMessage.warning('没有找到可重新生成的回答。');
    return;
  }

  activeStream.value = {
    kind: 'regenerate',
    assistantMessage: {
      ...targetMessage,
      content: '',
      citations: [],
      tag: getStreamingTag('regenerate'),
      tagTone: 'info',
      isStreaming: true,
      canRegenerate: false
    }
  };

  await executeStream(
    (streamOptions) =>
      chatApi.regenerateSessionMessage(sessionId.value, messageId, streamOptions),
    '重新生成失败，请稍后重试。'
  );
}

async function handleStopStream() {
  if (!sessionId.value || !isStreaming.value || isStopPending.value || !activeStream.value) {
    return;
  }

  isStopPending.value = true;
  patchActiveAssistantMessage((message) => ({
    ...message,
    tag: getStreamingTag(activeStream.value?.kind ?? 'send', true)
  }));

  try {
    const result = await chatApi.stopSessionStream(sessionId.value);

    if (!result.stopped) {
      isStopPending.value = false;
      patchActiveAssistantMessage((message) => ({
        ...message,
        tag: getStreamingTag(activeStream.value?.kind ?? 'send')
      }));
      ElMessage.info('当前会话没有正在生成的内容。');
    }
  } catch (error) {
    isStopPending.value = false;
    patchActiveAssistantMessage((message) => ({
      ...message,
      tag: getStreamingTag(activeStream.value?.kind ?? 'send')
    }));
    ElMessage.error(getErrorMessage(error, '停止生成失败，请稍后重试。'));
  }
}

watch(
  sessionId,
  (nextSessionId, previousSessionId) => {
    if (previousSessionId && nextSessionId !== previousSessionId) {
      streamAbortController.value?.abort();
      streamAbortController.value = null;
      activeStream.value = null;
      isStopPending.value = false;
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
      <div class="shrink-0 border-b border-slate-200/70 px-5 py-4 lg:px-8">
        <div class="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
          <div class="min-w-0">
            <p class="text-xs uppercase tracking-[0.32em] text-slate-400">Session overview</p>
            <div class="mt-2 flex flex-wrap items-center gap-2">
              <h1 class="font-serif text-[2rem] leading-tight text-slate-900">
                {{ session?.name || '未命名会话' }}
              </h1>
              <AppStatusBadge :tone="session?.knowledge_base_id ? 'success' : 'info'">
                {{ session?.knowledge_base_name || '未绑定知识库' }}
              </AppStatusBadge>
            </div>
            <p class="mt-2 text-sm leading-6 text-slate-500">
              创建于 {{ formatDateTime(session?.created_at) }}，模型标识为 {{ session?.model || DEFAULT_CHAT_MODEL }}。
            </p>
          </div>

          <div class="flex shrink-0 flex-wrap items-center gap-2">
            <button
              type="button"
              class="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
              @click="chatUiStore.openKnowledgeBaseDrawer()"
            >
              {{ session?.knowledge_base_id ? '换一个知识库开始新会话' : '为下一次会话选择知识库' }}
            </button>
          </div>
        </div>
      </div>

      <MessageThread
        :messages="displayMessages"
        :streaming="isStreaming"
        :regenerating-message-id="regeneratingMessageId"
        @regenerate="handleRegenerate"
      />

      <MessageComposer
        v-model="draftInput"
        :disabled="!canSendMessage || isStreaming"
        :hint="composerHint"
        :placeholder="composerPlaceholder"
        :submit-label="composerSubmitLabel"
        :show-stop-action="isStreaming"
        :stop-pending="isStopPending"
        @submit="handleSubmit"
        @stop="handleStopStream"
      />
    </template>
  </div>
</template>
