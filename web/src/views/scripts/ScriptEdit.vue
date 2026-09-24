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
          <el-text tag="b">脚本内容</el-text>
          <el-space wrap>
            <el-tag effect="plain" size="small">{{ languageLabel(form.language) }}</el-tag>
            <el-button
              type="primary"
              size="small"
              :icon="VideoPlay"
              :loading="runLoading"
              @click="onRunPreview"
            >
              执行
            </el-button>
          </el-space>
        </el-row>
      </template>
      <div ref="editorHost" class="editor-host" />
    </el-card>

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
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { ArrowLeft, Check, FolderAdd, VideoPlay } from '@element-plus/icons-vue';
import ScriptRunResultDialog from '@/components/ScriptRunResultDialog.vue';
import { EditorView, basicSetup } from 'codemirror';
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

const route = useRoute();
const router = useRouter();

const editorHost = ref(null);
let editorView = null;

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

const form = reactive({
  name: '',
  description: '',
  language: 'javascript',
  interpreterId: '',
  content: ''
});

const languageOptions = computed(() =>
  languageOptionsFromInterpreters(interpreters.value, form.language)
);

const languageInterpreterOptions = computed(() =>
  interpretersForLanguage(interpreters.value, form.language)
);

function applyDefaultLanguage() {
  const opts = languageOptionsFromInterpreters(interpreters.value);
  if (opts.some((o) => o.value === form.language)) return;
  const preferred = opts.find((o) => o.value === 'javascript') || opts[0];
  form.language = preferred?.value || 'javascript';
}

function interpreterOptionLabel(interp) {
  const ver = interp.version?.trim();
  const base = ver || interp.path || interp.id;
  return interp.isDefault ? `${base}（默认）` : base;
}

function onLanguageChange() {
  const stillValid = languageInterpreterOptions.value.some((i) => i.id === form.interpreterId);
  if (!stillValid) {
    form.interpreterId = '';
  }
  rebuildEditor();
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

function createEditor(doc, lang) {
  if (!editorHost.value) return;
  if (editorView) {
    editorView.destroy();
    editorView = null;
  }
  editorView = new EditorView({
    doc: doc || '',
    extensions: [basicSetup, langExtension(lang)],
    parent: editorHost.value
  });
}

function rebuildEditor() {
  const doc = editorView ? editorView.state.doc.toString() : form.content;
  createEditor(doc, form.language);
}

function getPayload() {
  return {
    name: form.name.trim(),
    description: form.description.trim(),
    language: form.language,
    interpreterId: form.interpreterId || '',
    content: editorView ? editorView.state.doc.toString() : form.content
  };
}

async function onRunPreview() {
  const payload = getPayload();
  if (!payload.content.trim()) {
    ElMessage.warning('脚本内容为空');
    return;
  }
  runResultVisible.value = true;
  runLoading.value = true;
  runResult.value = null;
  try {
    runResult.value = await scriptsApi.runScriptPreview({
      language: payload.language,
      interpreterId: payload.interpreterId,
      content: payload.content
    });
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
    form.content = data.content || '';
    createEditor(form.content, form.language);
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
    form.content = '';
    applyDefaultLanguage();
    createEditor('', form.language);
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
