<script setup lang="ts">
import { Monitor, Operation, Setting, User } from '@element-plus/icons-vue';
import { useRouter } from 'vue-router';

import { PRODUCT_NAME, UserAvatar } from '@private-kb/shared';
import { useAuthStore } from '@private-kb/shared/auth/useAuthStore';

const router = useRouter();
const authStore = useAuthStore();

const navigation = [
  { name: 'dashboard', label: 'Dashboard', icon: Monitor },
  { name: 'users', label: '用户', icon: User },
  { name: 'providers', label: '模型配置', icon: Operation },
  { name: 'tasks', label: '任务', icon: Operation },
  { name: 'audit-logs', label: '审计', icon: Operation },
  { name: 'settings-general', label: '系统设置', icon: Setting },
  { name: 'settings-quota', label: '配额策略', icon: Setting }
];

async function handleLogout() {
  try {
    await authStore.logout();
  } catch {
    // Auth state is cleared in the store finally block.
  } finally {
    router.push({ name: 'admin-login' });
  }
}
</script>

<template>
  <div class="flex min-h-screen flex-col lg:flex-row">
    <aside class="border-b border-slate-200 bg-slate-950 px-5 py-6 text-white lg:min-h-screen lg:w-[300px] lg:border-b-0 lg:border-r lg:px-6">
      <div class="space-y-2">
        <p class="text-xs uppercase tracking-[0.34em] text-white/45">Admin workspace</p>
        <h1 class="font-serif text-3xl">{{ PRODUCT_NAME }}</h1>
        <p class="text-sm leading-6 text-white/62">后台菜单已经拆分完成，可直接承接后续真实接口。</p>
      </div>

      <nav class="mt-8 grid gap-2">
        <router-link
          v-for="item in navigation"
          :key="item.name"
          :to="{ name: item.name }"
          class="flex items-center gap-3 rounded-[18px] px-4 py-3 text-sm font-medium transition"
          active-class="bg-white text-slate-900"
          exact-active-class="bg-white text-slate-900"
        >
          <el-icon>
            <component :is="item.icon" />
          </el-icon>
          <span>{{ item.label }}</span>
        </router-link>
      </nav>

      <div class="mt-8 rounded-[22px] border border-white/10 bg-white/8 px-4 py-4">
        <div class="flex items-center gap-3">
          <UserAvatar
            :label="authStore.displayName"
            :avatar-url="authStore.user?.avatar_url"
            :size="48"
            tone="light"
          />
          <div class="min-w-0">
            <p class="truncate text-sm font-semibold">{{ authStore.displayName }}</p>
            <p class="mt-1 truncate text-xs text-white/45">{{ authStore.user?.email }}</p>
          </div>
        </div>
        <button
          type="button"
          class="mt-4 rounded-full bg-white px-4 py-2.5 text-sm font-semibold text-slate-900 transition hover:bg-slate-100"
          @click="handleLogout"
        >
          退出登录
        </button>
      </div>
    </aside>

    <main class="flex-1 px-5 py-6 lg:px-8">
      <router-view />
    </main>
  </div>
</template>
