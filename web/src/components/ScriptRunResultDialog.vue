<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="720px"
    destroy-on-close
    align-center
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div v-loading="loading" class="run-dialog-body">
      <el-row v-if="result && !loading" :gutter="12" class="run-meta">
        <el-col :span="8">
          <el-text type="info" size="small">退出码</el-text>
          <div>
            <el-tag :type="result.success ? 'success' : 'danger'" effect="plain" size="small">
              {{ result.exitCode }}
            </el-tag>
          </div>
        </el-col>
        <el-col :span="8">
          <el-text type="info" size="small">耗时</el-text>
          <div>
            <el-text>{{ result.durationMs ?? 0 }} ms</el-text>
          </div>
        </el-col>
        <el-col v-if="result.recordId" :span="8">
          <el-text type="info" size="small">记录 ID</el-text>
          <div>
            <el-text truncated>{{ result.recordId }}</el-text>
          </div>
        </el-col>
      </el-row>

      <el-alert
        v-if="result?.error"
        type="error"
        :closable="false"
        show-icon
        :title="result.error"
        class="run-alert"
      />

      <el-tabs v-if="result && !loading" model-value="stdout">
        <el-tab-pane label="标准输出" name="stdout">
          <pre class="run-output">{{ result.stdout || '（无输出）' }}</pre>
        </el-tab-pane>
        <el-tab-pane label="标准错误" name="stderr">
          <pre class="run-output is-stderr">{{ result.stderr || '（无输出）' }}</pre>
        </el-tab-pane>
      </el-tabs>
    </div>
    <template #footer>
      <el-button type="primary" @click="$emit('update:modelValue', false)">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
defineProps({
  modelValue: { type: Boolean, default: false },
  title: { type: String, default: '执行结果' },
  loading: { type: Boolean, default: false },
  result: { type: Object, default: null }
});

defineEmits(['update:modelValue']);
</script>

<style scoped>
.run-dialog-body {
  min-height: 120px;
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
  max-height: 360px;
  overflow: auto;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
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
