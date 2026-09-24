<template>
  <div class="page-block">
    <el-row justify="space-between" align="middle">
      <el-text tag="h2" size="large">执行记录</el-text>
      <el-button :icon="Refresh" circle @click="loadList" :loading="loading" />
    </el-row>

    <el-card shadow="hover" class="page-card">
      <template #header>
        <el-text type="info">脚本列表「执行」产生的运行记录（编辑页试运行为不入库）</el-text>
      </template>

      <el-table v-loading="loading" :data="records" border stripe style="width: 100%">
        <el-table-column type="index" label="#" width="56" align="center" />
        <el-table-column prop="scriptName" label="脚本名称" min-width="120" show-overflow-tooltip />
        <el-table-column label="语言" width="112" align="center">
          <template #default="{ row }">
            <el-tag effect="plain" round size="small">{{ languageLabel(row.language) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="结果" width="96" align="center">
          <template #default="{ row }">
            <el-tag :type="row.success ? 'success' : 'danger'" effect="plain" size="small">
              {{ row.success ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="退出码" width="88" align="center" prop="exitCode" />
        <el-table-column label="耗时" width="100" align="center">
          <template #default="{ row }">{{ row.durationMs }} ms</template>
        </el-table-column>
        <el-table-column label="执行时间" width="172" align="center">
          <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="100" align="center" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="openDetail(row)">详情</el-button>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无执行记录" />
        </template>
      </el-table>
    </el-card>

    <ScriptRunResultDialog
      v-model="detailVisible"
      :title="detailTitle"
      :loading="false"
      :result="detailResult"
    />
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { Refresh } from '@element-plus/icons-vue';
import * as executionsApi from '@/api/executions';
import { formatTime, languageLabel } from '@/utils/format';
import ScriptRunResultDialog from '@/components/ScriptRunResultDialog.vue';

const loading = ref(false);
const records = ref([]);

const detailVisible = ref(false);
const detailTitle = ref('执行详情');
const detailResult = ref(null);

async function loadList() {
  loading.value = true;
  try {
    records.value = await executionsApi.listExecutionRecords();
  } catch (e) {
    ElMessage.error(e.message || '加载执行记录失败');
    records.value = [];
  } finally {
    loading.value = false;
  }
}

function openDetail(row) {
  detailTitle.value = `执行详情 · ${row.scriptName || row.scriptId || '—'}`;
  detailResult.value = {
    exitCode: row.exitCode,
    success: row.success,
    durationMs: row.durationMs,
    stdout: row.stdout,
    stderr: row.stderr,
    error: row.error,
    recordId: row.id
  };
  detailVisible.value = true;
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
