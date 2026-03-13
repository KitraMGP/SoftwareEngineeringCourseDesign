<script setup lang="ts">
import { SectionHeading, StatPanel, SurfaceCard } from '@private-kb/shared';

const cards = [
  { label: '用户数', value: '--', helper: '等待 `/admin/users` 接入真实列表统计' },
  { label: '知识库数', value: '--', helper: '当前接口未开放后台汇总视图' },
  { label: '文档数', value: '--', helper: '未来可汇总全部知识库文档体量' },
  { label: '任务状态', value: '--', helper: '任务管理接口准备替换占位卡片' }
];
</script>

<template>
  <div class="space-y-6">
    <SectionHeading
      eyebrow="Dashboard"
      title="后台总览"
      description="管理端接口当前仍以 501 占位为主，因此这里先保留信息架构、摘要卡片和操作入口。"
    />

    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <StatPanel
        v-for="item in cards"
        :key="item.label"
        :label="item.label"
        :value="item.value"
        :helper="item.helper"
      />
    </div>

    <SurfaceCard>
      <p class="text-xs uppercase tracking-[0.3em] text-slate-400">Ready for integration</p>
      <div class="mt-5 grid gap-4 lg:grid-cols-3">
        <div class="rounded-[22px] bg-slate-50 px-5 py-5 text-sm leading-6 text-slate-500">
          用户、任务、审计和系统设置页面已经拆到独立路由。
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5 text-sm leading-6 text-slate-500">
          当前共享同一套鉴权刷新逻辑，管理员角色守卫已生效。
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5 text-sm leading-6 text-slate-500">
          OpenAPI 已补齐 501 占位接口，后续替换页面数据层会更直接。
        </div>
      </div>
    </SurfaceCard>
  </div>
</template>
