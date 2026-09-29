<template>
  <el-container class="layout-root">
    <el-aside width="220px" class="layout-aside">
      <div class="layout-brand">
        <span class="layout-mark" aria-hidden="true">IT</span>
        <div class="layout-brand-text">
          <span class="layout-brand-title">集成测试平台</span>
          <span class="layout-brand-sub">Integration Test</span>
        </div>
      </div>
      <el-menu class="layout-menu" :default-active="activeMenu" router>
        <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
          <el-icon><component :is="iconMap[item.icon]" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container class="layout-body" direction="vertical">
      <el-header class="layout-header" height="56px">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ path: '/scripts' }">首页</el-breadcrumb-item>
          <el-breadcrumb-item v-if="breadcrumbTitle">{{ breadcrumbTitle }}</el-breadcrumb-item>
        </el-breadcrumb>
        <ThemeSwitcher />
      </el-header>
      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { ChromeFilled, Cpu, Document, List } from '@element-plus/icons-vue';
import ThemeSwitcher from '@/components/ThemeSwitcher.vue';
import { menuItemsFromRoutes, menuRoutes } from '@/router/menu';

const route = useRoute();
const menuItems = menuItemsFromRoutes(menuRoutes);
const iconMap = { Cpu, Document, List, ChromeFilled };

const activeMenu = computed(() =>
  route.path.startsWith('/script-editor') ? '/scripts' : route.path
);

const breadcrumbTitle = computed(() => {
  if (route.path.startsWith('/script-editor')) {
    return route.params.id === 'new' ? '新建脚本' : '编辑脚本';
  }
  const hit = menuItems.find((m) => m.path === route.path);
  return hit?.title ?? '';
});
</script>

<style scoped>
.layout-root {
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.layout-aside {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  background-color: var(--el-bg-color);
  border-right: 1px solid var(--el-border-color-light);
}

.layout-brand {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
  padding: 16px 16px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.layout-mark {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  flex-shrink: 0;
  border-radius: 8px;
  background-color: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-size: 13px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.layout-brand-text {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.layout-brand-title {
  color: var(--el-text-color-primary);
  font-size: 15px;
  font-weight: 700;
  line-height: 1.2;
}

.layout-brand-sub {
  color: var(--el-text-color-secondary);
  font-size: 11px;
  line-height: 1.2;
}

.layout-menu {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  border-right: none;
  padding: 8px;
  background-color: transparent;
  --el-menu-bg-color: transparent;
  --el-menu-item-height: 44px;
  --el-menu-hover-bg-color: var(--el-fill-color-light);
}

.layout-menu :deep(.el-menu-item) {
  margin-bottom: 4px;
  border-radius: 8px;
}

.layout-menu :deep(.el-menu-item.is-active) {
  background-color: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  font-weight: 600;
}

.layout-body {
  height: 100%;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}

.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-shrink: 0;
  padding: 0 20px;
  background-color: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-light);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.layout-main {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 20px;
  background-color: var(--el-bg-color-page);
}
</style>
