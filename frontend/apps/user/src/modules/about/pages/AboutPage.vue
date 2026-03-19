<script setup lang="ts">
import {
  ASSISTANT_NAME,
  BUILD_VERSION,
  PRODUCT_NAME,
  PRODUCT_TAGLINE,
  SectionHeading,
  SurfaceCard
} from '@private-kb/shared';

const quickSteps = [
  '登录并进入工作区。',
  '创建知识库并整理资料。',
  '在会话中继续提问与追问。'
];

const creators = '解子敬，贺博，赵登智，曾思耀，官柱天';

const faqs = [
  {
    question: '如何开始使用？',
    answer: '根据资料主题创建知识库，随后从会话页面开始提问。'
  },
  {
    question: '什么时候适合使用知识库？',
    answer: '当你需要围绕固定资料持续提问、整理答案或追踪文档时，知识库会更合适。'
  },
  {
    question: '在哪里管理账户信息？',
    answer: '你可以在侧栏进入个人资料与安全设置页面，维护昵称、头像和密码。'
  }
];
</script>

<template>
  <div class="soft-scrollbar flex h-full flex-1 flex-col overflow-y-auto px-5 py-6 lg:px-8">
    <div class="space-y-6">
      <SectionHeading
        eyebrow="About"
        :title="`关于 ${PRODUCT_NAME}`"
        :description="PRODUCT_TAGLINE"
      />

      <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
        <SurfaceCard>
          <p class="text-xs uppercase tracking-[0.3em] text-slate-400">Product intro</p>
          <h2 class="mt-4 font-serif text-3xl text-slate-900">{{ ASSISTANT_NAME }}</h2>
          <p class="mt-4 text-sm leading-7 text-slate-500">
            我们将会话、知识库与个人工作区融为一体，帮助你高效整理资料和开展研究。
          </p>
        </SurfaceCard>

        <SurfaceCard>
          <p class="text-xs uppercase tracking-[0.3em] text-slate-400">Quick start</p>
          <ol class="mt-4 space-y-4 text-sm leading-7 text-slate-600">
            <li v-for="(step, index) in quickSteps" :key="step" class="flex gap-4">
              <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-slate-900 text-xs font-semibold text-white">
                {{ index + 1 }}
              </span>
              <span>{{ step }}</span>
            </li>
          </ol>
        </SurfaceCard>
      </div>

      <SurfaceCard>
        <p class="text-xs uppercase tracking-[0.3em] text-slate-400">FAQ</p>
        <div class="mt-5 grid gap-4 lg:grid-cols-3">
          <div
            v-for="item in faqs"
            :key="item.question"
            class="rounded-[24px] border border-slate-100 bg-white px-5 py-5"
          >
            <p class="text-lg font-semibold text-slate-900">{{ item.question }}</p>
            <p class="mt-3 text-sm leading-6 text-slate-500">{{ item.answer }}</p>
          </div>
        </div>
      </SurfaceCard>

      <SurfaceCard tone="night">
        <div class="grid gap-4 md:grid-cols-3">
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-white/45">Build</p>
            <p class="mt-3 text-xl font-semibold text-white">{{ BUILD_VERSION }}</p>
          </div>
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-white/45">Made by</p>
            <p class="mt-3 text-sm leading-6 text-white/76">
              由HNUST 23计科四班课程设计团队精心打造
            </p>
          </div>
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-white/45">CREATORS</p>
            <p class="mt-3 text-sm leading-6 text-white/76">
              {{ creators }}
            </p>
          </div>
        </div>
      </SurfaceCard>
    </div>
  </div>
</template>
