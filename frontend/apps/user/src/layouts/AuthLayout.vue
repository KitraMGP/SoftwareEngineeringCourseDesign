<script setup lang="ts">
import { computed } from 'vue';

import { ASSISTANT_NAME, PRODUCT_NAME, PRODUCT_TAGLINE } from '@private-kb/shared/constants/app';

const backgroundImage = computed(
  () => `url(${new URL('../../../../UI原型/登录界面背景图.jpg', import.meta.url).href})`
);
</script>

<template>
  <div class="relative min-h-screen overflow-hidden">
    <div class="absolute inset-0 bg-cover bg-center" :style="{ backgroundImage }" />
    <div class="absolute inset-0 bg-[rgba(6,12,26,0.46)] backdrop-blur-[1.5px]" />
    <div
      class="absolute inset-0 bg-[radial-gradient(circle_at_top,rgba(145,186,255,0.28),transparent_35%),linear-gradient(180deg,rgba(6,10,20,0.16),rgba(6,10,20,0.46))]"
    />

    <div class="relative z-10 mx-auto flex min-h-screen max-w-7xl items-center px-6 py-12">
      <div class="grid w-full gap-10 lg:grid-cols-[1.1fr_520px] lg:items-center">
        <div class="hidden text-white lg:block">
          <p class="text-sm uppercase tracking-[0.4em] text-white/70">
            Private Knowledge Workspace
          </p>
          <h1 class="mt-6 max-w-2xl font-serif text-5xl leading-tight">
            {{ PRODUCT_NAME }}
          </h1>
          <p class="mt-5 max-w-xl text-lg leading-8 text-white/78">
            {{ PRODUCT_TAGLINE }}。{{ ASSISTANT_NAME }} 会在统一上下文中整理你的会话、知识库与文档状态，
            让团队问答回到可追踪、可运营的工作流里。
          </p>
          <div class="mt-12 grid max-w-2xl grid-cols-3 gap-4">
            <div class="rounded-[24px] border border-white/20 bg-white/10 p-5 backdrop-blur">
              <p class="text-xs uppercase tracking-[0.28em] text-white/60">会话管理</p>
              <p class="mt-3 text-sm leading-6 text-white/84">
                登录后直接进入最近一次工作区，延续上下文。
              </p>
            </div>
            <div class="rounded-[24px] border border-white/20 bg-white/10 p-5 backdrop-blur">
              <p class="text-xs uppercase tracking-[0.28em] text-white/60">知识库联调</p>
              <p class="mt-3 text-sm leading-6 text-white/84">
                已接入真实知识库与文档接口，便于先把资产整理起来。
              </p>
            </div>
            <div class="rounded-[24px] border border-white/20 bg-white/10 p-5 backdrop-blur">
              <p class="text-xs uppercase tracking-[0.28em] text-white/60">后续扩展</p>
              <p class="mt-3 text-sm leading-6 text-white/84">
                空会话流式问答已接通，知识库问答、重生成和管理端将沿当前骨架继续接入。
              </p>
            </div>
          </div>
        </div>

        <div
          class="mx-auto w-full max-w-[520px] rounded-[36px] border border-white/22 bg-[rgba(255,255,255,0.12)] p-3 shadow-[0_40px_120px_rgba(4,10,22,0.34)] backdrop-blur-xl"
        >
          <div
            class="rounded-[30px] border border-white/45 bg-[rgba(245,248,253,0.92)] p-6 shadow-[0_24px_80px_rgba(14,28,58,0.18)] backdrop-blur-2xl md:p-10"
          >
            <router-view />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
