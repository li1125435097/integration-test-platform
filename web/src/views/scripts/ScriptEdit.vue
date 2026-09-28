<template>
  <div class="page-block">
    <el-page-header :icon="ArrowLeft" @back="router.push('/scripts')">
      <template #content>
        <el-text tag="span" size="large">{{ pageTitle }}</el-text>
      </template>
      <template #extra>
        <el-space wrap>
          <el-button type="primary" :icon="Check" :loading="saving" @click="save">保存</el-button>
          <el-button :icon="FolderAdd" :disabled="!scriptId" @click="onAddVersion">添加版本</el-button>
        </el-space>
      </template>
    </el-page-header>

    <el-card shadow="never">
      <template #header>
        <el-text tag="b">基本信息</el-text>
      </template>
      <el-form label-width="96px" label-position="right" @submit.prevent>
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :lg="8">
            <el-form-item label="脚本名称" required>
              <el-input v-model="form.name" placeholder="请输入脚本名称" clearable />
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :lg="8">
            <el-form-item label="脚本语言">
              <el-select
                v-model="form.language"
                placeholder="请先在解释器管理配置语言"
                :disabled="!languageOptions.length"
                style="width: 100%"
                @change="onLanguageChange"
              >
                <el-option
                  v-for="opt in languageOptions"
                  :key="opt.value"
                  :label="opt.label"
                  :value="opt.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :lg="8">
            <el-form-item label="解释器">
              <el-select
                v-model="form.interpreterId"
                placeholder="默认"
                clearable
                style="width: 100%"
              >
                <el-option label="默认（按语言使用解释器管理中的默认项）" value="" />
                <el-option
                  v-for="opt in languageInterpreterOptions"
                  :key="opt.id"
                  :label="interpreterOptionLabel(opt)"
                  :value="opt.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="脚本描述">
              <el-input
                v-model="form.description"
                type="textarea"
                :rows="2"
                placeholder="可选，简要说明脚本用途"
                maxlength="500"
                show-word-limit
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
    </el-card>

    <el-card shadow="never" class="editor-panel">
      <template #header>
        <el-row justify="space-between" align="middle">
          <el-space wrap alignment="center">
            <el-text tag="b">脚本内容</el-text>
            <el-tag effect="plain" size="small">{{ languageLabel(form.language) }}</el-tag>
            <el-text type="info" size="small">{{ requireHint(form.language) }}</el-text>
          </el-space>
          <el-button
            type="primary"
            size="small"
            :icon="VideoPlay"
            :loading="runLoading"
            @click="onRunPreview"
          >
            执行
          </el-button>
        </el-row>
      </template>

      <el-tabs
        :model-value="activeFile"
        type="card"
        class="file-tabs"
        addable
        @tab-change="selectFile"
        @tab-add="openAddDialog"
        @tab-remove="onTabRemove"
      >
        <el-tab-pane
          v-for="f in files"
          :key="f.name"
          :name="f.name"
          :closable="f.kind !== 'main'"
        >
          <template #label>
            <span class="tab-label">
              {{ f.name }}
              <el-icon v-if="f.kind === 'ref'" class="tab-icon"><Lock /></el-icon>
              <el-icon v-if="f.missing" class="tab-icon is-missing"><Warning /></el-icon>
            </span>
          </template>
        </el-tab-pane>
      </el-tabs>

      <div v-if="currentFile?.kind === 'ref'" class="ref-banner">
        <el-text type="info" size="small">
          引用自 {{ currentFile.sourceScriptName || currentFile.sourceScriptId }} /
          {{ currentFile.sourceFileName }}（只读）
        </el-text>
        <el-text v-if="currentFile.missing" type="danger" size="small">源文件缺失</el-text>
      </div>

      <div ref="editorHost" class="editor-host" />
    </el-card>

    <el-dialog
      v-model="addVisible"
      title="添加文件"
      width="520px"
      destroy-on-close
      align-center
      @closed="resetAddForm"
    >
      <el-form label-width="108px" @submit.prevent>
        <el-form-item label="添加方式">
          <el-radio-group v-model="addMode">
            <el-radio label="blank">空白文件</el-radio>
            <el-radio label="ref">引用其他脚本</el-radio>
          </el-radio-group>
        </el-form-item>
        <template v-if="addMode === 'ref'">
          <el-form-item label="来源脚本" required>
            <el-select
              v-model="refScriptId"
              placeholder="选择同语言脚本"
              filterable
              style="width: 100%"
              @change="onRefScriptChange"
            >
              <el-option
                v-for="sc in refScriptOptions"
                :key="sc.id"
                :label="sc.name"
                :value="sc.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="来源文件" required>
            <el-select
              v-model="refFileName"
              placeholder="选择要引用的文件"
              style="width: 100%"
              :disabled="!refFileOptions.length"
              @change="onRefFileChange"
            >
              <el-option
                v-for="rf in refFileOptions"
                :key="rf.name"
                :label="rf.kind === 'main' ? `${rf.name}（main）` : rf.name"
                :value="rf.name"
              />
            </el-select>
          </el-form-item>
        </template>
        <el-form-item label="本脚本文件名" required>
          <el-input
            v-model="addName"
            :placeholder="`例如 helper，将保存为 helper.${extForLanguage(form.language)}`"
            clearable
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addVisible = false">取消</el-button>
        <el-button type="primary" :loading="addingFile" @click="confirmAddFile">确认</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="addVersionVisible"
      title="添加版本"
      width="480px"
      destroy-on-close
      align-center
      @closed="resetAddVersionForm"
    >
      <el-form label-width="80px" @submit.prevent>
        <el-form-item label="版本 ID">
          <el-text>{{ pendingVersionId || '—' }}</el-text>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="addVersionRemark"
            type="textarea"
            :rows="3"
            placeholder="可选，例如本次变更说明"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addVersionVisible = false">取消</el-button>
        <el-button type="primary" :loading="addingVersion" @click="submitAddVersion">确认</el-button>
      </template>
    </el-dialog>

    <ScriptRunResultDialog
      v-model="runResultVisible"
      title="试执行结果（未写入执行记录）"
      :loading="runLoading"
      :result="runResult"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { ArrowLeft, Check, FolderAdd, VideoPlay, Lock, Warning } from '@element-plus/icons-vue';
