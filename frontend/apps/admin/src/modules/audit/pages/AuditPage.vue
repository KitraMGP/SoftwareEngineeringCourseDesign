<script setup lang="ts">
import { computed, ref } from 'vue';
import { useQuery } from '@tanstack/vue-query';

import {
  AppStatusBadge,
  EmptyStatePanel,
  SectionHeading,
  SurfaceCard,
  adminApi,
  formatDateTime,
  queryKeys,
  shortenText
} from '@private-kb/shared';
import { getErrorMessage } from '@private-kb/shared/utils/errors';

const pageSize = 20;
const page = ref(1);

const auditQuery = useQuery({
  queryKey: computed(() => queryKeys.adminAuditLogs({ page: page.value, size: pageSize })),
  queryFn: () => adminApi.listAuditLogs({ page: page.value, size: pageSize })
});

const total = computed(() => auditQuery.data.value?.total ?? 0);

function formatMetadata(value: unknown) {
  if (value === null || value === undefined) {
    return '{}';
  }

  try {
    return JSON.stringify(value);
  } catch {
    return '{}';
  }
}
</script>

<template>
  <div class="space-y-6">
    <SectionHeading eyebrow="Audit" title="审计日志" />

    <SurfaceCard>
      <div class="grid gap-4 md:grid-cols-3">
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">日志总数</p>
          <p class="mt-4 font-serif text-3xl text-slate-900">{{ total }}</p>
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">当前页容量</p>
          <p class="mt-4 font-serif text-3xl text-slate-900">{{ pageSize }}</p>
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">最新刷新</p>
          <p class="mt-4 text-sm text-slate-500">
            {{ auditQuery.isFetching.value ? '同步中...' : '已完成' }}
          </p>
        </div>
      </div>
    </SurfaceCard>

    <SurfaceCard
      v-if="auditQuery.data.value?.items.length"
      :padded="false"
      class="overflow-hidden"
    >
      <el-table :data="auditQuery.data.value.items" stripe>
        <el-table-column label="时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="操作者" min-width="180">
          <template #default="{ row }">
            {{ row.actor_label || row.actor_user_id || row.actor_role }}
          </template>
        </el-table-column>

        <el-table-column label="动作" min-width="180" prop="action" />

        <el-table-column label="资源" min-width="180">
          <template #default="{ row }">
            <span v-if="row.resource_type">
              {{ row.resource_type }}
              <span v-if="row.resource_id"> / {{ row.resource_id }}</span>
            </span>
            <span v-else>无</span>
          </template>
        </el-table-column>

        <el-table-column label="结果" width="120">
          <template #default="{ row }">
            <AppStatusBadge :tone="row.result === 'success' ? 'success' : 'danger'">
              {{ row.result === 'success' ? '成功' : '失败' }}
            </AppStatusBadge>
          </template>
        </el-table-column>

        <el-table-column label="元数据" min-width="260">
          <template #default="{ row }">
            <span :title="formatMetadata(row.metadata)">
              {{ shortenText(formatMetadata(row.metadata), 120) }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </SurfaceCard>

    <EmptyStatePanel
      v-else-if="auditQuery.isError.value"
      title="暂时无法读取审计日志"
      :description="getErrorMessage(auditQuery.error.value)"
      action-label="重新加载"
      @action="auditQuery.refetch()"
    />

    <EmptyStatePanel
      v-else
      title="当前没有审计日志"
      description="数据库中还没有可展示的审计记录。"
    />

    <div v-if="total > pageSize" class="flex justify-end">
      <el-pagination
        v-model:current-page="page"
        :page-size="pageSize"
        layout="prev, pager, next"
        :total="total"
      />
    </div>
  </div>
</template>
