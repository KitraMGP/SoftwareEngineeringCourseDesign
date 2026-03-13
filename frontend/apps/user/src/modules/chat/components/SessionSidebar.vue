<script setup lang="ts">
import { Collection, Expand, Fold, Lock, Plus, SwitchButton } from '@element-plus/icons-vue';

import type { Session } from '@private-kb/shared';
import { UserAvatar, formatRelativeTime, shortenText } from '@private-kb/shared';

withDefaults(
  defineProps<{
    sessions: Session[];
    activeSessionId?: string;
    loading?: boolean;
    collapsed?: boolean;
    displayName: string;
    avatarUrl?: string | null;
    roleLabel: string;
    preferredKnowledgeBaseName?: string | null;
    isCreating?: boolean;
  }>(),
  {
    activeSessionId: '',
    loading: false,
    collapsed: false,
    avatarUrl: null,
    preferredKnowledgeBaseName: null,
    isCreating: false
  }
);

const emit = defineEmits<{
  'toggle-collapse': [];
  'create-session': [knowledgeBaseId?: string];
  'delete-session': [sessionId: string];
  'open-knowledge-base-drawer': [];
  logout: [];
}>();

function sessionTitle(session: Session): string {
  return session.name || '未命名会话';
}
</script>

<template>
  <aside
    class="flex h-full w-[88vw] max-w-[320px] flex-col rounded-[32px] border border-white/60 bg-white/72 p-4 shadow-frost backdrop-blur-2xl lg:h-[calc(100vh-3rem)] lg:w-auto"
    :class="collapsed ? 'lg:w-[108px]' : 'lg:w-[300px]'"
  >
    <div class="flex items-center justify-between gap-3">
      <div v-if="!collapsed" class="space-y-1">
        <p class="text-xs uppercase tracking-[0.34em] text-slate-400">Workspace</p>
        <h2 class="font-serif text-2xl text-slate-900">对话台</h2>
      </div>
      <button
        v-if="$attrs.onToggleCollapse"
        type="button"
        class="hidden h-11 w-11 items-center justify-center rounded-full border border-slate-200/80 bg-white/85 text-slate-700 lg:flex"
        @click="emit('toggle-collapse')"
      >
        <el-icon :size="18">
          <component :is="collapsed ? Expand : Fold" />
        </el-icon>
      </button>
    </div>

    <button
      type="button"
      class="mt-5 flex items-center justify-center gap-2 rounded-[20px] bg-slate-900 px-4 py-3 text-sm font-semibold text-white transition hover:-translate-y-0.5 hover:bg-slate-800"
      @click="emit('create-session')"
    >
      <el-icon><Plus /></el-icon>
      <span v-if="!collapsed">{{ isCreating ? '创建中...' : '新建会话' }}</span>
    </button>

    <button
      type="button"
      class="mt-3 flex items-center justify-center gap-2 rounded-[18px] border border-slate-200 bg-white/90 px-4 py-3 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-white"
      @click="emit('open-knowledge-base-drawer')"
    >
      <el-icon><Collection /></el-icon>
      <span v-if="!collapsed">准备知识库</span>
    </button>

    <div v-if="!collapsed && preferredKnowledgeBaseName" class="mt-3 rounded-[18px] bg-slate-900/5 px-4 py-3">
      <p class="text-xs uppercase tracking-[0.3em] text-slate-400">下一次会话</p>
      <p class="mt-2 text-sm font-medium text-slate-800">
        将绑定到「{{ preferredKnowledgeBaseName }}」
      </p>
    </div>

    <div class="mt-5 min-h-0 flex-1 overflow-hidden">
      <div v-if="!collapsed" class="mb-3 flex items-center justify-between">
        <p class="text-xs uppercase tracking-[0.3em] text-slate-400">最近会话</p>
        <span class="text-xs text-slate-400">{{ sessions.length }} 条</span>
      </div>

      <div class="soft-scrollbar flex h-full flex-col gap-3 overflow-y-auto pr-1">
        <div v-if="loading" class="space-y-3">
          <div
            v-for="item in 4"
            :key="item"
            class="h-20 animate-pulse rounded-[22px] bg-slate-100/80"
          />
        </div>

        <router-link
          v-for="session in sessions"
          :key="session.id"
          :to="{ name: 'session-detail', params: { sessionId: session.id } }"
          class="group rounded-[24px] border px-4 py-4 transition"
          :class="
            activeSessionId === session.id
              ? 'border-slate-900/10 bg-slate-900 text-white shadow-soft'
              : 'border-white/60 bg-white/85 text-slate-900 hover:border-slate-200 hover:bg-white'
          "
        >
          <div class="flex items-start justify-between gap-3">
            <div v-if="!collapsed" class="min-w-0 flex-1">
              <p class="truncate text-sm font-semibold">
                {{ sessionTitle(session) }}
              </p>
              <p
                class="mt-2 line-clamp-2 text-xs leading-5"
                :class="activeSessionId === session.id ? 'text-white/78' : 'text-slate-500'"
              >
                {{ shortenText(session.knowledge_base_name || '未绑定知识库，将按空会话上下文展示。', 56) }}
              </p>
              <div class="mt-3 flex items-center justify-between gap-2">
                <span
                  class="rounded-full px-2.5 py-1 text-[11px]"
                  :class="
                    activeSessionId === session.id
                      ? 'bg-white/16 text-white/88'
                      : 'bg-slate-100 text-slate-500'
                  "
                >
                  {{ formatRelativeTime(session.updated_at) }}
                </span>
                <button
                  type="button"
                  class="text-[11px] font-medium"
                  :class="activeSessionId === session.id ? 'text-white/84' : 'text-slate-400'"
                  @click.prevent.stop="emit('delete-session', session.id)"
                >
                  删除
                </button>
              </div>
            </div>

            <div
              v-else
              class="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-slate-900/8 text-xs font-semibold"
            >
              {{ sessionTitle(session).slice(0, 1) }}
            </div>
          </div>
        </router-link>

        <div
          v-if="!loading && sessions.length === 0 && !collapsed"
          class="rounded-[24px] border border-dashed border-slate-200 bg-white/80 px-4 py-6 text-sm leading-6 text-slate-500"
        >
          当前还没有会话。你可以直接创建空会话，或先选择知识库后再开始新对话。
        </div>
      </div>
    </div>

    <div class="mt-5 space-y-2 border-t border-slate-200/70 pt-4">
      <router-link
        class="flex items-center gap-3 rounded-[18px] px-4 py-3 text-sm font-medium text-slate-700 transition hover:bg-white"
        :to="{ name: 'knowledge-bases' }"
      >
        <el-icon><Collection /></el-icon>
        <span v-if="!collapsed">知识库</span>
      </router-link>
      <router-link
        class="flex items-center gap-3 rounded-[18px] px-4 py-3 text-sm font-medium text-slate-700 transition hover:bg-white"
        :to="{ name: 'about' }"
      >
        <span class="flex h-5 w-5 items-center justify-center rounded-full bg-slate-900 text-[11px] text-white">
          i
        </span>
        <span v-if="!collapsed">关于</span>
      </router-link>
      <router-link
        class="flex items-center gap-3 rounded-[18px] px-4 py-3 text-sm font-medium text-slate-700 transition hover:bg-white"
        :to="{ name: 'me-security' }"
      >
        <el-icon><Lock /></el-icon>
        <span v-if="!collapsed">安全设置</span>
      </router-link>

      <router-link
        class="mt-3 flex items-center gap-3 rounded-[22px] border border-white/60 bg-white/90 px-4 py-3 text-slate-700 transition hover:bg-white"
        :to="{ name: 'me-profile' }"
      >
        <UserAvatar :label="displayName" :avatar-url="avatarUrl" :size="44" />
        <div v-if="!collapsed" class="min-w-0">
          <p class="truncate text-sm font-semibold text-slate-900">{{ displayName }}</p>
          <p class="text-xs text-slate-400">{{ roleLabel }}</p>
        </div>
      </router-link>

      <button
        type="button"
        class="flex items-center gap-3 rounded-[18px] px-4 py-3 text-left text-sm font-medium text-rose-600 transition hover:bg-rose-50"
        @click="emit('logout')"
      >
        <el-icon><SwitchButton /></el-icon>
        <span v-if="!collapsed">退出登录</span>
      </button>
    </div>
  </aside>
</template>
