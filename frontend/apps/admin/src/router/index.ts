import type { Pinia } from 'pinia';
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';

import { PRODUCT_NAME } from '@private-kb/shared/constants/app';
import { useAuthStore } from '@private-kb/shared/auth/useAuthStore';

import AdminAuthLayout from '../layouts/AdminAuthLayout.vue';
import AdminShellLayout from '../layouts/AdminShellLayout.vue';
import AccessDeniedPage from '../modules/access/pages/AccessDeniedPage.vue';
import AdminLoginPage from '../modules/auth/pages/AdminLoginPage.vue';
import AuditPage from '../modules/audit/pages/AuditPage.vue';
import DashboardPage from '../modules/dashboard/pages/DashboardPage.vue';
import ProvidersPage from '../modules/providers/pages/ProvidersPage.vue';
import GeneralSettingsPage from '../modules/settings/pages/GeneralSettingsPage.vue';
import QuotaSettingsPage from '../modules/settings/pages/QuotaSettingsPage.vue';
import TasksPage from '../modules/tasks/pages/TasksPage.vue';
import UsersPage from '../modules/users/pages/UsersPage.vue';

declare module 'vue-router' {
  interface RouteMeta {
    title?: string;
    guestOnly?: boolean;
    requiresAuth?: boolean;
    requiresAdmin?: boolean;
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: AdminShellLayout,
    meta: { requiresAuth: true, requiresAdmin: true },
    children: [
      {
        path: '',
        redirect: '/dashboard'
      },
      {
        path: 'dashboard',
        name: 'dashboard',
        component: DashboardPage,
        meta: { title: 'Dashboard' }
      },
      {
        path: 'users',
        name: 'users',
        component: UsersPage,
        meta: { title: '用户管理' }
      },
      {
        path: 'providers',
        name: 'providers',
        component: ProvidersPage,
        meta: { title: '模型配置' }
      },
      {
        path: 'tasks',
        name: 'tasks',
        component: TasksPage,
        meta: { title: '任务管理' }
      },
      {
        path: 'audit-logs',
        name: 'audit-logs',
        component: AuditPage,
        meta: { title: '审计日志' }
      },
      {
        path: 'settings/general',
        name: 'settings-general',
        component: GeneralSettingsPage,
        meta: { title: '系统设置' }
      },
      {
        path: 'settings/quota',
        name: 'settings-quota',
        component: QuotaSettingsPage,
        meta: { title: '配额策略' }
      }
    ]
  },
  {
    path: '/',
    component: AdminAuthLayout,
    children: [
      {
        path: 'login',
        name: 'admin-login',
        component: AdminLoginPage,
        meta: { title: '后台登录', guestOnly: true }
      }
    ]
  },
  {
    path: '/forbidden',
    name: 'forbidden',
    component: AccessDeniedPage,
    meta: { title: '访问受限', requiresAuth: true }
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/dashboard'
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 })
});

function buildTitle(title?: string): string {
  return title ? `${title} | ${PRODUCT_NAME} Admin` : `${PRODUCT_NAME} Admin`;
}

export function installRouterGuards(instance: typeof router, pinia: Pinia) {
  instance.beforeEach(async (to) => {
    const authStore = useAuthStore(pinia);

    if (!authStore.isInitialized) {
      await authStore.bootstrap();
    }

    const requiresAuth = to.matched.some((record) => record.meta.requiresAuth);
    const requiresAdmin = to.matched.some((record) => record.meta.requiresAdmin);
    const guestOnly = to.matched.some((record) => record.meta.guestOnly);

    if (requiresAuth && !authStore.isAuthenticated) {
      return {
        name: 'admin-login',
        query: {
          redirect: to.fullPath
        }
      };
    }

    if (guestOnly && authStore.isAuthenticated) {
      return authStore.isAdmin ? { name: 'dashboard' } : { name: 'forbidden' };
    }

    if (requiresAdmin && !authStore.isAdmin) {
      return { name: 'forbidden' };
    }

    return true;
  });

  instance.afterEach((to) => {
    document.title = buildTitle(to.meta.title);
  });
}

export default router;
