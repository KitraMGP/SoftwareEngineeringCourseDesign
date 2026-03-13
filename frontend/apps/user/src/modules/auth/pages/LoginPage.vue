<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus';

import { useMutation } from '@tanstack/vue-query';
import { ElMessage } from 'element-plus';
import { reactive, ref, watch } from 'vue';
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

watch(
  () => route.query.account,
  (account) => {
    if (typeof account === 'string' && account && !form.account) {
      form.account = account;
    }
  },
  { immediate: true }
);

const rules: FormRules<typeof form> = {
  account: [{ required: true, message: '请输入用户名或邮箱', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
};

const loginMutation = useMutation({
  mutationFn: () => authStore.login(form),
  onSuccess: async () => {
    ElMessage.success('欢迎回来，已恢复你的工作区。');
    await router.push((route.query.redirect as string) || '/');
  },
  onError: (error) => {
    fieldErrors.value = toFieldErrorMap(error);
    ElMessage.error(getErrorMessage(error, '登录失败，请检查账号和密码。'));
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
      <p class="text-xs uppercase tracking-[0.36em] text-slate-400">Welcome back</p>
      <h1 class="font-serif text-3xl text-slate-900 md:text-[2.3rem]">
        登录 {{ PRODUCT_NAME }}
      </h1>
      <p class="text-sm leading-6 text-slate-600">
        继续访问聊天工作区、知识库文档和你的个人设置。
      </p>
    </div>

    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" size="large">
      <el-form-item
        label="账号"
        prop="account"
        :error="fieldErrors.account"
      >
        <el-input
          v-model="form.account"
          placeholder="用户名或邮箱"
          autocomplete="username"
          clearable
        />
      </el-form-item>

      <el-form-item
        label="密码"
        prop="password"
        :error="fieldErrors.password"
      >
        <el-input
          v-model="form.password"
          type="password"
          show-password
          placeholder="输入登录密码"
          autocomplete="current-password"
          @keyup.enter="handleSubmit"
        />
      </el-form-item>

      <button
        type="button"
        class="mt-4 flex w-full items-center justify-center rounded-[18px] bg-slate-900 px-5 py-4 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
        :disabled="loginMutation.isPending.value"
        @click="handleSubmit"
      >
        {{ loginMutation.isPending.value ? '登录中...' : '进入工作区' }}
      </button>
    </el-form>

    <div class="flex items-center justify-between text-sm text-slate-600">
      <span>还没有账号？</span>
      <router-link class="font-medium text-slate-900 hover:text-brand-night" to="/register">
        立即注册
      </router-link>
    </div>
  </div>
</template>
