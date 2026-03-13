<script setup lang="ts">
import { SectionHeading, SurfaceCard } from '@private-kb/shared';
</script>

<template>
  <div class="space-y-6">
    <SectionHeading
      eyebrow="Quota"
      title="配额策略"
      description="配额管理会跟随管理员接口补齐，这里先保留策略列表与编辑工作区。"
    />

    <SurfaceCard>
      <div class="rounded-[22px] border border-dashed border-slate-200 bg-slate-50 px-5 py-8 text-sm leading-6 text-slate-500">
        `/admin/quota-policies` 已在 OpenAPI 中补齐占位结构，前端页面可直接等待真实响应接入。
      </div>
    </SurfaceCard>
  </div>
</template>
