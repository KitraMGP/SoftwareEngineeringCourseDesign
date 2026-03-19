<script setup lang="ts">
import type { ProviderConfig } from '@private-kb/shared';
import type { FormInstance, FormRules } from 'element-plus';
import { computed, nextTick, reactive, ref } from 'vue';
import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { ElMessage } from 'element-plus';

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

const queryClient = useQueryClient();

const providersQuery = useQuery({
  queryKey: queryKeys.adminProviderConfigs,
  queryFn: () => adminApi.listProviderConfigs()
});

const provider = computed(() => providersQuery.data.value?.[0] ?? null);
const dialogVisible = ref(false);
const formRef = ref<FormInstance>();
const formModel = reactive({
  api_key: ''
});

const formRules: FormRules = {
  api_key: [{ required: true, message: '请输入新的 API Key', trigger: 'blur' }]
};

const updateKeyMutation = useMutation({
  mutationFn: (apiKey: string) =>
    adminApi.updateProviderApiKey('deepseek', { api_key: apiKey }),
  onSuccess: async (item) => {
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.adminProviderConfigs }),
      queryClient.invalidateQueries({ queryKey: queryKeys.adminAuditLogsRoot })
    ]);
    dialogVisible.value = false;
    resetDialogState();
    ElMessage.success(
      item.api_key_source === 'environment'
        ? '数据库中的 DeepSeek API Key 已保存；当前运行仍优先使用 .env。'
        : 'DeepSeek API Key 已更新。'
    );
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

function getKeySourceLabel(source: ProviderConfig['api_key_source']) {
  return (
    {
      database: '数据库',
      environment: '.env',
      missing: '未配置'
    } satisfies Record<ProviderConfig['api_key_source'], string>
  )[source];
}

function getKeySourceTone(item: ProviderConfig) {
  if (!item.has_api_key) {
    return 'danger';
  }
  if (item.api_key_source === 'database') {
    return 'success';
  }
  return 'info';
}

function getActionLabel(item: ProviderConfig) {
  if (item.api_key_source === 'environment') {
    return '写入数据库备用 Key';
  }
  if (item.has_api_key) {
    return '重新输入覆盖';
  }
  return '配置 API Key';
}

function openKeyDialog() {
  formModel.api_key = '';
  dialogVisible.value = true;
  void nextTick(() => {
    formRef.value?.clearValidate();
  });
}

function resetDialogState() {
  formModel.api_key = '';
  formRef.value?.clearValidate();
}

async function submitKeyForm() {
  const valid = await formRef.value?.validate().catch(() => false);
  if (!valid) {
    return;
  }

  updateKeyMutation.mutate(formModel.api_key.trim());
}
</script>

