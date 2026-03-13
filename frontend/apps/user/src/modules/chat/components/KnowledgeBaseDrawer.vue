<script setup lang="ts">
import { computed } from 'vue';

import type { KnowledgeBase } from '@private-kb/shared';
import { AppStatusBadge, formatDateTime, shortenText } from '@private-kb/shared';

const props = withDefaults(
  defineProps<{
    modelValue: boolean;
    knowledgeBases: KnowledgeBase[];
    loading?: boolean;
    selectedId?: string | null;
  }>(),
  {
    loading: false,
    selectedId: null
  }
);

const emit = defineEmits<{
  'update:modelValue': [value: boolean];
  select: [{ id: string; name: string }];
  clear: [];
  'create-session': [knowledgeBaseId?: string];
}>();

const selectedKnowledgeBase = computed(() =>
  props.knowledgeBases.find((item) => item.id === props.selectedId)
);

function handleSelect(item: KnowledgeBase) {
  emit('select', { id: item.id, name: item.name });
}
</script>

<template>
  <el-drawer
    :model-value="modelValue"
    size="480px"
    direction="rtl"
    :with-header="false"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div class="flex h-full flex-col">
      <div class="space-y-3 pb-5">
        <p class="text-xs uppercase tracking-[0.34em] text-slate-400">Knowledge context</p>
        <h2 class="font-serif text-3xl text-slate-900">选择知识库</h2>
        <p class="text-sm leading-6 text-slate-500">
          为了保持上下文清晰，切换知识库会从新会话开始。你可以先挑选资料库，再开启下一轮对话。
        </p>
      </div>

      <div
        v-if="selectedKnowledgeBase"
        class="mb-5 rounded-[24px] border border-slate-200 bg-slate-900 px-5 py-4 text-white"
      >
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-white/55">已准备</p>
            <p class="mt-2 text-lg font-semibold">{{ selectedKnowledgeBase.name }}</p>
            <p class="mt-2 text-sm leading-6 text-white/72">
              {{ shortenText(selectedKnowledgeBase.description || '将以该知识库作为下一次会话的上下文来源。', 96) }}
            </p>
          </div>
          <button
            type="button"
            class="text-sm text-white/72 transition hover:text-white"
            @click="emit('clear')"
          >
            清空
          </button>
        </div>
      </div>

      <div class="soft-scrollbar flex-1 space-y-3 overflow-y-auto pr-1">
        <div
          v-for="item in knowledgeBases"
          :key="item.id"
          class="cursor-pointer rounded-[24px] border px-5 py-4 transition"
          :class="
            selectedId === item.id
              ? 'border-slate-900 bg-slate-900 text-white shadow-soft'
              : 'border-slate-200 bg-white hover:border-slate-300'
          "
          @click="handleSelect(item)"
        >
          <div class="flex items-start justify-between gap-4">
            <div class="space-y-2">
              <p class="text-lg font-semibold">
                {{ item.name }}
              </p>
              <p
                class="text-sm leading-6"
                :class="selectedId === item.id ? 'text-white/72' : 'text-slate-500'"
              >
                {{ shortenText(item.description || '当前知识库尚未补充摘要说明。', 92) }}
              </p>
            </div>
            <AppStatusBadge :tone="item.last_indexed_at ? 'success' : 'info'" :dot="false">
              {{ item.last_indexed_at ? '已建索引' : '待整理' }}
            </AppStatusBadge>
          </div>
          <div
            class="mt-4 flex flex-wrap items-center gap-2 text-xs"
            :class="selectedId === item.id ? 'text-white/65' : 'text-slate-400'"
          >
            <span>Embedding: {{ item.embedding_model }}</span>
            <span>Top K: {{ item.retrieval_top_k }}</span>
            <span>更新于 {{ formatDateTime(item.updated_at) }}</span>
          </div>
        </div>

        <div
          v-if="!loading && knowledgeBases.length === 0"
          class="rounded-[24px] border border-dashed border-slate-200 bg-white px-5 py-8 text-sm leading-6 text-slate-500"
        >
          还没有可用知识库。你可以先前往知识库页面创建资料库，再回来开始新会话。
        </div>
      </div>

      <div class="mt-5 space-y-3 border-t border-slate-200 pt-5">
        <button
          type="button"
          class="flex w-full items-center justify-center rounded-[18px] bg-slate-900 px-4 py-3.5 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-300"
          :disabled="loading"
          @click="emit('create-session', selectedId || undefined)"
        >
          {{ selectedId ? '基于当前知识库开启新会话' : '直接开启空会话' }}
        </button>
        <p class="text-xs leading-5 text-slate-400">
          如果你在下一次会话前仍未选定资料库，系统会以普通空会话方式创建。
        </p>
      </div>
    </div>
  </el-drawer>
</template>