import ScriptRunResultDialog from '@/components/ScriptRunResultDialog.vue';
import { EditorView, basicSetup } from 'codemirror';
import { EditorState, Compartment } from '@codemirror/state';
import { javascript } from '@codemirror/lang-javascript';
import { python } from '@codemirror/lang-python';
import { StreamLanguage } from '@codemirror/language';
import { shell } from '@codemirror/legacy-modes/mode/shell';
import * as scriptsApi from '@/api/scripts';
import * as interpretersApi from '@/api/interpreters';
import { languageLabel } from '@/utils/format';
import {
  interpretersForLanguage,
  languageOptionsFromInterpreters
} from '@/utils/interpreters';
import { nextVersionId } from '@/utils/versions';
import {
  extForLanguage,
  mainFileName,
  withLanguageExt,
  isValidFileName,
  emptyMainFile,
  filesFromDetail,
  renameOwnedFilesForLanguage,
  requireHint
} from '@/utils/scriptFiles';

const route = useRoute();
const router = useRouter();

const editorHost = ref(null);
let editorView = null;
const langCompartment = new Compartment();
const readOnlyCompartment = new Compartment();

const scriptId = ref(null);
const saving = ref(false);
const addVersionVisible = ref(false);
const addingVersion = ref(false);
const pendingVersionId = ref('');
const addVersionRemark = ref('');
const interpreters = ref([]);
const runLoading = ref(false);
const runResultVisible = ref(false);
const runResult = ref(null);
const files = ref([emptyMainFile('javascript')]);
const activeFile = ref(mainFileName('javascript'));
const languageBeforeChange = ref('javascript');

const addVisible = ref(false);
const addMode = ref('blank');
const addName = ref('');
const addingFile = ref(false);
const otherScripts = ref([]);
const refScriptId = ref('');
const refFileName = ref('');

const form = reactive({
  name: '',
  description: '',
  language: 'javascript',
  interpreterId: ''
});

const languageOptions = computed(() =>
  languageOptionsFromInterpreters(interpreters.value, form.language)
);

const languageInterpreterOptions = computed(() =>
  interpretersForLanguage(interpreters.value, form.language)
);

