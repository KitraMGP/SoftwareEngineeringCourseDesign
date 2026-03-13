<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus';

import { useMutation } from '@tanstack/vue-query';
import { ElMessage } from 'element-plus';
import { reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import { PRODUCT_NAME, authApi } from '@private-kb/shared';
import { getErrorMessage, toFieldErrorMap } from '@private-kb/shared/utils/errors';

const router = useRouter();
const formRef = ref<FormInstance>();
const fieldErrors = ref<Record<string, string>>({});

const form = reactive({
  username: '',
  email: '',
  password: '',
  confirmPassword: '',
  agreement: false
});

const rules: FormRules<typeof form> = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [{ required: true, message: '请输入邮箱地址', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  confirmPassword: [
    {
      validator: (_, value, callback) => {
        if (!value) {
          callback(new Error('请再次输入密码'));
          return;
        }
        if (value !== form.password) {
          callback(new Error('两次输入的密码不一致'));
          return;
        }
        callback();
      },
      trigger: 'blur'
    }
  ],
  agreement: [
    {
      validator: (_, value, callback) => {
        if (!value) {
          callback(new Error('请先确认使用约定'));
          return;
        }
        callback();
      },
      trigger: 'change'
    }
  ]
};

const registerMutation = useMutation({
  mutationFn: () =>
    authApi.register({
      username: form.username,
      email: form.email,
      password: form.password
    }),
  onSuccess: async () => {
    ElMessage.success('注册成功，请使用新账号登录。');
    await router.push({
      name: 'login',
      query: { account: form.email }
    });
  },
  onError: (error) => {
    fieldErrors.value = toFieldErrorMap(error);
    ElMessage.error(getErrorMessage(error, '注册未完成，请检查输入信息。'));
  }
});

async function handleSubmit() {
  if (!formRef.value) {
    return;
  }

  fieldErrors.value = {};
  await formRef.value.validate();
  registerMutation.mutate();
}
</script>

<template>
  <div class="space-y-8">
    <div class="space-y-3">
      <p class="text-xs uppercase tracking-[0.36em] text-slate-400">Create account</p>
      <h1 class="font-serif text-3xl text-slate-900 md:text-[2.3rem]">
        注册 {{ PRODUCT_NAME }}
      </h1>
      <p class="text-sm leading-6 text-slate-600">
        首次加入时先建立你的个人账号，之后即可进入知识库工作台。
      </p>
    </div>

    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" size="large">
      <el-form-item label="用户名" prop="username" :error="fieldErrors.username">
        <el-input v-model="form.username" placeholder="3-50 位字母、数字或下划线" clearable />
      </el-form-item>

      <el-form-item label="邮箱" prop="email" :error="fieldErrors.email">
        <el-input v-model="form.email" placeholder="用于登录和通知" clearable />
      </el-form-item>

      <el-form-item label="密码" prop="password" :error="fieldErrors.password">
        <el-input
          v-model="form.password"
          type="password"
          show-password
          placeholder="至少 8 位，包含字母和数字"
        />
      </el-form-item>

      <el-form-item label="确认密码" prop="confirmPassword">
        <el-input
          v-model="form.confirmPassword"
          type="password"
          show-password
          placeholder="再次输入密码"
          @keyup.enter="handleSubmit"
        />
      </el-form-item>

      <el-form-item prop="agreement">
        <el-checkbox v-model="form.agreement">
          我已了解该工作台面向团队内部知识整理与问答使用。
        </el-checkbox>
      </el-form-item>

      <button
        type="button"
        class="mt-4 flex w-full items-center justify-center rounded-[18px] bg-slate-900 px-5 py-4 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
        :disabled="registerMutation.isPending.value"
        @click="handleSubmit"
      >
        {{ registerMutation.isPending.value ? '创建中...' : '创建账号' }}
      </button>
    </el-form>

    <div class="flex items-center justify-between text-sm text-slate-600">
      <span>已经有账号？</span>
      <router-link class="font-medium text-slate-900 hover:text-brand-night" to="/login">
        返回登录
      </router-link>
    </div>
  </div>
</template>
