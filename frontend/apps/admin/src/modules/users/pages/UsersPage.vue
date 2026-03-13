<script setup lang="ts">
import { SectionHeading, SurfaceCard } from '@private-kb/shared';

const rows = [
  ['用户列表', '待接入 `/admin/users`'],
  ['冻结/解冻', '待接入 `/admin/users/{userId}/freeze` 与 `/unfreeze`'],
  ['密码重置', '待接入 `/admin/users/{userId}/password`']
];
</script>

<template>
  <div class="space-y-6">
    <SectionHeading
      eyebrow="Users"
      title="用户管理"
      description="当前页面先固定操作层级和表格结构，后续接口可直接替换占位数据源。"
    />

    <SurfaceCard>
      <table class="min-w-full text-left">
        <thead class="text-xs uppercase tracking-[0.28em] text-slate-400">
          <tr>
            <th class="px-2 py-3">能力</th>
            <th class="px-2 py-3">当前状态</th>
          </tr>
        </thead>
        <tbody class="text-sm text-slate-600">
          <tr v-for="row in rows" :key="row[0]" class="border-t border-slate-100">
            <td class="px-2 py-4 font-medium text-slate-900">{{ row[0] }}</td>
            <td class="px-2 py-4">{{ row[1] }}</td>
          </tr>
        </tbody>
      </table>
    </SurfaceCard>
  </div>
</template>