const currentFile = computed(() => files.value.find((f) => f.name === activeFile.value) || null);

const refScriptOptions = computed(() =>
  otherScripts.value.filter((sc) => sc.language === form.language && sc.id !== scriptId.value)
);

const refFileOptions = computed(() => {
  const sc = refScriptOptions.value.find((s) => s.id === refScriptId.value);
  if (!sc) return [];
  if (Array.isArray(sc.files) && sc.files.length) return sc.files;
  return [{ name: mainFileName(sc.language), kind: 'main' }];
});

function applyDefaultLanguage() {
  const opts = languageOptionsFromInterpreters(interpreters.value);
  if (opts.some((o) => o.value === form.language)) return;
  const preferred = opts.find((o) => o.value === 'javascript') || opts[0];
  form.language = preferred?.value || 'javascript';
  languageBeforeChange.value = form.language;
}

function interpreterOptionLabel(interp) {
  const ver = interp.version?.trim();
  const base = ver || interp.path || interp.id;
  return interp.isDefault ? `${base}（默认）` : base;
}

function onLanguageChange(next) {
  if (files.value.some((f) => f.kind === 'ref')) {
    ElMessage.warning('请先移除引用的子脚本再切换语言');
    form.language = languageBeforeChange.value;
    return;
  }
  const stillValid = languageInterpreterOptions.value.some((i) => i.id === form.interpreterId);
  if (!stillValid) {
    form.interpreterId = '';
  }
  flushEditorToFile();
  const prevActive = activeFile.value;
  files.value = renameOwnedFilesForLanguage(files.value, next);
  const renamed = files.value.find((f) => f.kind === 'main') || files.value[0];
  const still = files.value.find((f) => f.name === withLanguageExt(prevActive, next));
  activeFile.value = still?.name || renamed?.name;
  languageBeforeChange.value = next;
  applyEditor(currentFile.value);
}

const pageTitle = computed(() =>
  route.params.id === 'new' || !scriptId.value ? '新建脚本' : `编辑 · ${form.name || '未命名'}`
);

function langExtension(lang) {
  switch (lang) {
    case 'python':
      return python();
    case 'shell':
      return StreamLanguage.define(shell);
    default:
      return javascript();
  }
}

function isReadOnlyFile(file) {
  return !!file && (file.kind === 'ref' || file.missing);
}

function createEditor(doc, lang, readOnly) {
  if (!editorHost.value) return;
  if (editorView) {
    editorView.destroy();
    editorView = null;
  }
  editorView = new EditorView({
    doc: doc || '',
    extensions: [
      basicSetup,
      langCompartment.of(langExtension(lang)),
      readOnlyCompartment.of(EditorState.readOnly.of(!!readOnly))
    ],
    parent: editorHost.value
  });
}

function applyEditor(file) {
  if (!file) return;
  const readOnly = isReadOnlyFile(file);
  if (!editorView) {
    createEditor(file.content, form.language, readOnly);
    return;
  }
  editorView.dispatch({
    changes: { from: 0, to: editorView.state.doc.length, insert: file.content || '' },
    effects: [
      langCompartment.reconfigure(langExtension(form.language)),
      readOnlyCompartment.reconfigure(EditorState.readOnly.of(readOnly))
    ]
  });
}

function flushEditorToFile() {
  const cur = files.value.find((f) => f.name === activeFile.value);
  if (!cur || isReadOnlyFile(cur) || !editorView) return;
  cur.content = editorView.state.doc.toString();
}

function selectFile(name) {
  if (!name || name === activeFile.value) return;
  flushEditorToFile();
  activeFile.value = name;
  applyEditor(currentFile.value);
}

async function onTabRemove(name) {
  const target = files.value.find((f) => f.name === name);
  if (!target || target.kind === 'main') return;
  try {
    await ElMessageBox.confirm(
      `确定关闭并移除文件「${name}」？保存脚本后才会从磁盘删除。`,
      '关闭文件',
      { type: 'warning', confirmButtonText: '确认', cancelButtonText: '取消' }
    );
  } catch {
    return;
  }
  const removingActive = activeFile.value === name;
  files.value = files.value.filter((f) => f.name !== name);
  if (removingActive) {
    activeFile.value = mainFileName(form.language);
    applyEditor(currentFile.value);
  }
}

