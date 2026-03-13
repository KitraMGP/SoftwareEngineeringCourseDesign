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
  '先登录并创建知识库，确认资料归属范围。',
  '上传 txt、markdown、docx 或 pdf 文档，等待状态轮询完成。',
  '从侧栏进入会话，新建与知识库绑定的新上下文。'
];

const faqs = [
  {
    question: '为什么切换知识库会从新会话开始？',
    answer: '当前产品将会话上下文和知识库上下文一起管理，从新会话开始可以避免历史消息混入新的资料边界。'
  },
  {
    question: '现在能直接发送问题吗？',
    answer: '可以，但当前只开放未绑定知识库的空会话实时问答；知识库问答、停止生成和重生成还在后续联调范围内。'
  },
  {
    question: 'PDF 为什么可能上传后失败？',
    answer: '当前后端会先接受 PDF 文件，但解析器尚未接入，因此失败信息会在文档状态里返回。'
  }
];
</script>

<template>
  <div class="soft-scrollbar flex h-full flex-1 flex-col overflow-y-auto px-5 py-6 lg:px-8">
    <div class="space-y-6">
      <SectionHeading
        eyebrow="About"
        :title="`关于 ${PRODUCT_NAME}`"
        :description="`${PRODUCT_TAGLINE}。这里汇总当前版本的定位、使用方式和项目构建信息。`"
      />

      <div class="grid gap-5 xl:grid-cols-[1.05fr_0.95fr]">
        <SurfaceCard>
          <p class="text-xs uppercase tracking-[0.3em] text-slate-400">Product intro</p>
          <h2 class="mt-4 font-serif text-3xl text-slate-900">{{ ASSISTANT_NAME }}</h2>
          <p class="mt-4 text-sm leading-7 text-slate-500">
            当前版本已经具备“私有知识库问答系统”的用户端主路径：认证、空会话实时问答、知识库与文档管理、个人中心，以及管理端的独立应用骨架。
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
            <p class="text-xs uppercase tracking-[0.3em] text-white/45">User app</p>
            <p class="mt-3 text-sm leading-6 text-white/76">
              登录、空会话实时问答、知识库管理、关于页和个人中心已落入 V1。
            </p>
          </div>
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-white/45">Admin app</p>
            <p class="mt-3 text-sm leading-6 text-white/76">
              已建立独立后台骨架，等待管理端接口逐步替换占位数据。
            </p>
          </div>
        </div>
      </SurfaceCard>
    </div>
  </div>
</template>
