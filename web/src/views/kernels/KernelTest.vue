<template>
  <div class="page-block">
    <div class="page-header">
      <div class="page-header-row">
        <el-text tag="h2" size="large">内核测试</el-text>
        <div class="page-header-actions">
          <el-button type="primary" :loading="running" :disabled="fingerprintBusy" @click="onExecute">
            执行
          </el-button>
          <el-button :disabled="running || fingerprintBusy" @click="onClear">清空</el-button>
          <el-button
            :loading="savingPlan"
            :disabled="running || fingerprintBusy"
            @click="onSavePlan"
          >
            保存方案
          </el-button>
        </div>
      </div>
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
              <el-button :icon="FolderOpened" :disabled="fingerprintBusy" @click="pickFingerprint">打开</el-button>
              <el-text type="info" size="small">{{ fingerprintFileName || '内置示例' }}</el-text>
            </el-space>
            <el-space>
              <el-switch
                v-model="cipherMode"
                inline-prompt
                active-text="密文"
                inactive-text="明文"
                :disabled="saving"
                :before-change="guardCipherMode"
              />
              <el-switch
                v-model="treeMode"
                inline-prompt
                active-text="树形"
                inactive-text="编辑"
                :disabled="cipherMode || fingerprintBusy"
                :before-change="guardTreeMode"
              />
              <el-button type="primary" :loading="saving" :disabled="fingerprintBusy" @click="saveFingerprint">
                保存
              </el-button>
            </el-space>
          </div>
          <el-form-item label="baseUrl">
            <el-input v-model="baseUrl" clearable placeholder="指纹加密 baseUrl" />
          </el-form-item>
          <el-form-item label="Bearer">
            <el-input v-model="bearer" clearable placeholder="指纹加密 Bearer" />
          </el-form-item>
          <input ref="fileInput" class="file-input" type="file" @change="onFingerprintFile" />
          <el-form-item label="内容">
            <div class="fingerprint-body">
              <aside class="fingerprint-files">
                <el-text size="small" type="info">已保存</el-text>
                <el-scrollbar class="fingerprint-file-scroll">
                  <button
                    v-for="item in savedFiles"
                    :key="item.name"
                    type="button"
                    class="fingerprint-file"
                    :class="{ active: item.name === activeSavedName }"
                    :disabled="fingerprintBusy"
                    @click="openSaved(item.name)"
                  >
                    {{ item.name }}
                  </button>
                  <el-text v-if="!savedFiles.length" size="small" type="info">暂无文件</el-text>
                </el-scrollbar>
              </aside>
              <div v-loading="fingerprintBusy" class="fingerprint-main">
                <el-input
                  v-show="!treeMode"
                  v-model="fingerprintText"
                  type="textarea"
                  :rows="18"
                  resize="vertical"
                  class="fingerprint-editor"
                  :placeholder="editorPlaceholder"
                  @input="fingerprintDirty = true"
                />
                <div v-show="treeMode && !cipherMode" class="fingerprint-tree">
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
              </div>
            </div>
          </el-form-item>
        </section>

        <el-divider />

        <section class="form-section">
          <div class="form-section-header">
            <el-text type="info">脚本选择（脚本管理中名称以「内核测试-」开头的脚本）</el-text>
            <el-button :icon="Refresh" circle :loading="scriptLoading" @click="loadScripts" />
          </div>
          <el-form-item label="脚本">
            <el-select
              v-model="selectedScriptId"
              filterable
              clearable
              placeholder="选择脚本"
              no-data-text="没有名称以「内核测试-」开头的脚本"
              :loading="scriptLoading"
              style="width: 100%"
            >
              <el-option
                v-for="item in kernelScripts"
                :key="item.id"
                :label="item.name"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
        </section>
      </el-form>
    </el-card>

    <el-dialog v-model="concurrencyVisible" title="执行并发数" width="420px" align-center>
      <el-form label-width="72px" @submit.prevent>
        <el-form-item label="并发数">
          <el-input-number v-model="concurrency" :min="1" :max="selectedKernels.length" :step="1" />
        </el-form-item>
        <el-text type="info" size="small">已选 {{ selectedKernels.length }} 个内核，每个内核单独执行一次脚本。</el-text>
      </el-form>
      <template #footer>
        <el-button @click="concurrencyVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmConcurrency">执行</el-button>
      </template>
    </el-dialog>

    <KernelTestResultDialog v-model="resultVisible" :loading="resultLoading" :results="runResults" />
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Delete, FolderOpened, Plus, Refresh } from '@element-plus/icons-vue';
import * as kernelsApi from '@/api/kernels';
import * as scriptsApi from '@/api/scripts';
import * as fingerprintsApi from '@/api/fingerprints';
import KernelTestResultDialog from '@/components/KernelTestResultDialog.vue';
import {
  clearCachedForm,
  fingerprintPayload as fingerprintPayloadOf,
  launchArgs as launchArgsOf,
  readCachedForm,
  writeCachedForm
} from './formState';

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