async function openAddDialog() {
  addMode.value = 'blank';
  addName.value = '';
  refScriptId.value = '';
  refFileName.value = '';
  addVisible.value = true;
  try {
    otherScripts.value = await scriptsApi.listScripts();
  } catch {
    otherScripts.value = [];
  }
}

function resetAddForm() {
  addMode.value = 'blank';
  addName.value = '';
  addingFile.value = false;
  refScriptId.value = '';
  refFileName.value = '';
}

function suggestLocalName(fileName) {
  const ext = extForLanguage(form.language);
  let name = withLanguageExt(fileName, form.language);
  const reserved = mainFileName(form.language);
  const taken = (n) => n === reserved || files.value.some((f) => f.name === n);
  if (!taken(name)) return name;
  const base = String(fileName || 'file').replace(/\.[^.]+$/, '') || 'file';
  let candidate = `ref_${base}.${ext}`;
  let n = 1;
  while (taken(candidate)) {
    candidate = `ref_${base}${n}.${ext}`;
    n += 1;
  }
  return candidate;
}

function onRefScriptChange() {
  const first = refFileOptions.value[0];
  refFileName.value = first?.name || '';
  if (refFileName.value) {
    addName.value = suggestLocalName(refFileName.value);
  }
}

function onRefFileChange() {
  if (refFileName.value) {
    addName.value = suggestLocalName(refFileName.value);
  }
}

async function confirmAddFile() {
  const name = withLanguageExt(addName.value, form.language);
  if (!name || !isValidFileName(name)) {
    ElMessage.warning('文件名仅支持字母、数字、下划线、点、短横线');
    return;
  }
  if (name === mainFileName(form.language)) {
    ElMessage.warning(`${name} 为入口文件，请换一个名字`);
    return;
  }
  if (files.value.some((f) => f.name === name)) {
    ElMessage.warning('文件名已存在');
    return;
  }
  if (addMode.value === 'blank') {
    flushEditorToFile();
    files.value.push({
      name,
      kind: 'local',
      content: '',
      sourceScriptId: '',
      sourceFileName: '',
      sourceScriptName: '',
      missing: false
    });
    addVisible.value = false;
    await nextTick();
    selectFile(name);
    return;
  }
  if (!refScriptId.value || !refFileName.value) {
    ElMessage.warning('请选择要引用的脚本和文件');
    return;
  }
  addingFile.value = true;
  try {
    const data = await scriptsApi.getScript(refScriptId.value);
    const srcFile = (data.files || []).find((f) => f.name === refFileName.value);
    flushEditorToFile();
    files.value.push({
      name,
      kind: 'ref',
      content: srcFile?.content || '',
      sourceScriptId: refScriptId.value,
      sourceFileName: refFileName.value,
      sourceScriptName: data.name || '',
      missing: !srcFile || !!srcFile.missing
    });
    addVisible.value = false;
    await nextTick();
    selectFile(name);
  } catch (e) {
    ElMessage.error(e.message || '加载引用文件失败');
  } finally {
    addingFile.value = false;
  }
}

function getPayload() {
  flushEditorToFile();
  const main = files.value.find((f) => f.kind === 'main') || files.value[0];
  return {
    name: form.name.trim(),
    description: form.description.trim(),
    language: form.language,
    interpreterId: form.interpreterId || '',
    content: main?.content || '',
    files: files.value.map((f) => ({
      name: f.name,
      kind: f.kind,
      content: f.kind === 'ref' ? '' : f.content || '',
      sourceScriptId: f.sourceScriptId || '',
      sourceFileName: f.sourceFileName || ''
    }))
  };
}

async function onRunPreview() {
  const payload = getPayload();
  if (!(payload.content || '').trim()) {
    ElMessage.warning('main 脚本内容为空');
    return;
  }
  runResultVisible.value = true;
  runLoading.value = true;
  runResult.value = null;
  try {
    runResult.value = await scriptsApi.runScriptPreview(payload);
  } catch (e) {
    runResult.value = {
      success: false,
      exitCode: -1,
      durationMs: 0,
      stdout: '',
      stderr: '',
      error: e.message || '执行失败'
    };
  } finally {
    runLoading.value = false;
  }
}

