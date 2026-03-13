<script setup lang="ts">
import { EmptyStatePanel } from '@private-kb/shared';
</script>

<template>
  <div class="flex min-h-screen items-center justify-center px-5 py-10">
    <div class="w-full max-w-3xl">
      <EmptyStatePanel
        title="当前账号没有后台权限"
        description="你已经完成登录，但该账号并不具备管理员角色。后端会继续把普通用户鉴权与后台鉴权分开处理。"
      />
    </div>
  </div>
</template>
