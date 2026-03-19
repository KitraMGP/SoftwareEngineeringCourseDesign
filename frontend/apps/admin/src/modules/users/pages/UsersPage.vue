<script setup lang="ts">
import type { AdminUser } from '@private-kb/shared';
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
  useAuthStore
} from '@private-kb/shared';
import { getErrorMessage } from '@private-kb/shared/utils/errors';

const authStore = useAuthStore();
const queryClient = useQueryClient();

const pageSize = 20;
const page = ref(1);
const keyword = ref('');
const searchKeyword = ref('');
const status = ref('');
const role = ref('');

const usersQuery = useQuery({
  queryKey: computed(() =>
    queryKeys.adminUsers({
      page: page.value,
      size: pageSize,
      keyword: searchKeyword.value,
      status: status.value,
      role: role.value
    })
  ),
  queryFn: () =>
    adminApi.listUsers({
      page: page.value,
      size: pageSize,
      keyword: searchKeyword.value || undefined,
      status: status.value || undefined,
      role: role.value || undefined
    })
});

const freezeMutation = useMutation({
  mutationFn: (userId: string) => adminApi.freezeUser(userId),
  onSuccess: async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.adminUsersRoot }),
      queryClient.invalidateQueries({ queryKey: queryKeys.adminOverview }),
      queryClient.invalidateQueries({ queryKey: queryKeys.adminAuditLogsRoot })
    ]);
    ElMessage.success('账号已冻结。');
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

const unfreezeMutation = useMutation({
  mutationFn: (userId: string) => adminApi.unfreezeUser(userId),
  onSuccess: async () => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.adminUsersRoot }),
      queryClient.invalidateQueries({ queryKey: queryKeys.adminOverview }),
      queryClient.invalidateQueries({ queryKey: queryKeys.adminAuditLogsRoot })
    ]);
    ElMessage.success('账号已恢复。');
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

const total = computed(() => usersQuery.data.value?.total ?? 0);

function submitSearch() {
  page.value = 1;
  searchKeyword.value = keyword.value.trim();
}

function handleFilterChange() {
  page.value = 1;
}

function getStatusTone(userStatus: AdminUser['status']) {
  return userStatus === 'active' ? 'success' : 'warning';
}

function getRoleLabel(userRole: AdminUser['role']) {
  return userRole === 'admin' ? '管理员' : '普通用户';
}

function isMutating() {
  return freezeMutation.isPending.value || unfreezeMutation.isPending.value || false;
}

async function handleToggleStatus(item: AdminUser) {
  const nextAction = item.status === 'active' ? '冻结' : '恢复';

  try {
    await ElMessageBox.confirm(
      `${nextAction}账号 ${item.username} 后会立即影响其后续登录和接口访问，是否继续？`,
      `${nextAction}账号`,
      {
        type: item.status === 'active' ? 'warning' : 'info',
        confirmButtonText: nextAction,
        cancelButtonText: '取消'
      }
    );
  } catch {
    return;
  }

  if (item.status === 'active') {
    freezeMutation.mutate(item.id);
    return;
  }

  unfreezeMutation.mutate(item.id);
}
</script>

<template>
  <div class="space-y-6">
    <SectionHeading eyebrow="Users" title="用户管理" />

    <SurfaceCard>
      <div class="grid gap-4 xl:grid-cols-[minmax(0,1.4fr),180px,180px]">
        <div class="flex gap-3">
          <el-input
            v-model="keyword"
            clearable
            placeholder="按用户名、邮箱或昵称搜索"
            @keyup.enter="submitSearch"
          />
          <button
            type="button"
            class="min-w-[92px] rounded-full border border-slate-200 px-4 py-2.5 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
            @click="submitSearch"
          >
            搜索
          </button>
        </div>

        <el-select
          v-model="status"
          placeholder="账号状态"
          clearable
          @change="handleFilterChange"
        >
          <el-option label="活跃" value="active" />
          <el-option label="冻结" value="frozen" />
        </el-select>

        <el-select v-model="role" placeholder="用户角色" clearable @change="handleFilterChange">
          <el-option label="普通用户" value="user" />
          <el-option label="管理员" value="admin" />
        </el-select>
      </div>

      <div class="mt-4 flex items-center justify-between text-sm text-slate-500">
        <span>共 {{ total }} 个账号</span>
        <span v-if="usersQuery.isFetching.value">刷新中...</span>
      </div>
    </SurfaceCard>

    <SurfaceCard
      v-if="usersQuery.data.value?.items.length"
      :padded="false"
      class="overflow-hidden"
    >
      <el-table :data="usersQuery.data.value.items" stripe>
        <el-table-column label="账号" min-width="260">
          <template #default="{ row }">
            <div class="space-y-1 py-1">
              <div class="flex items-center gap-2">
                <span class="font-medium text-slate-900">{{ row.username }}</span>
                <span
                  v-if="row.id === authStore.user?.id"
                  class="rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-500"
                >
                  当前账号
                </span>
              </div>
              <p class="text-sm text-slate-500">{{ row.email }}</p>
              <p v-if="row.nickname" class="text-xs text-slate-400">昵称：{{ row.nickname }}</p>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            {{ getRoleLabel(row.role) }}
          </template>
        </el-table-column>

        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <AppStatusBadge :tone="getStatusTone(row.status)">
              {{ row.status === 'active' ? '活跃' : '冻结' }}
            </AppStatusBadge>
          </template>
        </el-table-column>

        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>

        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">
            {{ formatDateTime(row.updated_at) }}
          </template>
        </el-table-column>

        <el-table-column label="操作" width="140" align="right">
          <template #default="{ row }">
            <button
              type="button"
              class="rounded-full px-4 py-2 text-sm font-medium transition"
              :class="
                row.status === 'active'
                  ? 'bg-amber-100 text-amber-800 hover:bg-amber-200 disabled:bg-slate-100 disabled:text-slate-400'
                  : 'bg-emerald-100 text-emerald-800 hover:bg-emerald-200 disabled:bg-slate-100 disabled:text-slate-400'
              "
              :disabled="isMutating() || row.id === authStore.user?.id"
              @click="handleToggleStatus(row)"
            >
              {{ row.status === 'active' ? '冻结' : '恢复' }}
            </button>
          </template>
        </el-table-column>
      </el-table>
    </SurfaceCard>

    <EmptyStatePanel
      v-else-if="usersQuery.isError.value"
      title="暂时无法读取用户列表"
      :description="getErrorMessage(usersQuery.error.value)"
      action-label="重新加载"
      @action="usersQuery.refetch()"
    />

    <EmptyStatePanel v-else title="没有匹配的账号" description="当前筛选条件下没有找到用户。" />

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
