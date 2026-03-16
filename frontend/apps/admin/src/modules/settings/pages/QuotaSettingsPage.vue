<script setup lang="ts">
import { useQuery } from '@tanstack/vue-query';

import {
  EmptyStatePanel,
  SectionHeading,
  SurfaceCard,
  adminApi,
  formatBytes,
  formatDateTime,
  queryKeys
} from '@private-kb/shared';
import { getErrorMessage } from '@private-kb/shared/utils/errors';

const quotaQuery = useQuery({
  queryKey: queryKeys.adminQuotaPolicies,
  queryFn: () => adminApi.listQuotaPolicies()
});
</script>

<template>
  <div class="space-y-6">
    <SectionHeading eyebrow="Quota" title="配额策略" />

    <SurfaceCard>
      <div class="grid gap-4 md:grid-cols-3">
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">策略总数</p>
          <p class="mt-4 font-serif text-3xl text-slate-900">
            {{ quotaQuery.data.value?.length ?? 0 }}
          </p>
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">系统默认</p>
          <p class="mt-4 font-serif text-3xl text-slate-900">
            {{
              quotaQuery.data.value?.filter((item) => item.scope_type === 'system_default')
                .length ?? 0
            }}
          </p>
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">用户定制</p>
          <p class="mt-4 font-serif text-3xl text-slate-900">
            {{
              quotaQuery.data.value?.filter((item) => item.scope_type === 'user').length ?? 0
            }}
          </p>
        </div>
      </div>
    </SurfaceCard>

    <SurfaceCard v-if="quotaQuery.data.value?.length" :padded="false" class="overflow-hidden">
      <el-table :data="quotaQuery.data.value" stripe>
        <el-table-column label="作用域" width="140">
          <template #default="{ row }">
            {{ row.scope_type === 'system_default' ? '系统默认' : '用户专属' }}
          </template>
        </el-table-column>
        <el-table-column label="目标 ID" min-width="220">
          <template #default="{ row }">
            {{ row.scope_id || '全局默认' }}
          </template>
        </el-table-column>
        <el-table-column label="日总 Token" width="140" prop="daily_total_tokens_limit" />
        <el-table-column label="存储上限" width="140">
          <template #default="{ row }">
            {{ formatBytes(row.storage_bytes_limit) }}
          </template>
        </el-table-column>
        <el-table-column label="文档上限" width="120" prop="document_count_limit" />
        <el-table-column label="告警阈值" width="120">
          <template #default="{ row }">
            {{ `${Math.round(Number(row.warn_ratio) * 100)}%` }}
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
      v-else-if="quotaQuery.isError.value"
      title="暂时无法读取配额策略"
      :description="getErrorMessage(quotaQuery.error.value)"
      action-label="重新加载"
      @action="quotaQuery.refetch()"
    />

    <EmptyStatePanel
      v-else
      title="当前没有配额策略"
      description="数据库中还没有可展示的配额配置。"
    />
  </div>
</template>
