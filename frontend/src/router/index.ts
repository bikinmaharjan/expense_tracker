import { createRouter, createWebHistory } from 'vue-router';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/dashboard',
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: {
        title: 'Dashboard',
        icon: 'mdi-view-dashboard',
      },
    },
    {
      path: '/payments',
      name: 'Payments',
      component: () => import('../views/PaymentsView.vue'),
      meta: {
        title: 'Payments',
        icon: 'mdi-cash',
      },
    },
    {
      path: '/documents',
      name: 'Documents',
      component: () => import('../views/DocumentsView.vue'),
      meta: {
        title: 'Documents',
        icon: 'mdi-file-document',
      },
    },
    {
      path: '/tags',
      name: 'Tags',
      component: () => import('../views/TagsView.vue'),
      meta: {
        title: 'Tags',
        icon: 'mdi-tag-multiple',
      },
    },
  ],
});

// Navigation guard to update document title
router.beforeEach((to, from, next) => {
  document.title = `${to.meta.title} - Expense Tracker` || 'Expense Tracker';
  next();
});

export default router;