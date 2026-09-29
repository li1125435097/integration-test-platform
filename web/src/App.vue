<template>
  <router-view />
</template>

<style>
html,
body,
#app {
  height: 100%;
  margin: 0;
  overflow: hidden;
}

body {
  font-family: var(--el-font-family);
  background-color: var(--el-bg-color-page);
  color: var(--el-text-color-primary);
}

.page-fill {
  flex: 1;
  height: 100%;
  min-height: 0;
  overflow: hidden;
  box-sizing: border-box;
}

.table-card {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.table-card > .el-card__body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
</style>
