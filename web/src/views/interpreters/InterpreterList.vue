<template>
  <div class="page-block">
    <el-row justify="space-between" align="middle">
      <el-text tag="h2" size="large">解释器管理</el-text>
      <el-space>
        <el-button :icon="Download" @click="openDiscover()">自动获取</el-button>
        <el-button type="primary" :icon="Plus" @click="openCreate">新增</el-button>
      </el-space>
    </el-row>

    <el-card shadow="hover" class="page-card">
      <template #header>
        <el-row justify="space-between" align="middle">
          <el-text type="info">配置各语言解释器路径与默认启动参数；自动获取会扫描 PATH 等环境变量目录</el-text>
          <el-button :icon="Refresh" circle @click="loadList" :loading="loading" />
        </el-row>
      </template>

      <el-table v-loading="loading" :data="interpreters" border stripe style="width: 100%">
        <el-table-column type="index" label="#" width="56" align="center" />
        <el-table-column label="语言" width="132" align="center">
          <template #default="{ row }">
            <el-tag effect="plain" round size="small">{{ languageLabel(row.language) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="path" label="解释器路径" min-width="220" show-overflow-tooltip />
        <el-table-column label="默认参数" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ formatDefaultArgs(row.defaultArgs) }}</template>
        </el-table-column>
        <el-table-column label="版本" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.version || '—' }}</template>
        </el-table-column>
        <el-table-column label="默认" width="88" align="center">
          <template #default="{ row }">
            <el-switch
              :model-value="!!row.isDefault"
              :loading="defaultTogglingId === row.id"
              @change="(val) => onDefaultChange(row, val)"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="EditPen" @click="openEdit(row)">编辑</el-button>
            <el-divider direction="vertical" />
            <el-button type="primary" link :icon="Download" @click="openDiscover(row)">
              自动获取路径
            </el-button>
            <el-divider direction="vertical" />
            <el-button type="danger" link :icon="Delete" @click="confirmDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无解释器">
            <el-button type="primary" :icon="Plus" @click="openCreate">新增解释器</el-button>
          </el-empty>
        </template>
      </el-table>
    </el-card>

    <el-dialog
      v-model="formVisible"
      :title="formTitle"
      width="560px"
      destroy-on-close
      align-center
      @closed="resetForm"
    >
      <el-form label-width="96px" @submit.prevent>
        <el-form-item label="语言" required>
          <el-select v-model="form.language" placeholder="选择语言" style="width: 100%">
            <el-option
              v-for="opt in languageOptions"
              :key="opt.value"
              :label="opt.label"
              :value="opt.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="解释器路径" required>
          <el-input v-model="form.path" placeholder="可执行文件绝对路径" clearable />
        </el-form-item>
        <el-form-item label="默认参数">
          <div class="args-editor">
            <el-empty v-if="!form.defaultArgs.length" description="暂无参数" :image-size="48" />
            <div v-for="(_, index) in form.defaultArgs" :key="index" class="args-row">
              <el-input
                v-model="form.defaultArgs[index]"
                placeholder="例如 -u"
                clearable
              />
              <el-button type="danger" link :icon="Delete" @click="removeArg(index)">删除</el-button>
            </div>
            <el-button type="primary" link :icon="Plus" @click="addArg">添加参数</el-button>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="discoverVisible"
      :title="discoverTitle"
      width="760px"
      destroy-on-close
      align-center
      @closed="resetDiscover"
    >
      <div v-loading="discoverLoading">
        <el-empty v-if="!discoverLoading && !candidates.length" description="未检测到可用解释器" />
        <el-table
          v-else
          ref="discoverTableRef"
          :data="candidates"
          max-height="360"
          @selection-change="onDiscoverSelectionChange"
        >
          <el-table-column type="selection" width="48" align="center" />
          <el-table-column label="语言" width="120" align="center">
            <template #default="{ row }">
              <el-tag effect="plain" round size="small">{{ languageLabel(row.language) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="path" label="路径" min-width="220" show-overflow-tooltip />
          <el-table-column prop="version" label="版本" min-width="140" show-overflow-tooltip>
            <template #default="{ row }">{{ row.version || '—' }}</template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="discoverVisible = false">取消</el-button>
        <el-button
          type="primary"
          :disabled="!selectedCandidates.length"
          :loading="discoverSaving"
          @click="confirmDiscover"
        >
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, Refresh, EditPen, Download, Delete } from '@element-plus/icons-vue';
import * as interpretersApi from '@/api/interpreters';
import { interpreterLanguageOptions, languageLabel } from '@/utils/format';

const languageOptions = interpreterLanguageOptions;

const loading = ref(false);
const interpreters = ref([]);
const defaultTogglingId = ref('');

const formVisible = ref(false);
const formTitle = ref('新增解释器');
const saving = ref(false);
const editingId = ref('');
const form = ref({
  language: 'javascript',
  path: '',
  defaultArgs: []
});

const discoverVisible = ref(false);
const discoverTitle = ref('自动获取解释器');
const discoverLoading = ref(false);
const discoverSaving = ref(false);
const candidates = ref([]);
const selectedCandidates = ref([]);
const discoverTableRef = ref(null);
const discoverTargetId = ref('');

function formatDefaultArgs(args) {
  if (!args?.length) return '—';
  return args.join(' ');
}

function normalizeArgs(list) {
  return (list || []).map((s) => String(s).trim()).filter(Boolean);
}

async function loadList() {
  loading.value = true;
  try {
    interpreters.value = await interpretersApi.listInterpreters();
  } catch (e) {
    ElMessage.error(e.message || '加载解释器列表失败');
    interpreters.value = [];
  } finally {
    loading.value = false;
  }
}

function openCreate() {
  editingId.value = '';
  formTitle.value = '新增解释器';
  form.value = {
    language: 'javascript',
    path: '',
    defaultArgs: []
  };
  formVisible.value = true;
}

function openEdit(row) {
  editingId.value = row.id;
  formTitle.value = '编辑解释器';
  form.value = {
    language: row.language,
    path: row.path,
    defaultArgs: [...(row.defaultArgs || [])]
  };
  formVisible.value = true;
}

function resetForm() {
  editingId.value = '';
  form.value = { language: 'javascript', path: '', defaultArgs: [] };
}

function addArg() {
  form.value.defaultArgs.push('');
}

function removeArg(index) {
  form.value.defaultArgs.splice(index, 1);
}

async function submitForm() {
  const payload = {
    language: form.value.language,
    path: form.value.path.trim(),
    defaultArgs: normalizeArgs(form.value.defaultArgs)
  };
  if (!payload.path) {
    ElMessage.warning('请填写解释器路径');
    return;
  }
  saving.value = true;
  try {
    if (editingId.value) {
      await interpretersApi.updateInterpreter(editingId.value, payload);
      ElMessage.success('已更新');
    } else {
      await interpretersApi.createInterpreter(payload);
      ElMessage.success('已新增');
    }
    formVisible.value = false;
    await loadList();
  } catch (e) {
    ElMessage.error(e.message || '保存失败');
  } finally {
    saving.value = false;
  }
}

function openDiscover(row) {
  discoverTargetId.value = row?.id ?? '';
  discoverTitle.value = row ? '自动获取路径' : '自动获取解释器';
  discoverVisible.value = true;
  loadDiscover(row?.language ?? '');
}

function resetDiscover() {
  discoverTargetId.value = '';
  candidates.value = [];
  selectedCandidates.value = [];
}

async function loadDiscover(language) {
  discoverLoading.value = true;
  try {
    candidates.value = await interpretersApi.discoverInterpreters(language);
    selectedCandidates.value = [];
  } catch (e) {
    ElMessage.error(e.message || '扫描解释器失败');
    discoverVisible.value = false;
  } finally {
    discoverLoading.value = false;
  }
}

function onDiscoverSelectionChange(rows) {
  selectedCandidates.value = rows || [];
}

async function onDefaultChange(row, isDefault) {
  defaultTogglingId.value = row.id;
  try {
    await interpretersApi.setInterpreterDefault(row.id, isDefault);
    await loadList();
  } catch (e) {
    ElMessage.error(e.message || '更新默认解释器失败');
  } finally {
    defaultTogglingId.value = '';
  }
}

async function confirmDelete(row) {
  const label = row.path || languageLabel(row.language) || row.id;
  try {
    await ElMessageBox.confirm(`确定删除解释器「${label}」？此操作不可恢复。`, '删除解释器', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    });
  } catch {
    return;
  }
  try {
    await interpretersApi.deleteInterpreter(row.id);
    ElMessage.success('已删除');
    await loadList();
  } catch (e) {
    ElMessage.error(e.message || '删除失败');
  }
}

