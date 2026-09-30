<template>
  <el-dialog
    :model-value="modelValue"
    title="执行结果"
    width="760px"
    align-center
    @update:model-value="emit('update:modelValue', $event)"
  >
    <div v-loading="loading" class="run-dialog-body">
      <el-text v-if="results.length && !loading" class="run-summary">
        成功 {{ successCount }} / 失败 {{ failCount }}
      </el-text>
      <el-collapse v-if="results.length && !loading" v-model="openResults">
        <el-collapse-item
          v-for="(item, index) in results"
          :key="`${item.kernel}-${index}`"
          :name="String(index)"
          :title="item.kernel || `内核 ${index + 1}`"
        >
          <el-row :gutter="12" class="run-meta">
            <el-col :span="8">
              <el-text type="info" size="small">退出码</el-text>
              <div>
                <el-tag :type="item.success ? 'success' : 'danger'" effect="plain" size="small">
                  {{ item.exitCode }}
                </el-tag>
              </div>
            </el-col>
            <el-col :span="8">
              <el-text type="info" size="small">耗时</el-text>
              <div>
                <el-text>{{ item.durationMs ?? 0 }} ms</el-text>
              </div>
            </el-col>
          </el-row>
          <el-alert
            v-if="item.error"
            type="error"
            :closable="false"
            show-icon
            :title="item.error"
            class="run-alert"
          />
          <el-tabs model-value="stdout">
            <el-tab-pane label="标准输出" name="stdout">
              <pre class="run-output">{{ item.stdout || '（无输出）' }}</pre>
            </el-tab-pane>
            <el-tab-pane label="标准错误" name="stderr">
              <pre class="run-output is-stderr">{{ item.stderr || '（无输出）' }}</pre>
            </el-tab-pane>
          </el-tabs>
        </el-collapse-item>
      </el-collapse>
    </div>
    <template #footer>
      <el-button type="primary" @click="emit('update:modelValue', false)">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, ref, watch } from 'vue';

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
  results: { type: Array, default: () => [] }
});

const emit = defineEmits(['update:modelValue']);

const openResults = ref([]);

const successCount = computed(() => props.results.filter((item) => item.success).length);
const failCount = computed(() => props.results.length - successCount.value);

watch(
  () => props.results,
  (list) => {
    openResults.value = (list || []).map((_, index) => String(index));
  },
  { immediate: true }
);
</script>

<style scoped>
.run-dialog-body {
  min-height: 120px;
}

.run-summary {
  display: block;
  margin-bottom: 12px;
}

.run-meta {
  margin-bottom: 12px;
}

.run-alert {
  margin-bottom: 12px;
}

.run-output {
  margin: 0;
  padding: 12px;
  max-height: 240px;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  font-size: 13px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
  background: var(--el-fill-color-light);
  border-radius: var(--el-border-radius-base);
}

.run-output.is-stderr {
  color: var(--el-color-danger);
}
</style>
