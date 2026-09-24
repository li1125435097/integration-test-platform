<template>
  <el-container class="layout-root">
    <el-aside width="220px" class="layout-aside">
      <div class="layout-brand">
        <el-text tag="b" size="large">集成测试平台</el-text>
        <el-text type="info" size="small">Integration Test Platform</el-text>
      </div>
      <el-menu :default-active="activeMenu" router>
        <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
          <el-icon><component :is="iconMap[item.icon]" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container direction="vertical">
      <el-header class="layout-header" height="48px">
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
import { Cpu, Document } from '@element-plus/icons-vue';
import ThemeSwitcher from '@/components/ThemeSwitcher.vue';
import { menuItemsFromRoutes, menuRoutes } from '@/router/menu';

const route = useRoute();
const menuItems = menuItemsFromRoutes(menuRoutes);
const iconMap = { Cpu, Document };

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
}

.layout-aside {
  display: flex;
  flex-direction: column;
  background-color: var(--el-bg-color);
  border-right: 1px solid var(--el-border-color-light);
}

.layout-brand {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  padding: 20px 20px 12px;
  border-bottom: 1px solid var(--el-border-color-lighter);
}

.layout-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 0 20px;
  background-color: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color-light);
}

.layout-main {
  padding: 20px;
  background-color: var(--el-bg-color-page);
}
</style>
