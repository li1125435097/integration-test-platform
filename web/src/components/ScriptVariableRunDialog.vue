<template>
  <el-dialog
    :model-value="modelValue"
    title="执行变量"
    width="480px"
    destroy-on-close
    align-center
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <el-form label-width="auto" @submit.prevent>
      <el-form-item v-for="item in drafts" :key="item.name" :label="item.name">
        <el-input v-model="item.value" clearable />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" @click="confirm">执行</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch } from 'vue';
import { overridesFromDefaults } from '@/utils/scriptVariables';

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  variables: { type: Array, default: () => [] }
});

const emit = defineEmits(['update:modelValue', 'confirm']);

const drafts = ref([]);

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return;
    drafts.value = (props.variables || []).map((v) => ({
      name: v.name,
      value: v.value ?? ''
    }));
  }
);

function confirm() {
  emit('confirm', overridesFromDefaults(props.variables, drafts.value));
  emit('update:modelValue', false);
}
</script>
