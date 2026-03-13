<script setup lang="ts">
import { PRODUCT_NAME } from '@private-kb/shared/constants/app';
</script>

<template>
  <div class="flex min-h-screen items-center justify-center px-5 py-10">
    <div class="grid w-full max-w-6xl gap-8 lg:grid-cols-[1fr_460px] lg:items-center">
      <div class="hidden lg:block">
        <p class="text-xs uppercase tracking-[0.34em] text-slate-400">Admin console</p>
        <h1 class="mt-6 font-serif text-5xl text-slate-900">{{ PRODUCT_NAME }} 后台</h1>
        <p class="mt-5 max-w-2xl text-base leading-8 text-slate-500">
          管理端当前以真实鉴权与后台路由骨架为主，具体的用户、任务、审计和系统设置数据会随着接口逐步替换占位结构。
        </p>
      </div>

      <div class="rounded-[32px] border border-white/70 bg-white/82 p-8 shadow-frost backdrop-blur">
        <router-view />
      </div>
    </div>
  </div>
</template>
