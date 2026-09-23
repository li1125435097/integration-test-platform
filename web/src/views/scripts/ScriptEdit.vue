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
              <el-select v-model="form.language" placeholder="选择语言" @change="rebuildEditor">
                <el-option label="JavaScript" value="javascript" />
                <el-option label="Python" value="python" />
                <el-option label="Shell" value="shell" />
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
          <el-tag effect="plain" size="small">{{ languageLabel(form.language) }}</el-tag>
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
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessage } from 'element-plus';
import { ArrowLeft, Check, FolderAdd } from '@element-plus/icons-vue';
import { EditorView, basicSetup } from 'codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { python } from '@codemirror/lang-python';
import { StreamLanguage } from '@codemirror/language';
import { shell } from '@codemirror/legacy-modes/mode/shell';
import * as scriptsApi from '@/api/scripts';
import { languageLabel } from '@/utils/format';
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
const form = reactive({
  name: '',
  description: '',
  language: 'javascript',
  content: ''
});

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
    content: editorView ? editorView.state.doc.toString() : form.content
  };
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
    form.language = 'javascript';
    form.content = '';
    createEditor('', form.language);
    return;
  }
  loadScript(id);
}

onMounted(initFromRoute);

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
