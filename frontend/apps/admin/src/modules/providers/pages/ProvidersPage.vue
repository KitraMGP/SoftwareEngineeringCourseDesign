<script setup lang="ts">
import { computed } from 'vue';
import { useQuery } from '@tanstack/vue-query';

import {
  AppStatusBadge,
  EmptyStatePanel,
  SectionHeading,
  SurfaceCard,
  adminApi,
  formatDateTime,
  queryKeys
} from '@private-kb/shared';
import { getErrorMessage } from '@private-kb/shared/utils/errors';

const providersQuery = useQuery({
  queryKey: queryKeys.adminProviderConfigs,
  queryFn: () => adminApi.listProviderConfigs()
});

const providerCount = computed(() => providersQuery.data.value?.length ?? 0);
const enabledCount = computed(
  () => providersQuery.data.value?.filter((item) => item.is_enabled).length ?? 0
);
const configuredCount = computed(
  () => providersQuery.data.value?.filter((item) => item.has_api_key).length ?? 0
);
</script>

<template>
  <div class="space-y-6">
    <SectionHeading eyebrow="Providers" title="模型配置" />

    <SurfaceCard>
      <div class="grid gap-4 md:grid-cols-3">
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">Provider 总数</p>
          <p class="mt-4 font-serif text-3xl text-slate-900">{{ providerCount }}</p>
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">已启用</p>
          <p class="mt-4 font-serif text-3xl text-slate-900">{{ enabledCount }}</p>
        </div>
        <div class="rounded-[22px] bg-slate-50 px-5 py-5">
          <p class="text-xs uppercase tracking-[0.28em] text-slate-400">已配置密钥</p>
          <p class="mt-4 font-serif text-3xl text-slate-900">{{ configuredCount }}</p>
        </div>
      </div>
    </SurfaceCard>

    <SurfaceCard
      v-if="providersQuery.data.value?.length"
      :padded="false"
      class="overflow-hidden"
    >
      <el-table :data="providersQuery.data.value" stripe>
        <el-table-column label="Provider" min-width="160" prop="provider" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <AppStatusBadge :tone="row.is_enabled ? 'success' : 'warning'">
              {{ row.is_enabled ? '已启用' : '已停用' }}
            </AppStatusBadge>
          </template>
        </el-table-column>
        <el-table-column label="API Key" width="120">
          <template #default="{ row }">
            <AppStatusBadge :tone="row.has_api_key ? 'success' : 'info'">
              {{ row.has_api_key ? '已配置' : '未配置' }}
            </AppStatusBadge>
          </template>
        </el-table-column>
        <el-table-column label="Base URL" min-width="220" prop="base_url" />
        <el-table-column label="聊天模型" min-width="180" prop="default_chat_model" />
        <el-table-column
          label="Embedding 模型"
          min-width="180"
          prop="default_embedding_model"
        />
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>
      </el-table>
    </SurfaceCard>

    <EmptyStatePanel
      v-else-if="providersQuery.isError.value"
      title="暂时无法读取模型配置"
      :description="getErrorMessage(providersQuery.error.value)"
      action-label="重新加载"
      @action="providersQuery.refetch()"
    />

    <EmptyStatePanel
      v-else
      title="当前没有 Provider 配置"
      description="数据库中还没有可展示的模型服务配置。"
    />
  </div>
</template>
