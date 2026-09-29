<template>
  <div class="page-block">
    <div class="page-header">
      <el-row>
        <el-text tag="h2" size="large">内核测试</el-text>
      </el-row>
    </div>

    <el-card shadow="hover" class="page-card">
      <el-form label-width="88px" @submit.prevent>
        <section class="form-section">
          <div class="form-section-header">
            <el-text type="info">内核选择（扫描本机 AppData\Roaming 下开发、测试、生产目录中的内核）</el-text>
            <el-button :icon="Refresh" circle :loading="kernelLoading" @click="loadKernels" />
          </div>
          <el-form-item label="来源">
            <div class="source-row">
              <el-checkbox
                :model-value="allSourcesChecked"
                :indeterminate="sourcesIndeterminate"
                @change="onAllSourcesChange"
              >
                全部
              </el-checkbox>
              <el-checkbox-group v-model="sources">
                <el-checkbox value="dev">开发</el-checkbox>
                <el-checkbox value="test">测试</el-checkbox>
                <el-checkbox value="prod">生产</el-checkbox>
              </el-checkbox-group>
            </div>
          </el-form-item>
          <el-form-item label="内核">
            <el-select
              v-model="selectedKernels"
              multiple
              filterable
              clearable
              placeholder="选择内核"
              no-data-text="当前来源下没有内核"
              :loading="kernelLoading"
              style="width: 100%"
            >
              <el-option
                v-for="item in filteredKernels"
                :key="item.path"
                :label="item.name"
                :value="item.path"
              />
            </el-select>
          </el-form-item>
        </section>

        <el-divider />

        <section class="form-section">
          <div class="form-section-header">
            <el-text type="info">启动参数选择（选择预设启动参数，或追加自定义参数）</el-text>
          </div>
          <el-form-item label="预设参数">
            <el-select
              v-model="selectedFlags"
              multiple
              filterable
              clearable
              placeholder="选择启动参数"
              style="width: 100%"
            >
              <el-option v-for="flag in launchFlags" :key="flag" :label="flag" :value="flag" />
            </el-select>
          </el-form-item>
          <el-form-item label="自定义">
            <div class="args-editor">
              <div v-for="(_, index) in customFlags" :key="index" class="args-row">
                <el-input v-model="customFlags[index]" placeholder="输入自定义启动参数" clearable />
                <el-button type="danger" link :icon="Delete" @click="removeCustomFlag(index)">删除</el-button>
              </div>
              <el-button type="primary" link :icon="Plus" @click="addCustomFlag">添加参数</el-button>
            </div>
          </el-form-item>
        </section>

        <el-divider />

        <section class="form-section">
          <div class="form-section-header">
            <el-space>
              <el-text type="info">指纹选择（选择指纹文件，或内置示例）</el-text>
              <el-button :icon="FolderOpened" @click="pickFingerprint">打开</el-button>
              <el-text type="info" size="small">{{ fingerprintFileName || '内置示例' }}</el-text>
            </el-space>
            <el-switch
              v-model="treeMode"
              inline-prompt
              active-text="树形"
              inactive-text="编辑"
              :before-change="guardTreeMode"
            />
          </div>
          <input
            ref="fileInput"
            class="file-input"
            type="file"
            accept="application/json,.json"
            @change="onFingerprintFile"
          />
          <el-form-item label="内容">
            <el-input
              v-show="!treeMode"
              v-model="fingerprintText"
              type="textarea"
              :rows="18"
              resize="vertical"
              class="fingerprint-editor"
              placeholder="指纹 JSON"
            />
            <div v-show="treeMode" class="fingerprint-tree">
              <el-tree
                :data="treeNodes"
                node-key="id"
                :props="{ label: 'label', children: 'children' }"
                :expand-on-click-node="true"
              >
                <template #default="{ data }">
                  <span class="tree-label">
                    <span class="tree-key">{{ data.label }}</span>
                    <span class="tree-summary">{{ data.summary }}</span>
                  </span>
                </template>
              </el-tree>
            </div>
          </el-form-item>
        </section>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { Delete, FolderOpened, Plus, Refresh } from '@element-plus/icons-vue';
import * as kernelsApi from '@/api/kernels';

const SOURCE_IDS = ['dev', 'test', 'prod'];

const launchFlags = [
  '--enable-features=NetworkServiceInProcess2',
  '--disable-background-networking',
  '--enable-features=NetworkService,NetworkServiceInProcess',
  '--disable-background-timer-throttling',
  '--disable-breakpad',
  '--disable-client-side-phishing-detection',
  '--disable-component-extensions-with-background-pages',
  '--disable-default-apps',
  '--disable-dev-shm-usage',
  '--disable-extensions',
  '--disable-features=Translate,AcceptCHFrame,MediaRouter,OptimizationHints,Prerender2',
  '--disable-hang-monitor',
  '--disable-ipc-flooding-protection',
  '--disable-renderer-backgrounding',
  '--disable-sync',
  '--enable-automation'
];

