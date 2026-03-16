<script setup lang="ts">
import { computed } from 'vue';
import { useQuery } from '@tanstack/vue-query';

import {
  EmptyStatePanel,
  SectionHeading,
  StatPanel,
  SurfaceCard,
  adminApi,
  queryKeys
} from '@private-kb/shared';
import { getErrorMessage } from '@private-kb/shared/utils/errors';

const overviewQuery = useQuery({
  queryKey: queryKeys.adminOverview,
  queryFn: () => adminApi.getOverview()
});

const cards = computed(() => {
  const overview = overviewQuery.data.value;
  if (!overview) {
    return [];
  }

  return [
    {
      label: '用户总数',
      value: `${overview.user_count}`,
      helper: `活跃 ${overview.active_user_count}`
    },
    {
      label: '知识库',
      value: `${overview.knowledge_base_count}`,
      helper: `文档 ${overview.document_count}`
    },
    {
      label: '待处理任务',
      value: `${overview.pending_task_count}`,
      helper: `失败 ${overview.failed_task_count}`
    },
    {
      label: '活跃占比',
      value:
        overview.user_count > 0
          ? `${Math.round((overview.active_user_count / overview.user_count) * 100)}%`
          : '0%',
      helper: '按当前未冻结账号计算'
    }
  ];
});
</script>

<template>
  <div class="space-y-6">
    <SectionHeading eyebrow="Dashboard" title="后台总览" />

    <div v-if="overviewQuery.isLoading.value" class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <div
        v-for="item in 4"
        :key="item"
        class="h-[136px] animate-pulse rounded-[24px] bg-white/70"
      />
    </div>

    <template v-else-if="overviewQuery.data.value">
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatPanel
          v-for="item in cards"
          :key="item.label"
          :label="item.label"
          :value="item.value"
          :helper="item.helper"
        />
      </div>

      <div class="grid gap-5 xl:grid-cols-[1.4fr_1fr]">
        <SurfaceCard>
          <p class="text-xs uppercase tracking-[0.3em] text-slate-400">Workspace</p>
          <div class="mt-5 grid gap-4 lg:grid-cols-3">
            <div class="rounded-[22px] bg-slate-50 px-5 py-5">
              <p class="text-xs uppercase tracking-[0.28em] text-slate-400">用户状态</p>
              <p class="mt-4 font-serif text-3xl text-slate-900">
                {{ overviewQuery.data.value.active_user_count }}
              </p>
              <p class="mt-3 text-sm leading-6 text-slate-500">
                当前仍可登录和使用系统的账号数量。
              </p>
            </div>

            <div class="rounded-[22px] bg-slate-50 px-5 py-5">
              <p class="text-xs uppercase tracking-[0.28em] text-slate-400">知识沉淀</p>
              <p class="mt-4 font-serif text-3xl text-slate-900">
                {{ overviewQuery.data.value.document_count }}
              </p>
              <p class="mt-3 text-sm leading-6 text-slate-500">
                已接入文档总量，对应
                {{ overviewQuery.data.value.knowledge_base_count }} 个知识库。
              </p>
            </div>

            <div class="rounded-[22px] bg-slate-50 px-5 py-5">
              <p class="text-xs uppercase tracking-[0.28em] text-slate-400">任务积压</p>
              <p class="mt-4 font-serif text-3xl text-slate-900">
                {{
                  overviewQuery.data.value.pending_task_count +
                  overviewQuery.data.value.failed_task_count
                }}
              </p>
              <p class="mt-3 text-sm leading-6 text-slate-500">
                待处理与失败任务之和，可直接转到任务页继续处理。
              </p>
            </div>
          </div>
        </SurfaceCard>

        <SurfaceCard tone="soft">
          <p class="text-xs uppercase tracking-[0.3em] text-slate-400">Focus</p>
          <div class="mt-5 space-y-4 text-sm leading-6 text-slate-600">
            <div class="rounded-[22px] border border-slate-200/70 bg-white/80 px-5 py-4">
              当前失败任务
              <span class="font-semibold text-slate-900">
                {{ overviewQuery.data.value.failed_task_count }}
              </span>
              条。
            </div>
            <div class="rounded-[22px] border border-slate-200/70 bg-white/80 px-5 py-4">
              活跃用户占比约
              <span class="font-semibold text-slate-900">
                {{
                  overviewQuery.data.value.user_count > 0
                    ? Math.round(
                        (overviewQuery.data.value.active_user_count /
                          overviewQuery.data.value.user_count) *
                          100
                      )
                    : 0
                }}%
              </span>
              。
            </div>
            <div class="rounded-[22px] border border-slate-200/70 bg-white/80 px-5 py-4">
              若失败任务持续增长，优先检查 worker 与文档处理链路。
            </div>
          </div>
        </SurfaceCard>
      </div>
    </template>

    <EmptyStatePanel
      v-else
      title="暂时无法读取后台总览"
      :description="getErrorMessage(overviewQuery.error.value)"
    />
  </div>
</template>
