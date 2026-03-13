<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus';

import { useMutation } from '@tanstack/vue-query';
import { ElMessage } from 'element-plus';
import { reactive, ref, watch } from 'vue';

import { SectionHeading, SurfaceCard, UserAvatar, authApi, formatDateTime } from '@private-kb/shared';
import { useAuthStore } from '@private-kb/shared/auth/useAuthStore';
import { getErrorMessage, toFieldErrorMap } from '@private-kb/shared/utils/errors';

const authStore = useAuthStore();
const formRef = ref<FormInstance>();
const fieldErrors = ref<Record<string, string>>({});

const form = reactive({
  nickname: '',
  avatar_url: ''
});

watch(
  () => authStore.user,
  (user) => {
    form.nickname = user?.nickname || '';
    form.avatar_url = user?.avatar_url || '';
  },
  { immediate: true }
);

const rules: FormRules<typeof form> = {
  nickname: [{ max: 100, message: '昵称长度不能超过 100 个字符', trigger: 'blur' }],
  avatar_url: [
    {
      validator: (_, value, callback) => {
        if (!value) {
          callback();
          return;
        }

        try {
          new URL(value);
          callback();
        } catch {
          callback(new Error('请输入有效的头像地址'));
        }
      },
      trigger: 'blur'
    }
  ]
};

const updateMutation = useMutation({
  mutationFn: () =>
    authApi.updateCurrentUser({
      nickname: form.nickname || null,
      avatar_url: form.avatar_url || null
    }),
  onSuccess: async () => {
    await authStore.fetchCurrentUser();
    ElMessage.success('个人资料已更新。');
  },
  onError: (error) => {
    fieldErrors.value = toFieldErrorMap(error);
    ElMessage.error(getErrorMessage(error));
  }
});

async function handleSubmit() {
  if (!formRef.value) {
    return;
  }

  fieldErrors.value = {};
  await formRef.value.validate();
  updateMutation.mutate();
}
</script>

<template>
  <div class="soft-scrollbar flex h-full flex-1 flex-col overflow-y-auto px-5 py-6 lg:px-8">
    <div class="space-y-6">
      <SectionHeading
        eyebrow="Profile"
        title="个人资料"
        description="当前版本可以维护昵称和头像地址，侧栏与个人中心会同步展示。"
      />

      <div class="grid gap-5 xl:grid-cols-[0.9fr_1.1fr]">
        <SurfaceCard>
          <p class="text-xs uppercase tracking-[0.3em] text-slate-400">账户摘要</p>
          <div class="mt-5 flex items-center gap-4 rounded-[24px] bg-slate-50 px-5 py-5">
            <UserAvatar
              :label="authStore.displayName"
              :avatar-url="authStore.user?.avatar_url"
              :size="72"
            />
            <div class="min-w-0">
              <p class="truncate text-lg font-semibold text-slate-900">{{ authStore.displayName }}</p>
              <p class="mt-1 truncate text-sm text-slate-500">{{ authStore.user?.email }}</p>
            </div>
          </div>
          <div class="mt-5 space-y-4">
            <div class="rounded-[22px] bg-slate-50 px-5 py-4">
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">用户名</p>
              <p class="mt-2 text-sm font-medium text-slate-800">{{ authStore.user?.username }}</p>
            </div>
            <div class="rounded-[22px] bg-slate-50 px-5 py-4">
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">邮箱</p>
              <p class="mt-2 text-sm font-medium text-slate-800">{{ authStore.user?.email }}</p>
            </div>
            <div class="rounded-[22px] bg-slate-50 px-5 py-4">
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">角色</p>
              <p class="mt-2 text-sm font-medium text-slate-800">
                {{ authStore.user?.role === 'admin' ? '管理员' : '普通成员' }}
              </p>
            </div>
            <div class="rounded-[22px] bg-slate-50 px-5 py-4">
              <p class="text-xs uppercase tracking-[0.26em] text-slate-400">最近更新</p>
              <p class="mt-2 text-sm font-medium text-slate-800">
                {{ authStore.user?.updated_at ? formatDateTime(authStore.user.updated_at) : '尚无记录' }}
              </p>
            </div>
          </div>
        </SurfaceCard>

        <SurfaceCard>
          <p class="text-xs uppercase tracking-[0.3em] text-slate-400">可编辑信息</p>
          <el-form ref="formRef" :model="form" :rules="rules" label-position="top" class="mt-5">
            <el-form-item label="昵称" prop="nickname" :error="fieldErrors.nickname">
              <el-input v-model="form.nickname" placeholder="用于会话侧栏和个人中心展示" />
            </el-form-item>
            <el-form-item label="头像地址" prop="avatar_url" :error="fieldErrors.avatar_url">
              <el-input
                v-model="form.avatar_url"
                placeholder="填写可公开访问的头像图片 URL"
                clearable
              />
            </el-form-item>
            <p class="text-xs leading-5 text-slate-400">
              当前后端支持保存外部头像 URL；如果留空，将回退为用户名首字母头像。
            </p>
            <button
              type="button"
              class="mt-3 rounded-full bg-slate-900 px-5 py-3 text-sm font-semibold text-white transition hover:bg-slate-800"
              :disabled="updateMutation.isPending.value"
              @click="handleSubmit"
            >
              {{ updateMutation.isPending.value ? '保存中...' : '保存资料' }}
            </button>
          </el-form>
        </SurfaceCard>
      </div>
    </div>
  </div>
</template>
