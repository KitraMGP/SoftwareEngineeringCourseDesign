<script setup lang="ts">
import { SectionHeading, SurfaceCard } from '@private-kb/shared';
</script>

<template>
  <div class="space-y-6">
    <SectionHeading
      eyebrow="Audit"
      title="审计日志"
      description="后续会承接时间筛选、结果筛选和动作检索。当前先保留查询布局和说明区。"
    />

    <SurfaceCard>
      <div class="grid gap-4 lg:grid-cols-3">
        <div class="rounded-[22px] bg-slate-50 px-5 py-5 text-sm leading-6 text-slate-500">
          按时间范围筛选
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5 text-sm leading-6 text-slate-500">
          按动作和结果筛选
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5 text-sm leading-6 text-slate-500">
          审计详情列表等待 `/admin/audit-logs`
        </div>
      </div>
    </SurfaceCard>
  </div>
</template>