const route = useRoute();
const router = useRouter();
const hydrating = ref(true);
const savingPlan = ref(false);
const kernelsReady = ref(false);
const scriptsReady = ref(false);

const kernelLoading = ref(false);
const kernels = ref([]);
const sources = ref([...SOURCE_IDS]);
const selectedKernels = ref([]);
const selectedFlags = ref([]);
const customFlags = ref([]);
const fingerprintText = ref(JSON.stringify(defaultFingerprint, null, 2));
const fingerprintFileName = ref('');
const treeMode = ref(false);
const cipherMode = ref(false);
const baseUrl = ref('');
const bearer = ref('');
const savedFiles = ref([]);
const activeSavedName = ref('');
const fingerprintBusy = ref(false);
const saving = ref(false);
const fileInput = ref(null);
const fingerprintDirty = ref(true);
const scriptLoading = ref(false);
const scripts = ref([]);
const selectedScriptId = ref('');
const running = ref(false);
const concurrencyVisible = ref(false);
const concurrency = ref(1);
const resultVisible = ref(false);
const resultLoading = ref(false);
const runResults = ref([]);

const allSourcesChecked = computed(() => sources.value.length === SOURCE_IDS.length);
const sourcesIndeterminate = computed(
  () => sources.value.length > 0 && sources.value.length < SOURCE_IDS.length
);
const filteredKernels = computed(() =>
  kernels.value.filter((item) => sources.value.includes(item.source))
);
const editorPlaceholder = computed(() => (cipherMode.value ? '指纹密文（Base64）' : '指纹 JSON'));
const kernelScripts = computed(() =>
  scripts.value.filter((item) => (item.name || '').startsWith('内核测试-'))
);
const treeNodes = computed(() => {
  if (cipherMode.value) return [];
  try {
    return nodesOf(JSON.parse(fingerprintText.value), '');
  } catch {
    return [];
  }
});

watch(filteredKernels, (list) => {
  if (hydrating.value) return;
  const allow = new Set(list.map((item) => item.path));
  const next = selectedKernels.value.filter((path) => allow.has(path));
  if (next.length !== selectedKernels.value.length) {
    selectedKernels.value = next;
  }
});

