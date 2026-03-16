<script setup lang="ts">
import { PRODUCT_NAME } from '@private-kb/shared/constants/app';
</script>

<template>
  <div class="flex min-h-screen items-center justify-center px-5 py-10">
    <div class="grid w-full max-w-6xl gap-8 lg:grid-cols-[1fr_460px] lg:items-center">
      <div class="hidden lg:block">
        <p class="text-xs uppercase tracking-[0.34em] text-slate-400">Admin console</p>
        <h1 class="mt-6 font-serif text-5xl text-slate-900">{{ PRODUCT_NAME }} 后台</h1>
      </div>

      <div class="rounded-[32px] border border-white/70 bg-white/82 p-8 shadow-frost backdrop-blur">
        <router-view />
      </div>
    </div>
  </div>
</template>
