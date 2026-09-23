<template>
  <el-dropdown trigger="click" popper-class="theme-dropdown" @command="onSelect">
    <el-button :icon="Brush" round>
      {{ currentName }}
      <el-icon class="el-icon--right"><ArrowDown /></el-icon>
    </el-button>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item
          v-for="theme in themes"
          :key="theme.id"
          :command="theme.id"
          :class="{ 'theme-item-active': currentId === theme.id }"
        >
          <span class="theme-swatch" :style="{ backgroundColor: theme.primary }" />
          <span class="theme-label">
            <span class="theme-name">{{ theme.name }}</span>
            <el-text v-if="theme.hint" type="info" size="small" class="theme-hint">{{ theme.hint }}</el-text>
          </span>
          <el-icon v-if="currentId === theme.id" class="theme-check"><Check /></el-icon>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup>
import { computed, ref, onMounted } from 'vue';
import { Brush, ArrowDown, Check } from '@element-plus/icons-vue';
import { applyTheme, loadStoredThemeId } from '@/theme/applyTheme';
import { themes, getThemeById } from '@/theme/themes';

const currentId = ref(loadStoredThemeId());
const currentName = computed(() => getThemeById(currentId.value)?.name ?? '主题');

function onSelect(id) {
  currentId.value = applyTheme(id);
}

onMounted(() => {
  currentId.value = applyTheme(currentId.value);
});
</script>

<style>
.theme-dropdown .el-dropdown-menu__item {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 200px;
  line-height: 1.3;
  padding-top: 8px;
  padding-bottom: 8px;
}

.theme-dropdown .el-dropdown-menu__item.theme-item-active {
  background-color: var(--el-color-primary);
  color: #fff;
}

.theme-dropdown .el-dropdown-menu__item.theme-item-active .theme-hint,
.theme-dropdown .el-dropdown-menu__item.theme-item-active .el-text {
  color: rgba(255, 255, 255, 0.85) !important;
}

.theme-dropdown .el-dropdown-menu__item.theme-item-active .theme-check {
  color: #fff;
}

.theme-dropdown .theme-swatch {
  width: 14px;
  height: 14px;
  border-radius: var(--el-border-radius-small);
  flex-shrink: 0;
  box-shadow: inset 0 0 0 1px var(--el-border-color);
}

.theme-dropdown .theme-label {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex: 1;
  min-width: 0;
}

.theme-dropdown .theme-name {
  font-size: var(--el-font-size-base);
}

.theme-dropdown .theme-hint {
  display: block;
}

.theme-dropdown .theme-check {
  margin-left: auto;
  flex-shrink: 0;
}
</style>