watch(kernelScripts, (list) => {
  if (hydrating.value) return;
  if (!list.some((item) => item.id === selectedScriptId.value)) {
    selectedScriptId.value = '';
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
  if (!file || fingerprintBusy.value) return;
  fingerprintBusy.value = true;
  try {
    const bytes = new Uint8Array(await file.arrayBuffer());
    const text = await fingerprintsApi.decryptFingerprint(bytesToBase64(bytes));
    fingerprintText.value = text;
    fingerprintFileName.value = file.name;
    cipherMode.value = false;
    treeMode.value = false;
    activeSavedName.value = '';
    fingerprintDirty.value = true;
    ElMessage.success(`已载入 ${file.name}`);
  } catch (e) {
    ElMessage.error(e.message || '解密失败');
  } finally {
    fingerprintBusy.value = false;
  }
}

function guardTreeMode() {
  if (cipherMode.value) return false;
  if (treeMode.value) return true;
  try {
    JSON.parse(fingerprintText.value);
    return true;
  } catch {
    ElMessage.error('JSON 格式不正确，无法切换到树形');
    return false;
  }
}

async function guardCipherMode() {
  if (fingerprintBusy.value) return false;
  fingerprintBusy.value = true;
  try {
    if (!cipherMode.value) {
      assertPlainObject(fingerprintText.value);
      const encoded = await fingerprintsApi.encryptFingerprint({
        plaintext: fingerprintText.value,
        baseUrl: baseUrl.value,
        bearer: bearer.value
      });
      fingerprintText.value = encoded;
      treeMode.value = false;
      return true;
    }
    fingerprintText.value = await fingerprintsApi.decryptFingerprint(fingerprintText.value);
    return true;
  } catch (e) {
    ElMessage.error(e.message || '转换失败');
    return false;
  } finally {
    fingerprintBusy.value = false;
  }
}

function assertPlainObject(text) {
  let parsed;
  try {
    parsed = JSON.parse(text);
  } catch {
    throw new Error('明文不是合法 JSON');
  }
  if (parsed === null || typeof parsed !== 'object') {
    throw new Error('指纹内容必须是 JSON 对象');
  }
}

function validateSaveName(value) {
  const name = value ?? '';
  if (
    !name.trim() ||
    name !== name.trim() ||
    name === '.' ||
    name === '..' ||
    name.includes('..') ||
    /[\\/]/.test(name)
  ) {
    return '文件名不能为空，且不能包含路径';
  }
  return true;
}

async function saveFingerprint() {
  if (fingerprintBusy.value) return;
  let name = '';
  try {
    const result = await ElMessageBox.prompt('请输入保存文件名称', '保存指纹', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValue: activeSavedName.value || '',
      inputValidator: validateSaveName
    });
    name = result.value;
  } catch {
    return;
  }
  if (savedFiles.value.some((item) => item.name === name)) {
    try {
      await ElMessageBox.confirm(`「${name}」已存在，是否覆盖？`, '覆盖确认', {
        confirmButtonText: '覆盖',
        cancelButtonText: '取消',
        type: 'warning'
      });
    } catch {
      return;
    }
  }
  fingerprintBusy.value = true;
  saving.value = true;
  try {
    let encoded = fingerprintText.value.trim();
    if (!cipherMode.value) {
      assertPlainObject(fingerprintText.value);
      encoded = await fingerprintsApi.encryptFingerprint({
        plaintext: fingerprintText.value,
        baseUrl: baseUrl.value,
        bearer: bearer.value
      });
    }
    await fingerprintsApi.saveFingerprint(name, encoded);
    await loadSavedFiles();
    activeSavedName.value = name;
    fingerprintFileName.value = name;
    fingerprintDirty.value = false;
    ElMessage.success(`已保存 ${name}`);
  } catch (e) {
    ElMessage.error(e.message || '保存失败');
  } finally {
    fingerprintBusy.value = false;
    saving.value = false;
  }
}

async function openSaved(name) {
  if (fingerprintBusy.value) return;
  fingerprintBusy.value = true;
  try {
    const encoded = await fingerprintsApi.readFingerprint(name);
    fingerprintText.value = await fingerprintsApi.decryptFingerprint(encoded);
    cipherMode.value = false;
    treeMode.value = false;
    fingerprintFileName.value = name;
    activeSavedName.value = name;
    fingerprintDirty.value = false;
  } catch (e) {
    ElMessage.error(e.message || '读取指纹失败');
  } finally {
    fingerprintBusy.value = false;
  }
}

async function loadSavedFiles() {
  savedFiles.value = await fingerprintsApi.listFingerprints();
}

async function loadEncryptDefaults() {
  try {
    const items = await scriptsApi.listScripts();
    const script = items.find((item) => item.name === '指纹加密');
    const vars = script?.variables || [];
    baseUrl.value = vars.find((item) => item.name === 'baseUrl')?.value ?? '';
    bearer.value = vars.find((item) => item.name === 'Bearer')?.value ?? '';
  } catch (e) {
    ElMessage.error(e.message || '加载加密变量失败');
  }
}

function bytesToBase64(bytes) {
  let binary = '';
  const chunk = 0x8000;
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunk));
  }
  return btoa(binary);
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
    kernelsReady.value = true;
  } catch (e) {
    ElMessage.error(e.message || '加载内核失败');
    if (!kernelsReady.value) kernels.value = [];
  } finally {
    kernelLoading.value = false;
  }
}

async function loadScripts() {
  scriptLoading.value = true;
  try {
    scripts.value = await scriptsApi.listScripts();
    scriptsReady.value = true;
  } catch (e) {
    ElMessage.error(e.message || '加载脚本失败');
    if (!scriptsReady.value) scripts.value = [];
  } finally {
    scriptLoading.value = false;
  }
}

