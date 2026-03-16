<script setup lang="ts">
import { useRouter } from 'vue-router';

import { useAuthStore } from '@private-kb/shared/auth/useAuthStore';

const router = useRouter();
const authStore = useAuthStore();

async function handleLogout() {
  try {
    await authStore.logout();
  } finally {
    await router.push({ name: 'admin-login' });
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center px-5 py-10">
    <div
      class="w-full max-w-3xl rounded-[32px] border border-white/70 bg-white/82 px-8 py-10 text-center shadow-frost backdrop-blur"
    >
      <p class="text-xs uppercase tracking-[0.34em] text-slate-400">Admin access</p>
      <h1 class="mt-5 font-serif text-3xl text-slate-900">当前账号没有后台权限</h1>
      <p class="mx-auto mt-4 max-w-xl text-sm leading-6 text-slate-500">
        你的账号不具备管理员角色。你可以直接切换到管理员账号，或先退出当前账号后重新登录。
      </p>

      <div class="mt-8 flex flex-col justify-center gap-3 sm:flex-row">
        <button
          type="button"
          class="rounded-full bg-slate-950 px-5 py-3 text-sm font-semibold text-white transition hover:bg-slate-900"
          @click="router.push({ name: 'admin-login' })"
        >
          切换账号
        </button>
        <button
          type="button"
          class="rounded-full border border-slate-200 px-5 py-3 text-sm font-semibold text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
          @click="handleLogout"
        >
          退出当前账号
        </button>
      </div>
    </div>
  </div>
</template>