<template>
  <div class="space-y-6">
    <SectionHeading eyebrow="Providers" title="模型配置" />

    <SurfaceCard v-if="provider">
      <div class="flex flex-col gap-5 xl:flex-row xl:items-start xl:justify-between">
        <div class="space-y-4">
          <div class="flex flex-wrap items-center gap-3">
            <h2 class="font-serif text-3xl text-slate-900">{{ provider.provider }}</h2>
            <AppStatusBadge :tone="provider.is_enabled ? 'success' : 'warning'">
              {{ provider.is_enabled ? '已启用' : '已停用' }}
            </AppStatusBadge>
            <AppStatusBadge :tone="getKeySourceTone(provider)">
              {{ provider.has_api_key ? '已配置 API Key' : '未配置 API Key' }}
            </AppStatusBadge>
          </div>

          <div class="grid gap-4 md:grid-cols-3">
            <div class="rounded-[22px] bg-slate-50 px-5 py-4">
              <p class="text-xs uppercase tracking-[0.24em] text-slate-400">当前来源</p>
              <p class="mt-3 text-sm font-medium text-slate-900">
                {{ getKeySourceLabel(provider.api_key_source) }}
              </p>
            </div>
            <div class="rounded-[22px] bg-slate-50 px-5 py-4">
              <p class="text-xs uppercase tracking-[0.24em] text-slate-400">Base URL</p>
              <p class="mt-3 break-all text-sm font-medium text-slate-900">
                {{ provider.base_url || '未设置' }}
              </p>
            </div>
            <div class="rounded-[22px] bg-slate-50 px-5 py-4">
              <p class="text-xs uppercase tracking-[0.24em] text-slate-400">默认聊天模型</p>
              <p class="mt-3 text-sm font-medium text-slate-900">
                {{ provider.default_chat_model || '未设置' }}
              </p>
            </div>
          </div>

          <div class="rounded-[22px] border border-slate-200 bg-slate-50/80 px-5 py-4">
            <p class="text-sm font-semibold text-slate-900">读取优先级</p>
            <p class="mt-2 text-sm leading-6 text-slate-600">
              系统会优先读取 `.env` 里的 `DEEPSEEK_API_KEY`。只有 `.env`
              没提供时，才会回退读取数据库中的 DeepSeek API Key。
            </p>
            <p class="mt-2 text-sm leading-6 text-slate-600">
              管理后台不会回显旧 key；如需修改数据库值，只能重新输入新的 key 覆盖。
            </p>
            <p v-if="!provider.has_api_key" class="mt-2 text-sm leading-6 text-amber-700">
              当前既没有从 `.env` 检测到 DeepSeek API
              Key，也没有数据库配置；聊天能力会不可用，需要在这里补充。
            </p>
          </div>
        </div>

        <div class="flex shrink-0 flex-col items-start gap-3 xl:items-end">
          <button
            type="button"
            class="rounded-full bg-slate-900 px-5 py-2.5 text-sm font-medium text-white transition hover:bg-slate-800 disabled:bg-slate-300"
            :disabled="updateKeyMutation.isPending.value"
            @click="openKeyDialog"
          >
            {{ getActionLabel(provider) }}
          </button>
          <p class="text-xs text-slate-500">
            最近更新时间：{{ formatDateTime(provider.updated_at) }}
          </p>
        </div>
      </div>
    </SurfaceCard>

    <EmptyStatePanel
      v-else-if="providersQuery.isError.value"
      title="暂时无法读取 DeepSeek 配置"
      :description="getErrorMessage(providersQuery.error.value)"
      action-label="重新加载"
      @action="providersQuery.refetch()"
    />

    <EmptyStatePanel
      v-else
      title="暂时没有可展示的 Provider"
      description="当前默认 DeepSeek 配置尚未加载出来。"
    />

    <el-dialog
      v-model="dialogVisible"
      width="520"
      title="更新 DeepSeek API Key"
      @closed="resetDialogState"
    >
      <div class="space-y-5">
        <div class="rounded-[18px] bg-slate-50 px-4 py-4 text-sm text-slate-600">
          <p class="font-medium text-slate-900">DeepSeek</p>
          <p class="mt-2 leading-6">
            数据库中的 key 仅在 `.env` 未提供 `DEEPSEEK_API_KEY`
            时生效。提交后只保留新输入的值，旧值不会显示也不会返回到页面。
          </p>
        </div>

        <el-form ref="formRef" :model="formModel" :rules="formRules" label-position="top">
          <el-form-item label="新的 API Key" prop="api_key">
            <el-input
              v-model="formModel.api_key"
              clearable
              show-password
              type="password"
              placeholder="请输入新的 DeepSeek API Key"
              autocomplete="new-password"
              @keyup.enter="submitKeyForm"
            />
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <div class="flex justify-end gap-3">
          <button
            type="button"
            class="rounded-full border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 transition hover:border-slate-300 hover:bg-slate-50"
            :disabled="updateKeyMutation.isPending.value"
            @click="dialogVisible = false"
          >
            取消
          </button>
          <button
            type="button"
            class="rounded-full bg-slate-900 px-4 py-2 text-sm font-medium text-white transition hover:bg-slate-800 disabled:bg-slate-300"
            :disabled="updateKeyMutation.isPending.value"
            @click="submitKeyForm"
          >
            {{ updateKeyMutation.isPending.value ? '提交中...' : '保存数据库 Key' }}
          </button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
