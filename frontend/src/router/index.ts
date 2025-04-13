import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'files',
      component: () => import('../views/FileList.vue'),
      meta: {
        title: '文件管理'
      }
    },
    {
      path: '/vnc',
      name: 'vnc',
      component: () => import('../views/VncView.vue'),
      meta: {
        title: 'VNC远程控制'
      }
    }
  ]
})

export default router