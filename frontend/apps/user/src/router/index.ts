import type { Pinia } from 'pinia';
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router';

import { PRODUCT_NAME } from '@private-kb/shared/constants/app';
import { useAuthStore } from '@private-kb/shared/auth/useAuthStore';

import AuthLayout from '../layouts/AuthLayout.vue';
import UserWorkspaceLayout from '../layouts/UserWorkspaceLayout.vue';
import LoginPage from '../modules/auth/pages/LoginPage.vue';
import RegisterPage from '../modules/auth/pages/RegisterPage.vue';
import AboutPage from '../modules/about/pages/AboutPage.vue';
import ChatHomePage from '../modules/chat/pages/ChatHomePage.vue';
import SessionPage from '../modules/chat/pages/SessionPage.vue';
import KnowledgeBaseDetailPage from '../modules/knowledge-base/pages/KnowledgeBaseDetailPage.vue';
import KnowledgeBaseListPage from '../modules/knowledge-base/pages/KnowledgeBaseListPage.vue';
import ProfilePage from '../modules/me/pages/ProfilePage.vue';
import SecurityPage from '../modules/me/pages/SecurityPage.vue';

declare module 'vue-router' {
  interface RouteMeta {
    title?: string;
    guestOnly?: boolean;
    requiresAuth?: boolean;
  }
}

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    component: UserWorkspaceLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'chat-home',
        component: ChatHomePage,
        meta: { title: '聊天工作区' }
      },
      {
        path: 'sessions/:sessionId',
        name: 'session-detail',
        component: SessionPage,
        meta: { title: '会话详情' }
      },
      {
        path: 'knowledge-bases',
        name: 'knowledge-bases',
        component: KnowledgeBaseListPage,
        meta: { title: '知识库' }
      },
      {
        path: 'knowledge-bases/:kbId',
        name: 'knowledge-base-detail',
        component: KnowledgeBaseDetailPage,
        meta: { title: '知识库详情' }
      },
      {
        path: 'about',
        name: 'about',
        component: AboutPage,
        meta: { title: '关于产品' }
      },
      {
        path: 'me/profile',
        name: 'me-profile',
        component: ProfilePage,
        meta: { title: '个人资料' }
      },
      {
        path: 'me/security',
        name: 'me-security',
        component: SecurityPage,
        meta: { title: '安全设置' }
      }
    ]
  },
  {
    path: '/',
    component: AuthLayout,
    children: [
      {
        path: 'login',
        name: 'login',
        component: LoginPage,
        meta: { title: '登录', guestOnly: true }
      },
      {
        path: 'register',
        name: 'register',
        component: RegisterPage,
        meta: { title: '注册', guestOnly: true }
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/'
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior: () => ({ top: 0 })
});

function buildTitle(title?: string): string {
  return title ? `${title} | ${PRODUCT_NAME}` : PRODUCT_NAME;
}

export function installRouterGuards(instance: typeof router, pinia: Pinia) {
  instance.beforeEach(async (to) => {
    const authStore = useAuthStore(pinia);

    if (!authStore.isInitialized) {
      await authStore.bootstrap();
    }

    const requiresAuth = to.matched.some((record) => record.meta.requiresAuth);
    const guestOnly = to.matched.some((record) => record.meta.guestOnly);

    if (requiresAuth && !authStore.isAuthenticated) {
      return {
        name: 'login',
        query: {
          redirect: to.fullPath
        }
      };
    }

    if (guestOnly && authStore.isAuthenticated) {
      return { name: 'chat-home' };
    }

    return true;
  });

  instance.afterEach((to) => {
    document.title = buildTitle(to.meta.title);
  });
}

export default router;