async function save() {
  const payload = getPayload();
  if (!payload.name) {
    ElMessage.warning('请填写脚本名称');
    return;
  }
  saving.value = true;
  try {
    if (scriptId.value) {
      await scriptsApi.updateScript(scriptId.value, payload);
    } else {
      await scriptsApi.createScript(payload);
    }
    ElMessage.success('保存成功');
    await router.push({ name: 'scripts' });
  } catch (e) {
    ElMessage.error(e.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

function resetAddVersionForm() {
  pendingVersionId.value = '';
  addVersionRemark.value = '';
}

async function onAddVersion() {
  if (!scriptId.value) {
    ElMessage.warning('请先保存脚本后再添加版本');
    return;
  }
  try {
    const versions = await scriptsApi.listVersions(scriptId.value);
    pendingVersionId.value = nextVersionId(versions);
    addVersionRemark.value = '';
    addVersionVisible.value = true;
  } catch (e) {
    ElMessage.error(e.message || '加载版本列表失败');
  }
}

async function submitAddVersion() {
  if (!scriptId.value || !pendingVersionId.value) return;
  addingVersion.value = true;
  try {
    const ver = await scriptsApi.addVersion(
      scriptId.value,
      pendingVersionId.value,
      addVersionRemark.value.trim()
    );
    ElMessage.success(`已添加版本：${ver.id}`);
    addVersionVisible.value = false;
  } catch (e) {
    ElMessage.error(e.message || '添加版本失败');
  } finally {
    addingVersion.value = false;
  }
}

async function loadScript(id) {
  try {
    const data = await scriptsApi.getScript(id);
    scriptId.value = data.id;
    form.name = data.name || '';
    form.description = data.description || '';
    form.language = data.language || 'javascript';
    form.interpreterId = data.interpreterId || '';
    languageBeforeChange.value = form.language;
    files.value = filesFromDetail(data);
    activeFile.value = (files.value.find((f) => f.kind === 'main') || files.value[0]).name;
    await nextTick();
    createEditor(currentFile.value?.content || '', form.language, isReadOnlyFile(currentFile.value));
  } catch {
    ElMessage.error('加载脚本失败');
    router.push('/scripts');
  }
}

function initFromRoute() {
  const id = route.params.id;
  if (id === 'new') {
    scriptId.value = null;
    form.name = '';
    form.description = '';
    form.interpreterId = '';
    applyDefaultLanguage();
    files.value = [emptyMainFile(form.language)];
    activeFile.value = mainFileName(form.language);
    languageBeforeChange.value = form.language;
    nextTick(() => {
      createEditor('', form.language, false);
    });
    return;
  }
  loadScript(id);
}

async function loadInterpreters() {
  try {
    interpreters.value = await interpretersApi.listInterpreters();
  } catch {
    interpreters.value = [];
  }
}

onMounted(async () => {
  await loadInterpreters();
  initFromRoute();
});

watch(
  () => route.params.id,
  (id, prev) => {
    if (id !== prev) initFromRoute();
  }
);

onBeforeUnmount(() => {
  if (editorView) {
    editorView.destroy();
    editorView = null;
  }
});
</script>

<style scoped>
.page-block {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.editor-panel :deep(.el-card__body) {
  padding: 0;
}

.file-tabs {
  padding: 8px 8px 0;
}

.file-tabs :deep(.el-tabs__header) {
  margin: 0;
}

.file-tabs :deep(.el-tabs__content) {
  display: none;
}

.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.tab-icon {
  font-size: 12px;
}

.tab-icon.is-missing {
  color: var(--el-color-danger);
}

.ref-banner {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 6px 12px;
  border-top: 1px solid var(--el-border-color-lighter);
  background: var(--el-fill-color-light);
}

.editor-host {
  min-height: 440px;
  border-top: 1px solid var(--el-border-color-lighter);
}

.editor-host :deep(.cm-editor) {
  min-height: 440px;
  font-size: var(--el-font-size-base);
  background-color: var(--el-fill-color-blank);
}

.editor-host :deep(.cm-scroller) {
  font-family: var(--el-font-family);
}
</style>
