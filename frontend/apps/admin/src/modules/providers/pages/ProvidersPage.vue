<script setup lang="ts">
import { SectionHeading, SurfaceCard } from '@private-kb/shared';
</script>

<template>
  <div class="space-y-6">
    <SectionHeading
      eyebrow="Providers"
      title="模型配置"
      description="这里会承接 provider 配置、默认模型和启停状态。OpenAPI 目前已补上 `/admin/provider-configs` 占位接口。"
    />

    <SurfaceCard>
      <div class="grid gap-4 md:grid-cols-2">
        <div class="rounded-[22px] bg-slate-50 px-5 py-5 text-sm leading-6 text-slate-500">
          Provider 列表与凭证状态尚未接通。
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5 text-sm leading-6 text-slate-500">
          当前页面先保留配置面板位置，等待真实接口覆盖。
        </div>
      </div>
    </SurfaceCard>
  </div>
</template>
