<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus';

import { useMutation } from '@tanstack/vue-query';
import { ElMessage } from 'element-plus';
import { reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { PRODUCT_NAME } from '@private-kb/shared/constants/app';
import { useAuthStore } from '@private-kb/shared/auth/useAuthStore';
import { getErrorMessage, toFieldErrorMap } from '@private-kb/shared/utils/errors';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const formRef = ref<FormInstance>();
const fieldErrors = ref<Record<string, string>>({});

const form = reactive({
  account: '',
  password: ''
});

const rules: FormRules<typeof form> = {
  account: [{ required: true, message: '请输入后台账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
};

const loginMutation = useMutation({
  mutationFn: () => authStore.login(form),
  onSuccess: async () => {
    if (!authStore.isAdmin) {
      await router.push({ name: 'forbidden' });
      return;
    }

    ElMessage.success('后台权限校验通过。');
    await router.push((route.query.redirect as string) || '/dashboard');
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
  loginMutation.mutate();
}
</script>

<template>
  <div class="space-y-8">
    <div class="space-y-3">
      <p class="text-xs uppercase tracking-[0.34em] text-slate-400">Admin sign in</p>
      <h1 class="font-serif text-3xl text-slate-900">{{ PRODUCT_NAME }} 后台登录</h1>
      <p class="text-sm leading-6 text-slate-500">
        使用具备管理员角色的账号登录。非管理员账号会被保留登录态，但无法进入后台工作区。
      </p>
    </div>

    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" size="large">
      <el-form-item label="账号" prop="account" :error="fieldErrors.account">
        <el-input v-model="form.account" placeholder="用户名或邮箱" />
      </el-form-item>
      <el-form-item label="密码" prop="password" :error="fieldErrors.password">
        <el-input
          v-model="form.password"
          type="password"
          show-password
          placeholder="输入后台登录密码"
          @keyup.enter="handleSubmit"
        />
      </el-form-item>

      <button
        type="button"
        class="mt-4 flex w-full items-center justify-center rounded-[18px] bg-slate-950 px-5 py-4 text-sm font-semibold text-white transition hover:bg-slate-900 disabled:cursor-not-allowed disabled:bg-slate-400"
        :disabled="loginMutation.isPending.value"
        @click="handleSubmit"
      >
        {{ loginMutation.isPending.value ? '验证中...' : '进入后台' }}
      </button>
    </el-form>
  </div>
</template>
