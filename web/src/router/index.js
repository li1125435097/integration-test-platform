import { createRouter, createWebHashHistory } from 'vue-router';
import MainLayout from '@/layouts/MainLayout.vue';
import { menuRoutes } from './menu';

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: '/',
      component: MainLayout,
      redirect: '/scripts',
      children: menuRoutes
    }
  ]
});

export default router;
