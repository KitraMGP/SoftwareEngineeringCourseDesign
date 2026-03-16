<script setup lang="ts">
import type { AdminTask } from '@private-kb/shared';
import { computed, ref } from 'vue';
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { ElMessage, ElMessageBox } from 'element-plus';

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

const queryClient = useQueryClient();

const pageSize = 20;
const page = ref(1);
const status = ref('');
const taskType = ref('');

const tasksQuery = useQuery({
  queryKey: computed(() =>
    queryKeys.adminTasks({
      page: page.value,
      size: pageSize,
      status: status.value,
      task_type: taskType.value
    })
  ),
  queryFn: () =>
    adminApi.listTasks({
      page: page.value,
      size: pageSize,
      status: status.value || undefined,
      task_type: taskType.value || undefined
    })
});

const retryMutation = useMutation({
  mutationFn: (taskId: string) => adminApi.retryTask(taskId),
  onSuccess: async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.adminTasksRoot }),
      queryClient.invalidateQueries({ queryKey: queryKeys.adminOverview }),
      queryClient.invalidateQueries({ queryKey: queryKeys.adminAuditLogsRoot })
    ]);
    ElMessage.success('任务已重新投入队列。');
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

const total = computed(() => tasksQuery.data.value?.total ?? 0);

function handleFilterChange() {
  page.value = 1;
}

function getStatusTone(taskStatus: AdminTask['status']) {
  if (taskStatus === 'succeeded') {
    return 'success';
  }
  if (taskStatus === 'failed') {
    return 'danger';
  }
  if (taskStatus === 'running') {
    return 'warning';
  }
  return 'info';
}

function getStatusLabel(taskStatus: AdminTask['status']) {
  return (
    {
      pending: '待处理',
      running: '运行中',
      succeeded: '成功',
      failed: '失败',
      cancelled: '已取消'
    } satisfies Record<AdminTask['status'], string>
  )[taskStatus];
}

async function handleRetry(item: AdminTask) {
  try {
    await ElMessageBox.confirm(
      `任务 ${item.task_type} 将被重新置为待处理状态，是否继续？`,
      '重试任务',
      {
        type: 'warning',
        confirmButtonText: '重试',
        cancelButtonText: '取消'
      }
    );
  } catch {
    return;
  }

  retryMutation.mutate(item.id);
}
</script>

<template>
  <div class="space-y-6">
    <SectionHeading eyebrow="Tasks" title="任务管理" />

    <SurfaceCard>
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
        <el-select
          v-model="status"
          placeholder="任务状态"
          clearable
          @change="handleFilterChange"
        >
          <el-option label="待处理" value="pending" />
          <el-option label="运行中" value="running" />
          <el-option label="成功" value="succeeded" />
          <el-option label="失败" value="failed" />
          <el-option label="已取消" value="cancelled" />
        </el-select>

        <el-select
          v-model="taskType"
          placeholder="任务类型"
          clearable
          @change="handleFilterChange"
        >
          <el-option label="文档摄取" value="document_ingest" />
          <el-option label="知识库重建" value="knowledge_base_reindex" />
          <el-option label="资源清理" value="resource_cleanup" />
        </el-select>

        <div
          class="flex items-center justify-between rounded-[22px] bg-slate-50 px-5 py-4 text-sm text-slate-500"
        >
          <span>共 {{ total }} 个任务</span>
          <span v-if="tasksQuery.isFetching.value">刷新中...</span>
        </div>
      </div>
    </SurfaceCard>

    <SurfaceCard
      v-if="tasksQuery.data.value?.items.length"
      :padded="false"
      class="overflow-hidden"
    >
      <el-table :data="tasksQuery.data.value.items" stripe>
        <el-table-column label="任务" min-width="220">
          <template #default="{ row }">
            <div class="space-y-1 py-1">
              <p class="font-medium text-slate-900">{{ row.task_type }}</p>
              <p class="text-xs text-slate-500">
                {{ row.resource_type }}
                <span v-if="row.resource_id"> / {{ row.resource_id.slice(0, 8) }}</span>
              </p>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="发起者" min-width="160">
          <template #default="{ row }">
            {{ row.user_label || row.user_id || '系统任务' }}
          </template>
        </el-table-column>

        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <AppStatusBadge :tone="getStatusTone(row.status)">
              {{ getStatusLabel(row.status) }}
            </AppStatusBadge>
          </template>
        </el-table-column>

        <el-table-column label="尝试次数" width="110">
          <template #default="{ row }">
            {{ row.attempt_count }} / {{ row.max_attempts }}
          </template>
        </el-table-column>

        <el-table-column label="下次执行" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.next_run_at || row.started_at || row.finished_at) }}
          </template>
        </el-table-column>

        <el-table-column label="错误信息" min-width="260">
          <template #default="{ row }">
            <span :title="row.error_message || ''">
              {{ shortenText(row.error_message || '无', 96) || '无' }}
            </span>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="120" align="right">
          <template #default="{ row }">
            <button
              type="button"
              class="rounded-full bg-slate-100 px-4 py-2 text-sm font-medium text-slate-700 transition hover:bg-slate-200 disabled:bg-slate-100 disabled:text-slate-400"
              :disabled="row.status !== 'failed' || retryMutation.isPending.value"
              @click="handleRetry(row)"
            >
              重试
            </button>
          </template>
        </el-table-column>
      </el-table>
    </SurfaceCard>

    <EmptyStatePanel
      v-else-if="tasksQuery.isError.value"
      title="暂时无法读取任务列表"
      :description="getErrorMessage(tasksQuery.error.value)"
      action-label="重新加载"
      @action="tasksQuery.refetch()"
    />

    <EmptyStatePanel
      v-else
      title="当前没有匹配的任务"
      description="筛选条件下暂时没有可展示的任务。"
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