async function confirmDiscover() {
  const picked = selectedCandidates.value;
  if (!picked.length) return;

  discoverSaving.value = true;
  try {
    if (discoverTargetId.value) {
      if (picked.length > 1) {
        ElMessage.warning('更新路径时请选择一条记录');
        return;
      }
      await interpretersApi.updateInterpreterPath(
        discoverTargetId.value,
        picked[0].path,
        picked[0].version || ''
      );
      ElMessage.success('路径已更新');
    } else {
      const items = picked.map((c) => ({
        language: c.language,
        path: c.path,
        defaultArgs: c.defaultArgs || [],
        version: c.version || ''
      }));
      const result = await interpretersApi.batchCreateInterpreters(items);
      const n = result.created?.length ?? 0;
      const skipped = result.skipped ?? 0;
      if (n === 0 && skipped > 0) {
        ElMessage.info('所选解释器均已存在，未新增');
      } else {
        ElMessage.success(
          skipped > 0 ? `已新增 ${n} 条，跳过 ${skipped} 条重复路径` : `已新增 ${n} 条`
        );
      }
    }
    discoverVisible.value = false;
    await loadList();
  } catch (e) {
    ElMessage.error(e.message || '保存失败');
  } finally {
    discoverSaving.value = false;
  }
}

onMounted(loadList);
</script>

<style scoped>
.page-block {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.page-card {
  border-radius: var(--el-border-radius-base);
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
</style>
