import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query';
import ElementPlus from 'element-plus';
import zhCn from 'element-plus/es/locale/lang/zh-cn';
import 'element-plus/dist/index.css';
import { createPinia } from 'pinia';
import { createApp } from 'vue';

import { configureApiClient, createQueryClientDefaults } from '@private-kb/shared/api/http';
import { useAuthStore } from '@private-kb/shared/auth/useAuthStore';

import App from './App.vue';
import router, { installRouterGuards } from './router';
import './styles/index.scss';

const app = createApp(App);
const pinia = createPinia();
const queryClient = new QueryClient({
  defaultOptions: createQueryClientDefaults()
});

app.use(pinia);

const authStore = useAuthStore(pinia);

configureApiClient({
  getAccessToken: () => authStore.accessToken,
  refreshAccessToken: () => authStore.refreshAccessToken(),
  clearAuth: () => authStore.clearAuth(),
  onUnauthorized: () => {
    if (router.currentRoute.value.name !== 'admin-login') {
      router.push({ name: 'admin-login' });
    }
  }
});

installRouterGuards(router, pinia);

app.use(router);
app.use(ElementPlus, { locale: zhCn });
app.use(VueQueryPlugin, { queryClient });
app.mount('#app');
