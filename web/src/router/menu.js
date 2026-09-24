/**
 * 侧栏菜单：由路由 meta.menu 推导（Vue Router 推荐做法）。
 */
export const menuRoutes = [
  {
    path: '/scripts',
    name: 'scripts',
    component: () => import('@/views/scripts/ScriptList.vue'),
    meta: {
      menu: {
        title: '脚本管理',
        icon: 'Document'
      }
    }
  },
  {
    path: '/interpreters',
    name: 'interpreters',
    component: () => import('@/views/interpreters/InterpreterList.vue'),
    meta: {
      menu: {
        title: '解释器管理',
        icon: 'Cpu'
      }
    }
  },
  {
    path: '/script-editor/:id',
    name: 'script-editor',
    component: () => import('@/views/scripts/ScriptEdit.vue'),
    meta: {
      title: '脚本编辑',
      hidden: true
    }
  }
];

export function menuItemsFromRoutes(routes) {
  return routes
    .filter((r) => r.meta?.menu && !r.meta?.hidden)
    .map((r) => ({
      path: r.path,
      name: r.name,
      title: r.meta.menu.title,
      icon: r.meta.menu.icon
    }));
}