function formSnapshot() {
  return {
    sources: [...sources.value],
    kernels: [...selectedKernels.value],
    flags: [...selectedFlags.value],
    customFlags: [...customFlags.value],
    scriptId: selectedScriptId.value,
    fingerprintText: fingerprintText.value,
    fingerprintFileName: fingerprintFileName.value,
    cipherMode: cipherMode.value,
    treeMode: treeMode.value,
    baseUrl: baseUrl.value,
    bearer: bearer.value,
    activeSavedName: activeSavedName.value,
    fingerprintDirty: fingerprintDirty.value
  };
}

function applyForm(form) {
  const nextSources = Array.isArray(form?.sources)
    ? form.sources.filter((id) => SOURCE_IDS.includes(id))
    : [...SOURCE_IDS];
  sources.value = nextSources;
  const allow = new Set(
    kernels.value.filter((item) => nextSources.includes(item.source)).map((item) => item.path)
  );
  const savedKernels = Array.isArray(form?.kernels) ? [...form.kernels] : [];
  selectedKernels.value = kernelsReady.value
    ? savedKernels.filter((path) => allow.has(path))
    : savedKernels;
  selectedFlags.value = Array.isArray(form?.flags) ? [...form.flags] : [];
  customFlags.value = Array.isArray(form?.customFlags)
    ? form.customFlags.map((item) => String(item ?? ''))
    : [];
  fingerprintText.value =
    typeof form?.fingerprintText === 'string'
      ? form.fingerprintText
      : JSON.stringify(defaultFingerprint, null, 2);
  fingerprintFileName.value = form?.fingerprintFileName || '';
  cipherMode.value = Boolean(form?.cipherMode);
  treeMode.value = Boolean(form?.treeMode) && !cipherMode.value;
  baseUrl.value = form?.baseUrl ?? '';
  bearer.value = form?.bearer ?? '';
  activeSavedName.value = form?.activeSavedName || '';
  fingerprintDirty.value = form?.fingerprintDirty !== false;
  const scriptId = form?.scriptId || '';
  selectedScriptId.value =
    !scriptsReady.value || kernelScripts.value.some((item) => item.id === scriptId) ? scriptId : '';
}

watch(formSnapshot, (snap) => {
  if (hydrating.value) return;
  writeCachedForm(snap);
});

function resetFormFields() {
  sources.value = [...SOURCE_IDS];
  selectedKernels.value = [];
  selectedFlags.value = [];
  customFlags.value = [];
  fingerprintText.value = JSON.stringify(defaultFingerprint, null, 2);
  fingerprintFileName.value = '';
  treeMode.value = false;
  cipherMode.value = false;
  activeSavedName.value = '';
  fingerprintDirty.value = true;
  selectedScriptId.value = '';
}

function launchArgs() {
  return launchArgsOf(formSnapshot());
}

function fingerprintPayload() {
  return fingerprintPayloadOf(formSnapshot());
}

function onExecute() {
  if (running.value) return;
  if (!selectedKernels.value.length) {
    ElMessage.warning('请选择内核');
    return;
  }
  if (!selectedScriptId.value) {
    ElMessage.warning('请选择脚本');
    return;
  }
  if (!fingerprintText.value.trim()) {
    ElMessage.warning('指纹内容不能为空');
    return;
  }
  if (selectedKernels.value.length > 1) {
    concurrency.value = 1;
    concurrencyVisible.value = true;
    return;
  }
  execute(1);
}

function confirmConcurrency() {
  const count = Number(concurrency.value);
  const max = selectedKernels.value.length;
  if (!Number.isInteger(count) || count < 1 || count > max) {
    ElMessage.warning('并发数无效');
    return;
  }
  concurrencyVisible.value = false;
  execute(count);
}

async function execute(count) {
  running.value = true;
  resultVisible.value = true;
  resultLoading.value = true;
  runResults.value = [];
  try {
    runResults.value = await kernelsApi.runKernelTest({
      scriptId: selectedScriptId.value,
      concurrency: count,
      kernels: [...selectedKernels.value],
      args: launchArgs(),
      fingerprint: fingerprintPayload()
    });
  } catch (e) {
    resultVisible.value = false;
    ElMessage.error(e.message || '执行失败');
  } finally {
    running.value = false;
    resultLoading.value = false;
  }
}