const defaultFingerprint = {
  Resource: {
    IDS_PRODUCT_NAME: 'BitBrowser',
    'IDS_PRODUCT_NAME@zh-CN': '比特浏览器',
    IDS_ABOUT_VERSION_COMPANY_NAME: 'BitBrowser',
    'IDS_ABOUT_VERSION_COMPANY_NAME@zh-CN': '比特浏览器',
    IDS_ABOUT_VERSION_COPYRIGHT:
      'Copyright ©2018-2025 HongKong Bit-Internet Technology Limited. All rights reserved.',
    'IDS_ABOUT_VERSION_COPYRIGHT@zh-CN': '版权所有 2025 北京元宇蓝图科技有限公司，保留所有权利。'
  },
  disableDevTools: false,
  blockExtensionInstall: false,
  isMobile: false,
  uiLang: '',
  touchEmulatorCursorColor: '',
  appLocale: 'de',
  acceptLang: 'de-DE,de,en-US,en',
  taskbarString: '22',
  toolbarString: '22',
  doNotTrack: false,
  userAgent:
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36',
  battery: 100,
  navigator: {
    platform: 'Win32',
    hardwareConcurrency: 24,
    deviceMemory: 8,
    maxTouchPoints: 0,
    vendor: 'Google Inc.'
  },
  UserAgentArchitecture: 'x86',
  UserAgentBitness: 64,
  UserAgentIsWow64: false,
  UserAgentIsMobile: false,
  UserAgentModel: '',
  userAgentVersion: '150.0.7871.101',
  UserAgentPlatform: 'Windows',
  UserAgentPlatformVersion: '10.0.0',
  maxImageBytes: null
};

const kernelLoading = ref(false);
const kernels = ref([]);
const sources = ref([...SOURCE_IDS]);
const selectedKernels = ref([]);
const selectedFlags = ref([]);
const customFlags = ref([]);
const fingerprintText = ref(JSON.stringify(defaultFingerprint, null, 2));
const fingerprintFileName = ref('');
const treeMode = ref(false);
const fileInput = ref(null);

const allSourcesChecked = computed(() => sources.value.length === SOURCE_IDS.length);
const sourcesIndeterminate = computed(
  () => sources.value.length > 0 && sources.value.length < SOURCE_IDS.length
);
const filteredKernels = computed(() =>
  kernels.value.filter((item) => sources.value.includes(item.source))
);
const treeNodes = computed(() => {
  try {
    return nodesOf(JSON.parse(fingerprintText.value), '');
  } catch {
    return [];
  }
});

watch(filteredKernels, (list) => {
  const allow = new Set(list.map((item) => item.path));
  const next = selectedKernels.value.filter((path) => allow.has(path));
  if (next.length !== selectedKernels.value.length) {
    selectedKernels.value = next;
  }
});

function onAllSourcesChange(checked) {
  sources.value = checked ? [...SOURCE_IDS] : [];
}

function addCustomFlag() {
  customFlags.value.push('');
}

function removeCustomFlag(index) {
  customFlags.value.splice(index, 1);
}

function pickFingerprint() {
  fileInput.value?.click();
}

async function onFingerprintFile(event) {
  const file = event.target.files?.[0];
  event.target.value = '';
  if (!file) return;
  try {
    const text = await file.text();
    const parsed = JSON.parse(text);
    fingerprintText.value = JSON.stringify(parsed, null, 2);
    fingerprintFileName.value = file.name;
    ElMessage.success(`已载入 ${file.name}`);
  } catch {
    ElMessage.error('所选文件不是合法 JSON');
  }
}

function guardTreeMode() {
  if (treeMode.value) return true;
  try {
    JSON.parse(fingerprintText.value);
    return true;
  } catch {
    ElMessage.error('JSON 格式不正确，无法切换到树形');
    return false;
  }
}

function nodesOf(value, parentPath) {
  if (value === null || typeof value !== 'object') return [];
  const entries = Array.isArray(value)
    ? value.map((item, index) => [String(index), item, true])
    : Object.entries(value).map(([key, child]) => [key, child, false]);
  return entries.map(([key, child, isIndex]) => {
    const path = !parentPath
      ? key
      : isIndex
        ? `${parentPath}[${key}]`
        : `${parentPath}.${key}`;
    const branch = child !== null && typeof child === 'object';
    const children = branch ? nodesOf(child, path) : undefined;
    return {
      id: path,
      label: key,
      summary: branch ? branchSummary(child) : formatPrimitive(child),
      children: children && children.length ? children : undefined
    };
  });
}

function branchSummary(value) {
  return Array.isArray(value) ? `Array(${value.length})` : `Object(${Object.keys(value).length})`;
}

function formatPrimitive(value) {
  if (typeof value === 'string') return JSON.stringify(value);
  return String(value);
}

async function loadKernels() {
  kernelLoading.value = true;
  try {
    kernels.value = await kernelsApi.listKernels();
  } catch (e) {
    ElMessage.error(e.message || '加载内核失败');
    kernels.value = [];
  } finally {
    kernelLoading.value = false;
  }
}

onMounted(loadKernels);
</script>

<style scoped>
.page-block {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 0;
}

.page-header {
  position: sticky;
  top: 0;
  z-index: 10;
  padding-bottom: 16px;
  margin-bottom: 0;
  background: var(--el-bg-color-page);
}

.page-card {
  border-radius: var(--el-border-radius-base);
}

.form-section {
  display: flex;
  flex-direction: column;
}

.form-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.form-section :deep(.el-form-item) {
  margin-bottom: 18px;
}

.form-section :deep(.el-form-item:last-child) {
  margin-bottom: 0;
}

.source-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.args-editor {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
}

.args-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.args-row .el-input {
  flex: 1;
}

.file-input {
  display: none;
}

.fingerprint-editor :deep(textarea) {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  line-height: 1.5;
}

.fingerprint-tree {
  min-height: 360px;
  max-height: 520px;
  overflow: auto;
  padding: 8px 4px;
  border: 1px solid var(--el-border-color);
  border-radius: var(--el-border-radius-base);
  background: var(--el-fill-color-blank);
}

.tree-label {
  display: inline-flex;
  gap: 8px;
  align-items: baseline;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
}

.tree-key {
  color: var(--el-color-primary);
}

.tree-summary {
  color: var(--el-text-color-secondary);
  word-break: break-all;
}
</style>
