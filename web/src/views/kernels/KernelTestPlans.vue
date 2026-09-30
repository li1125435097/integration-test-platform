<template>
  <div class="page-block page-fill">
    <el-row justify="space-between" align="middle">
      <el-text tag="h2" size="large">内核测试方案</el-text>
      <el-button :icon="Refresh" circle :loading="loading" @click="loadList" />
    </el-row>

    <el-card shadow="hover" class="page-card table-card">
      <template #header>
        <el-text type="info">内核测试页保存的方案，可回显、查看脚本或直接执行</el-text>
      </template>

      <el-table v-loading="loading" :data="plans" border stripe height="100%" style="width: 100%">
        <el-table-column type="index" label="#" width="56" align="center" />
        <el-table-column prop="name" label="方案名称" min-width="140" show-overflow-tooltip />
        <el-table-column label="脚本名称" min-width="160" show-overflow-tooltip>
          <template #default="{ row }">{{ scriptLabel(row) }}</template>
        </el-table-column>
        <el-table-column label="内核数量" width="100" align="center">
          <template #default="{ row }">{{ kernelCount(row) }}</template>
        </el-table-column>
        <el-table-column label="更新时间" width="172" align="center">
          <template #default="{ row }">{{ formatTime(row.updatedAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="340" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="goDetail(row)">详情</el-button>
            <el-divider direction="vertical" />
            <el-button type="primary" link :disabled="!row.form?.scriptId" @click="goScript(row)">
              脚本详情
            </el-button>
            <el-divider direction="vertical" />
            <el-button
              type="success"
              link
              :icon="VideoPlay"
              :loading="runningId === row.id"
              :disabled="Boolean(runningId) && runningId !== row.id"
              @click="onExecute(row)"
            >
              执行
            </el-button>
            <el-divider direction="vertical" />
            <el-button type="danger" link :icon="Delete" @click="confirmDelete(row)">删除</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无方案" />
        </template>
      </el-table>
    </el-card>

    <el-dialog v-model="concurrencyVisible" title="执行并发数" width="420px" align-center>
      <el-form label-width="72px" @submit.prevent>
        <el-form-item label="并发数">
          <el-input-number v-model="concurrency" :min="1" :max="pendingKernelCount" :step="1" />
        </el-form-item>
        <el-text type="info" size="small">
          已选 {{ pendingKernelCount }} 个内核，每个内核单独执行一次脚本。
        </el-text>
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
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Delete, Refresh, VideoPlay } from '@element-plus/icons-vue';
import * as kernelsApi from '@/api/kernels';
import * as scriptsApi from '@/api/scripts';
import KernelTestResultDialog from '@/components/KernelTestResultDialog.vue';
import { fingerprintPayload, launchArgs } from './formState';
import { formatTime } from '@/utils/format';

const router = useRouter();
const loading = ref(false);
const plans = ref([]);
const scripts = ref([]);
const scriptsReady = ref(false);
const runningId = ref('');
const concurrencyVisible = ref(false);
const concurrency = ref(1);
const pendingPlan = ref(null);
const resultVisible = ref(false);
const resultLoading = ref(false);
const runResults = ref([]);

const scriptNameById = computed(() => new Map(scripts.value.map((item) => [item.id, item.name])));
const pendingKernelCount = computed(() => kernelCount(pendingPlan.value));

function kernelCount(row) {
  return Array.isArray(row?.form?.kernels) ? row.form.kernels.length : 0;
}

function scriptLabel(row) {
  const id = row?.form?.scriptId;
  if (!id) return '—';
  if (!scriptsReady.value) return id;
  return scriptNameById.value.get(id) || '脚本不存在';
}

async function loadList() {
  loading.value = true;
  try {
    plans.value = await kernelsApi.listKernelTestPlans();
  } catch (e) {
    ElMessage.error(e.message || '加载方案失败');
    plans.value = [];
  }
  try {
    scripts.value = await scriptsApi.listScripts();
    scriptsReady.value = true;
  } catch (e) {
    ElMessage.error(e.message || '加载脚本失败');
    if (!scriptsReady.value) scripts.value = [];
  } finally {
    loading.value = false;
  }
}

function goDetail(row) {
  router.push({ name: 'kernel-test', query: { planId: row.id } });
}

function goScript(row) {
  const id = row.form?.scriptId;
  if (!id) return;
  router.push({ name: 'script-editor', params: { id } });
}

function onExecute(row) {
  if (runningId.value) return;
  const kernels = row.form?.kernels || [];
  if (!kernels.length) {
    ElMessage.warning('请选择内核');
    return;
  }
  if (!row.form?.scriptId) {
    ElMessage.warning('请选择脚本');
    return;
  }
  if (!String(row.form?.fingerprintText || '').trim()) {
    ElMessage.warning('指纹内容不能为空');
    return;
  }
  pendingPlan.value = row;
  if (kernels.length > 1) {
    concurrency.value = 1;
    concurrencyVisible.value = true;
    return;
  }
  execute(row, 1);
}

function confirmConcurrency() {
  const count = Number(concurrency.value);
  const max = pendingKernelCount.value;
  if (!Number.isInteger(count) || count < 1 || count > max) {
    ElMessage.warning('并发数无效');
    return;
  }
  const plan = pendingPlan.value;
  concurrencyVisible.value = false;
  if (plan) execute(plan, count);
}

async function execute(plan, count) {
  runningId.value = plan.id;
  resultVisible.value = true;
  resultLoading.value = true;
  runResults.value = [];
  try {
    runResults.value = await kernelsApi.runKernelTest({
      scriptId: plan.form.scriptId,
      concurrency: count,
      kernels: [...plan.form.kernels],
      args: launchArgs(plan.form),
      fingerprint: fingerprintPayload(plan.form)
    });
  } catch (e) {
    resultVisible.value = false;
    ElMessage.error(e.message || '执行失败');
  } finally {
    runningId.value = '';
    resultLoading.value = false;
  }
}

async function confirmDelete(row) {
  try {
    await ElMessageBox.confirm(`确定删除「${row.name}」？`, '删除方案', {
      confirmButtonText: '删除',
      cancelButtonText: '取消',
      type: 'warning'
    });
  } catch {
    return;
  }
  try {
    await kernelsApi.deleteKernelTestPlan(row.id);
    ElMessage.success(`已删除 ${row.name}`);
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
</style>