async function onClear() {
  try {
    await ElMessageBox.confirm('确定清空当前表单？', '清空', {
      confirmButtonText: '清空',
      cancelButtonText: '取消',
      type: 'warning'
    });
  } catch {
    return;
  }
  hydrating.value = true;
  resetFormFields();
  await loadEncryptDefaults();
  clearCachedForm();
  hydrating.value = false;
}

function validatePlanName(value) {
  const name = value ?? '';
  if (!name.trim()) return '方案名称不能为空';
  return true;
}

async function onSavePlan() {
  if (savingPlan.value) return;
  let name = '';
  try {
    const result = await ElMessageBox.prompt('请输入方案名称', '保存方案', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValidator: validatePlanName
    });
    name = result.value.trim();
  } catch {
    return;
  }
  const form = formSnapshot();
  let existing = null;
  try {
    const plans = await kernelsApi.listKernelTestPlans();
    existing = plans.find((item) => item.name === name) || null;
  } catch (e) {
    ElMessage.error(e.message || '保存方案失败');
    return;
  }
  if (existing) {
    try {
      await ElMessageBox.confirm(`「${name}」已存在，是否覆盖？`, '覆盖确认', {
        confirmButtonText: '覆盖',
        cancelButtonText: '取消',
        type: 'warning'
      });
    } catch {
      return;
    }
  }
  savingPlan.value = true;
  try {
    if (existing) {
      await kernelsApi.updateKernelTestPlan(existing.id, { name, form });
    } else {
      await kernelsApi.createKernelTestPlan({ name, form });
    }
    ElMessage.success(`已保存 ${name}`);
  } catch (e) {
    ElMessage.error(e.message || '保存方案失败');
  } finally {
    savingPlan.value = false;
  }
}

async function restoreForm() {
  const planId = typeof route.query.planId === 'string' ? route.query.planId : '';
  if (planId) {
    try {
      const plan = await kernelsApi.getKernelTestPlan(planId);
      applyForm(plan.form);
    } catch (e) {
      ElMessage.error(e.message || '加载方案失败');
      const cached = readCachedForm();
      if (cached) applyForm(cached);
    }
    await router.replace({ name: 'kernel-test' });
    return;
  }
  const cached = readCachedForm();
  if (cached) applyForm(cached);
}

onMounted(async () => {
  hydrating.value = true;
  try {
    await Promise.all([
      loadKernels(),
      loadScripts(),
      loadEncryptDefaults(),
      loadSavedFiles().catch((e) => {
        ElMessage.error(e.message || '加载指纹文件失败');
      })
    ]);
    await restoreForm();
  } finally {
    hydrating.value = false;
    writeCachedForm(formSnapshot());
  }
});

watch(
  () => (typeof route.query.planId === 'string' ? route.query.planId : ''),
  async (planId) => {
    if (!planId || hydrating.value) return;
    hydrating.value = true;
    try {
      await restoreForm();
    } finally {
      hydrating.value = false;
      writeCachedForm(formSnapshot());
    }
  }
);
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

.page-header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.page-header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
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

.fingerprint-body {
  display: flex;
  gap: 12px;
  width: 100%;
  min-height: 360px;
}

.fingerprint-files {
  flex: 0 0 180px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 360px;
  max-height: 520px;
  padding: 8px;
  border: 1px solid var(--el-border-color);
  border-radius: var(--el-border-radius-base);
  background: var(--el-fill-color-blank);
}

.fingerprint-file-scroll {
  flex: 1;
  min-height: 0;
}

.fingerprint-file {
  display: block;
  width: 100%;
  margin: 0 0 4px;
  padding: 6px 8px;
  border: 0;
  border-radius: var(--el-border-radius-base);
  background: transparent;
  color: var(--el-text-color-regular);
  font: inherit;
  text-align: left;
  cursor: pointer;
  word-break: break-all;
}

.fingerprint-file:hover,
.fingerprint-file.active {
  background: var(--el-fill-color-light);
  color: var(--el-color-primary);
}

.fingerprint-file:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.fingerprint-main {
  flex: 1;
  min-width: 0;
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
