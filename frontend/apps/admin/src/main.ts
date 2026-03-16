import { QueryClient, VueQueryPlugin } from '@tanstack/vue-query';
import 'element-plus/dist/index.css';
import { createPinia } from 'pinia';
import { createApp } from 'vue';

import { configureApiClient, createQueryClientDefaults } from '@private-kb/shared/api/http';
import { useAuthStore } from '@private-kb/shared/auth/useAuthStore';
import { installElementPlus } from '@private-kb/shared/utils/installElementPlus';

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
installElementPlus(app);
app.use(VueQueryPlugin, { queryClient });
app.mount('#app');
