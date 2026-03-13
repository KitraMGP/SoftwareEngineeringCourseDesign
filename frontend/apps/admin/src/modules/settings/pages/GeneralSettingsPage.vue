<script setup lang="ts">
import { SectionHeading, SurfaceCard } from '@private-kb/shared';
</script>

<template>
  <div class="space-y-6">
    <SectionHeading
      eyebrow="Settings"
      title="系统设置"
      description="系统参数管理、默认配置和基础开关会集中到这里。当前保留表单容器与版式。"
    />

    <SurfaceCard>
      <div class="grid gap-4 md:grid-cols-2">
        <div class="rounded-[22px] bg-slate-50 px-5 py-5 text-sm leading-6 text-slate-500">
          系统参数列表待接入 `/admin/settings`
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5 text-sm leading-6 text-slate-500">
          当前页面已经为后续表单、保存按钮和变更提示预留结构。
        </div>
      </div>
    </SurfaceCard>
  </div>
</template>
