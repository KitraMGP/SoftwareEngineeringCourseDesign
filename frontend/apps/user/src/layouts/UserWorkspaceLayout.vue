<script setup lang="ts">
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Menu } from '@element-plus/icons-vue';
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { chatApi, knowledgeBaseApi } from '@private-kb/shared';
import { DEFAULT_CHAT_MODEL, queryKeys } from '@private-kb/shared';
import { useAuthStore } from '@private-kb/shared/auth/useAuthStore';
import { getErrorMessage } from '@private-kb/shared/utils/errors';

import SessionSidebar from '../modules/chat/components/SessionSidebar.vue';
import KnowledgeBaseDrawer from '../modules/chat/components/KnowledgeBaseDrawer.vue';
import { useChatUiStore } from '../stores/useChatUiStore';
import { useUserAppStore } from '../stores/useUserAppStore';

const router = useRouter();
const route = useRoute();
const queryClient = useQueryClient();

const authStore = useAuthStore();
const chatUiStore = useChatUiStore();
const userAppStore = useUserAppStore();

const sessionsQuery = useQuery({
  queryKey: queryKeys.sessions({ page: 1, size: 20 }),
  queryFn: () => chatApi.listSessions({ page: 1, size: 20 })
});

const knowledgeBasesQuery = useQuery({
  queryKey: queryKeys.knowledgeBases({ page: 1, size: 50 }),
  queryFn: () => knowledgeBaseApi.listKnowledgeBases({ page: 1, size: 50 })
});

const activeSessionId = computed(() =>
  typeof route.params.sessionId === 'string' ? route.params.sessionId : ''
);

const createSessionMutation = useMutation({
  mutationFn: (knowledgeBaseId?: string) =>
    chatApi.createSession({
      model: DEFAULT_CHAT_MODEL,
      knowledge_base_id: knowledgeBaseId
    }),
  onSuccess: async (result, knowledgeBaseId) => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.sessionsRoot });

    if (knowledgeBaseId) {
      const selectedKnowledgeBase = knowledgeBasesQuery.data.value?.items.find(
        (item) => item.id === knowledgeBaseId
      );
      if (selectedKnowledgeBase) {
        chatUiStore.selectKnowledgeBase(selectedKnowledgeBase.id, selectedKnowledgeBase.name);
      }
    }

    chatUiStore.closeKnowledgeBaseDrawer();
    userAppStore.setMobileSidebarOpen(false);
    router.push({ name: 'session-detail', params: { sessionId: result.session_id } });
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

const deleteSessionMutation = useMutation({
  mutationFn: (sessionId: string) => chatApi.deleteSession(sessionId),
  onSuccess: async (_, sessionId) => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.sessionsRoot });

    if (sessionId === activeSessionId.value) {
      router.push({ name: 'chat-home' });
    }

    ElMessage.success('会话已移出工作区。');
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

const roleLabel = computed(() => (authStore.user?.role === 'admin' ? '管理员' : '成员'));

function handleCreateSession(knowledgeBaseId?: string) {
  createSessionMutation.mutate(knowledgeBaseId || chatUiStore.preferredKnowledgeBaseId || undefined);
}

async function handleLogout() {
  try {
    await authStore.logout();
  } catch {
    // Auth state is cleared in the store finally block.
  } finally {
    queryClient.clear();
    chatUiStore.clearKnowledgeBase();
    userAppStore.setMobileSidebarOpen(false);
    ElMessage.success('已退出登录。');
    await router.push({ name: 'login' });
  }
}

async function handleDeleteSession(sessionId: string) {
  try {
    await ElMessageBox.confirm('删除后会话记录将不可恢复，是否继续？', '删除会话', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    });
  } catch {
    return;
  }

  deleteSessionMutation.mutate(sessionId);
}
</script>

<template>
  <div class="relative min-h-screen overflow-hidden">
    <div class="user-app-backdrop" />

    <div class="relative z-10 mx-auto flex min-h-screen max-w-[1600px] gap-5 px-4 py-4 lg:px-6 lg:py-6">
      <div class="hidden lg:block">
        <SessionSidebar
          :sessions="sessionsQuery.data.value?.items ?? []"
          :active-session-id="activeSessionId"
          :loading="sessionsQuery.isLoading.value"
          :collapsed="userAppStore.sidebarCollapsed"
          :display-name="authStore.displayName"
          :avatar-url="authStore.user?.avatar_url"
          :role-label="roleLabel"
          :preferred-knowledge-base-name="chatUiStore.preferredKnowledgeBaseName"
          :is-creating="createSessionMutation.isPending.value"
          @toggle-collapse="userAppStore.toggleSidebar"
          @create-session="handleCreateSession"
          @delete-session="handleDeleteSession"
          @open-knowledge-base-drawer="chatUiStore.openKnowledgeBaseDrawer"
          @logout="handleLogout"
        />
      </div>

      <main class="relative flex min-h-[calc(100vh-2rem)] flex-1 flex-col overflow-hidden rounded-[32px] border border-white/55 bg-white/55 shadow-frost backdrop-blur-2xl lg:min-h-0">
        <div class="flex items-center justify-between border-b border-slate-200/60 px-5 py-4 lg:hidden">
          <button
            type="button"
            class="flex h-11 w-11 items-center justify-center rounded-full border border-slate-200/70 bg-white/90 text-slate-700"
            @click="userAppStore.setMobileSidebarOpen(true)"
          >
            <el-icon :size="18">
              <Menu />
            </el-icon>
          </button>
          <div class="text-right">
            <p class="text-xs uppercase tracking-[0.3em] text-slate-400">Workspace</p>
            <p class="font-medium text-slate-900">{{ authStore.displayName }}</p>
          </div>
        </div>

        <router-view />
      </main>
    </div>

    <el-drawer
      v-model="userAppStore.mobileSidebarOpen"
      size="88%"
      direction="ltr"
      :with-header="false"
      class="!bg-transparent"
    >
      <SessionSidebar
        :sessions="sessionsQuery.data.value?.items ?? []"
        :active-session-id="activeSessionId"
        :loading="sessionsQuery.isLoading.value"
        :display-name="authStore.displayName"
        :avatar-url="authStore.user?.avatar_url"
        :role-label="roleLabel"
        :preferred-knowledge-base-name="chatUiStore.preferredKnowledgeBaseName"
        :is-creating="createSessionMutation.isPending.value"
        @create-session="handleCreateSession"
        @delete-session="handleDeleteSession"
        @logout="handleLogout"
        @open-knowledge-base-drawer="
          () => {
            userAppStore.setMobileSidebarOpen(false);
            chatUiStore.openKnowledgeBaseDrawer();
          }
        "
      />
    </el-drawer>

    <KnowledgeBaseDrawer
      v-model="chatUiStore.knowledgeBaseDrawerOpen"
      :knowledge-bases="knowledgeBasesQuery.data.value?.items ?? []"
      :loading="knowledgeBasesQuery.isLoading.value"
      :selected-id="chatUiStore.preferredKnowledgeBaseId"
      @select="
        ({ id, name }) => {
          chatUiStore.selectKnowledgeBase(id, name);
        }
      "
      @clear="chatUiStore.clearKnowledgeBase"
      @create-session="handleCreateSession"
    />
  </div>
</template>
