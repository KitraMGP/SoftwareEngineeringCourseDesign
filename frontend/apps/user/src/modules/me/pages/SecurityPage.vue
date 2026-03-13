<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus';

import { useMutation } from '@tanstack/vue-query';
import { ElMessage } from 'element-plus';
import { reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { SectionHeading, SurfaceCard, authApi } from '@private-kb/shared';
import { useAuthStore } from '@private-kb/shared/auth/useAuthStore';
import { getErrorMessage, toFieldErrorMap } from '@private-kb/shared/utils/errors';

const router = useRouter();
const authStore = useAuthStore();

const formRef = ref<FormInstance>();
const fieldErrors = ref<Record<string, string>>({});

const form = reactive({
  old_password: '',
  new_password: '',
  confirm_password: ''
});

const rules: FormRules<typeof form> = {
  old_password: [{ required: true, message: '请输入当前密码', trigger: 'blur' }],
  new_password: [{ required: true, message: '请输入新密码', trigger: 'blur' }],
  confirm_password: [
    {
      validator: (_, value, callback) => {
        if (!value) {
          callback(new Error('请再次输入新密码'));
          return;
        }
        if (value !== form.new_password) {
          callback(new Error('两次输入的新密码不一致'));
          return;
        }
        callback();
      },
      trigger: 'blur'
    }
  ]
};

const changePasswordMutation = useMutation({
  mutationFn: () =>
    authApi.changePassword({
      old_password: form.old_password,
      new_password: form.new_password
    }),
  onSuccess: async () => {
    authStore.clearAuth();
    ElMessage.success('密码已更新，请重新登录。');
    await router.push({ name: 'login' });
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
  changePasswordMutation.mutate();
}
</script>

<template>
  <div class="soft-scrollbar flex h-full flex-1 flex-col overflow-y-auto px-5 py-6 lg:px-8">
    <div class="space-y-6">
      <SectionHeading
        eyebrow="Security"
        title="安全设置"
        description="修改密码成功后将清理当前登录态，并要求重新登录。"
      />

      <div class="grid gap-5 xl:grid-cols-[1fr_1fr]">
        <SurfaceCard>
          <p class="text-xs uppercase tracking-[0.3em] text-slate-400">安全提醒</p>
          <div class="mt-5 space-y-3 text-sm leading-7 text-slate-500">
            <p>新密码需要至少 8 位，并同时包含字母和数字。</p>
            <p>修改密码后，当前 refresh 会话会被后端回收，需要重新登录才能继续访问工作区。</p>
            <p>如果你在多个设备上登录，建议同步确认是否需要重新建立新会话。</p>
          </div>
        </SurfaceCard>

        <SurfaceCard>
          <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
            <el-form-item label="当前密码" prop="old_password" :error="fieldErrors.old_password">
              <el-input v-model="form.old_password" type="password" show-password />
            </el-form-item>
            <el-form-item label="新密码" prop="new_password" :error="fieldErrors.new_password">
              <el-input v-model="form.new_password" type="password" show-password />
            </el-form-item>
            <el-form-item label="确认新密码" prop="confirm_password">
              <el-input
                v-model="form.confirm_password"
                type="password"
                show-password
                @keyup.enter="handleSubmit"
              />
            </el-form-item>
            <button
              type="button"
              class="mt-3 rounded-full bg-slate-900 px-5 py-3 text-sm font-semibold text-white transition hover:bg-slate-800"
              :disabled="changePasswordMutation.isPending.value"
              @click="handleSubmit"
            >
              {{ changePasswordMutation.isPending.value ? '更新中...' : '更新密码' }}
            </button>
          </el-form>
        </SurfaceCard>
      </div>
    </div>
  </div>
</template>
