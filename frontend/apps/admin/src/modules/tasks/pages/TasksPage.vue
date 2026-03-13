<script setup lang="ts">
import { SectionHeading, SurfaceCard } from '@private-kb/shared';
</script>

<template>
  <div class="space-y-6">
    <SectionHeading
      eyebrow="Tasks"
      title="任务管理"
      description="任务管理页会对齐后端任务系统，支持状态筛选和失败重试。当前接口仍返回 501。"
    />

    <SurfaceCard>
      <div class="rounded-[22px] border border-dashed border-slate-200 bg-slate-50 px-5 py-8 text-sm leading-6 text-slate-500">
        等待 `/admin/tasks` 与 `/admin/tasks/{taskId}/retry` 接入后，这里会替换为真实表格、筛选器和重试操作。
      </div>
    </SurfaceCard>
  </div>
</template>
