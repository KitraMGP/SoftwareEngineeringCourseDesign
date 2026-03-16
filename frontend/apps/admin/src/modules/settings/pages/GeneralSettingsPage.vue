<script setup lang="ts">
import { computed } from 'vue';
import { useQuery } from '@tanstack/vue-query';

import {
  EmptyStatePanel,
  SectionHeading,
  SurfaceCard,
  adminApi,
  formatDateTime,
  queryKeys
} from '@private-kb/shared';
import { getErrorMessage } from '@private-kb/shared/utils/errors';

const settingsQuery = useQuery({
  queryKey: queryKeys.adminSystemSettings,
  queryFn: () => adminApi.listSystemSettings()
});

const categoryCount = computed(
  () => new Set((settingsQuery.data.value ?? []).map((item) => item.category)).size
);

function formatSettingValue(value: unknown): string {
  if (value === null || value === undefined) {
    return 'null';
  }
  if (typeof value === 'string') {
    return value;
  }
  return JSON.stringify(value, null, 2);
}
</script>

<template>
  <div class="space-y-6">
    <SectionHeading eyebrow="Settings" title="系统设置" />

    <SurfaceCard>
      <div class="grid gap-4 md:grid-cols-2">
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">设置项</p>
          <p class="mt-4 font-serif text-3xl text-slate-900">
            {{ settingsQuery.data.value?.length ?? 0 }}
          </p>
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">分类数</p>
          <p class="mt-4 font-serif text-3xl text-slate-900">{{ categoryCount }}</p>
        </div>
      </div>
    </SurfaceCard>

    <SurfaceCard
      v-if="settingsQuery.data.value?.length"
      :padded="false"
      class="overflow-hidden"
    >
      <el-table :data="settingsQuery.data.value" stripe>
        <el-table-column label="分类" width="160" prop="category" />
        <el-table-column label="键名" min-width="220" prop="key" />
        <el-table-column label="值" min-width="340">
          <template #default="{ row }">
            <pre class="whitespace-pre-wrap break-all py-2 font-mono text-xs text-slate-600">{{
              formatSettingValue(row.value)
            }}</pre>
          </template>
        </el-table-column>
        <el-table-column label="说明" min-width="220">
          <template #default="{ row }">
            {{ row.description || '无' }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
      </el-table>
    </SurfaceCard>

    <EmptyStatePanel
      v-else-if="settingsQuery.isError.value"
      title="暂时无法读取系统设置"
      :description="getErrorMessage(settingsQuery.error.value)"
      action-label="重新加载"
      @action="settingsQuery.refetch()"
    />

    <EmptyStatePanel
      v-else
      title="当前没有系统设置"
      description="数据库中还没有可展示的系统参数。"
    />
  </div>
</template>
