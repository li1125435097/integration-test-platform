<template>
  <div class="page-block">
    <el-row justify="space-between" align="middle">
      <el-text tag="h2" size="large">脚本管理</el-text>
      <el-button type="primary" :icon="Plus" @click="goNew">新增脚本</el-button>
    </el-row>

    <el-card shadow="hover" class="page-card">
      <template #header>
        <el-row justify="space-between" align="middle">
          <el-text type="info">管理测试脚本与版本快照</el-text>
          <el-button :icon="Refresh" circle @click="loadList" :loading="loading" />
        </el-row>
      </template>

      <el-table v-loading="loading" :data="scripts" border stripe style="width: 100%">
        <el-table-column type="index" label="#" width="56" align="center" />
        <el-table-column prop="name" label="脚本名称" min-width="120" show-overflow-tooltip />
        <el-table-column prop="description" label="脚本描述" min-width="140" show-overflow-tooltip />
        <el-table-column label="脚本语言" width="112" align="center">
          <template #default="{ row }">
            <el-tag effect="plain" round size="small">{{ languageLabel(row.language) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="解释器" width="108" align="center">
          <template #default="{ row }">
            <el-tag
              :type="row.hasInterpreter ? 'success' : 'info'"
              effect="light"
              size="small"
            >
              {{ row.hasInterpreter ? '有' : '无' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="当前版本" width="120" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.currentVersion" type="primary" effect="plain" size="small">
              {{ row.currentVersion }}
            </el-tag>
            <el-tag v-else type="warning" effect="plain" size="small">未快照</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="172" align="center">
          <template #default="{ row }">{{ formatTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="260" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :icon="EditPen" @click="goEdit(row.id)">编辑</el-button>
            <el-divider direction="vertical" />
            <el-button type="primary" link :icon="Clock" @click="openVersionDialog(row)">
              版本变更
            </el-button>
            <el-divider direction="vertical" />
            <el-button type="danger" link :icon="Delete" @click="confirmDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无脚本">
            <el-button type="primary" :icon="Plus" @click="goNew">新建脚本</el-button>
          </el-empty>
        </template>
      </el-table>
    </el-card>

    <el-dialog
      v-model="versionVisible"
      :title="versionDialogTitle"
      width="720px"
      destroy-on-close
      align-center
    >
      <el-alert
        type="info"
        :closable="false"
        show-icon
        title="选择历史版本后，将用该版本内容覆盖当前脚本文件。"
        class="dialog-alert"
      />
      <div v-loading="versionLoading">
        <el-empty
          v-if="!versionLoading && !versions.length"
          description="暂无版本，请先在编辑页添加版本快照"
        />
        <el-table
          v-else
          class="version-dialog-table"
          :data="versions"
          highlight-current-row
          max-height="320"
          @current-change="onVersionSelect"
        >
          <el-table-column label="版本 ID" width="96" align="center">
            <template #default="{ row }">{{ versionDisplayId(row) }}</template>
          </el-table-column>
          <el-table-column label="备注" min-width="140" show-overflow-tooltip>
            <template #default="{ row }">
              {{ versionRemark(row) || '—' }}
            </template>
          </el-table-column>
          <el-table-column label="创建时间" width="168" align="center">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column label="操作" width="200" align="center" class-name="version-op-col">
            <template #default="{ row }">
              <div class="version-op-actions">
                <el-button type="primary" link :icon="EditPen" @click="openRemarkDialog(row)">
                  编辑备注
                </el-button>
                <el-divider direction="vertical" />
                <el-button type="danger" link :icon="Delete" @click="confirmDeleteVersion(row)">
                  删除
                </el-button>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>
      <template #footer>
        <el-button @click="versionVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!selectedVersionId" :loading="restoring" @click="confirmRestore">
          确认变更
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="remarkVisible"
      title="编辑备注"
      width="480px"
      destroy-on-close
      align-center
      @closed="resetRemarkForm"
    >
      <el-form label-width="80px" @submit.prevent>
        <el-form-item label="版本 ID">
          <el-text>{{ remarkVersionDisplayId || '—' }}</el-text>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="remarkDraft"
            type="textarea"
            :rows="3"
            placeholder="可选，例如本次变更说明"
            maxlength="200"
            show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="remarkVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingRemark" @click="submitRemark">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Plus, Refresh, EditPen, Clock, Delete } from '@element-plus/icons-vue';
import * as scriptsApi from '@/api/scripts';
import { formatTime, languageLabel } from '@/utils/format';
import {
  nextAutoSnapshotVersionId,
  TEMP_SNAPSHOT_REMARK,
  versionDisplayId,
  versionRemark
} from '@/utils/versions';

const router = useRouter();
const loading = ref(false);
const scripts = ref([]);

const versionVisible = ref(false);
const versionLoading = ref(false);
const restoring = ref(false);
const restoreScriptId = ref(null);
const restoreNeedsTempSnapshot = ref(false);
const versionDialogTitle = ref('版本变更');
const versions = ref([]);
const selectedVersionId = ref('');

const remarkVisible = ref(false);
const savingRemark = ref(false);
const remarkVersionId = ref('');
const remarkVersionDisplayId = ref('');
const remarkDraft = ref('');

async function loadList() {
  loading.value = true;
  try {
    scripts.value = await scriptsApi.listScripts();
  } catch (e) {
    ElMessage.error(e.message || '加载脚本列表失败');
    scripts.value = [];
  } finally {
    loading.value = false;
  }
}

function goNew() {
  router.push({ name: 'script-editor', params: { id: 'new' } });
}

function goEdit(id) {
  router.push({ name: 'script-editor', params: { id } });
}

async function openVersionDialog(row) {
  restoreScriptId.value = row.id;
  restoreNeedsTempSnapshot.value = !row.currentVersion;
  selectedVersionId.value = '';
  versionDialogTitle.value = `版本变更 · ${row.name || row.id}`;
  versionVisible.value = true;
  versionLoading.value = true;
  try {
    versions.value = await scriptsApi.listVersions(row.id);
  } catch (e) {
    ElMessage.error(e.message || '加载版本列表失败');
    versionVisible.value = false;
  } finally {
    versionLoading.value = false;
  }
}

function onVersionSelect(row) {
  selectedVersionId.value = row?.id ?? '';
}

async function reloadVersions() {
  if (!restoreScriptId.value) return;
  versionLoading.value = true;
  try {
    versions.value = await scriptsApi.listVersions(restoreScriptId.value);
  } catch (e) {
    ElMessage.error(e.message || '加载版本列表失败');
  } finally {
    versionLoading.value = false;
  }
}

function openRemarkDialog(row) {
  remarkVersionId.value = row.id;
  remarkVersionDisplayId.value = versionDisplayId(row);
  remarkDraft.value = versionRemark(row);
  remarkVisible.value = true;
}

function resetRemarkForm() {
  remarkVersionId.value = '';
  remarkVersionDisplayId.value = '';
  remarkDraft.value = '';
}

async function submitRemark() {
  if (!restoreScriptId.value || !remarkVersionId.value) return;
  savingRemark.value = true;
  try {
    await scriptsApi.updateVersionRemark(
      restoreScriptId.value,
      remarkVersionId.value,
      remarkDraft.value
    );
    ElMessage.success('备注已更新');
    remarkVisible.value = false;
    await reloadVersions();
  } catch (e) {
    ElMessage.error(e.message || '更新备注失败');
  } finally {
    savingRemark.value = false;
  }
}

async function confirmDeleteVersion(row) {
  const label = versionDisplayId(row) || row.id;
  try {
    await ElMessageBox.confirm(
      `确定删除版本「${label}」？快照文件将一并删除，此操作不可恢复。`,
      '删除版本',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    );
  } catch {
    return;
  }
  try {
    await scriptsApi.deleteVersion(restoreScriptId.value, row.id);
    if (selectedVersionId.value === row.id) {
      selectedVersionId.value = '';
    }
    ElMessage.success('已删除');
    await reloadVersions();
  } catch (e) {
    ElMessage.error(e.message || '删除失败');
  }
}

async function confirmRestore() {
  if (!restoreScriptId.value || !selectedVersionId.value) return;
  restoring.value = true;
  try {
    if (restoreNeedsTempSnapshot.value) {
      const list = await scriptsApi.listVersions(restoreScriptId.value);
      const versionId = nextAutoSnapshotVersionId(list);
      await scriptsApi.addVersion(restoreScriptId.value, versionId, TEMP_SNAPSHOT_REMARK);
    }
    await scriptsApi.restoreVersion(restoreScriptId.value, selectedVersionId.value);
    ElMessage.success(
      restoreNeedsTempSnapshot.value ? '已保存临时快照并完成版本变更' : '版本已变更'
    );
    versionVisible.value = false;
    await loadList();
  } catch (e) {
    ElMessage.error(e.message || '版本变更失败');
  } finally {
    restoring.value = false;
  }
}

async function confirmDelete(row) {
  try {
    await ElMessageBox.confirm(
      `确定删除脚本「${row.name || row.id}」？此操作不可恢复。`,
      '删除脚本',
      { type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消' }
    );
  } catch {
    return;
  }
  try {
    await scriptsApi.deleteScript(row.id);
    ElMessage.success('已删除');
    await loadList();
  } catch (e) {
    ElMessage.error(e.message || '删除失败');
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

.dialog-alert {
  margin-bottom: 16px;
}

.version-op-actions {
  display: inline-flex;
  align-items: center;
  flex-wrap: nowrap;
  white-space: nowrap;
}

.version-op-actions :deep(.el-button.is-link) {
  padding-left: 6px;
  padding-right: 6px;
}

.version-op-actions :deep(.el-divider--vertical) {
  margin: 0 2px;
}

.version-dialog-table :deep(.version-op-col .cell) {
  white-space: nowrap;
  overflow: visible;
}
</style>
