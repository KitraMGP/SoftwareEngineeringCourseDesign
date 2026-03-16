<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';

const props = withDefaults(
  defineProps<{
    modelValue: string;
    disabled?: boolean;
    hint?: string;
    placeholder?: string;
    submitLabel?: string;
    showStopAction?: boolean;
    stopLabel?: string;
    stopPending?: boolean;
  }>(),
  {
    disabled: false,
    hint: '',
    placeholder: '输入问题。Enter 发送，Shift+Enter 换行。',
    submitLabel: '发送消息',
    showStopAction: false,
    stopLabel: '停止生成',
    stopPending: false
  }
);

const emit = defineEmits<{
  'update:modelValue': [value: string];
  submit: [];
  stop: [];
}>();

const textareaRef = ref<HTMLTextAreaElement>();

function syncTextareaHeight() {
  if (!textareaRef.value) {
    return;
  }

  textareaRef.value.style.height = '0px';
  textareaRef.value.style.height = `${Math.min(Math.max(textareaRef.value.scrollHeight, 44), 128)}px`;
}

function handleKeydown(event: KeyboardEvent) {
  if (
    event.key === 'Enter' &&
    !event.shiftKey &&
    !event.isComposing &&
    !props.disabled &&
    !props.showStopAction
  ) {
    event.preventDefault();
    emit('submit');
  }
}

watch(
  () => props.modelValue,
  async () => {
    await nextTick();
    syncTextareaHeight();
  },
  {
    immediate: true
  }
);
</script>

<template>
  <div class="shrink-0 border-t border-slate-200/70 px-5 pb-3 pt-2.5 lg:px-8">
    <div class="rounded-[22px] border border-white/78 bg-white/92 p-2.5 shadow-soft">
      <textarea
        ref="textareaRef"
        :value="modelValue"
        rows="1"
        :placeholder="placeholder"
        class="max-h-32 w-full resize-none overflow-y-auto border-none bg-transparent text-sm leading-6 text-slate-700 outline-none placeholder:text-slate-400"
        :disabled="disabled"
        @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
        @keydown="handleKeydown"
      />
      <div class="mt-2.5 flex flex-col gap-2.5 border-t border-slate-100 pt-2.5 md:flex-row md:items-center md:justify-between">
        <p class="pr-2 text-[11px] leading-5 text-slate-400">
          {{ hint }}
        </p>
        <div class="flex items-center justify-end gap-3">
          <button
            v-if="showStopAction"
            type="button"
            class="rounded-full border border-rose-200 bg-rose-50 px-4 py-2 text-sm font-semibold text-rose-600 transition hover:border-rose-300 hover:bg-rose-100 disabled:cursor-not-allowed disabled:border-slate-200 disabled:bg-slate-100 disabled:text-slate-400"
            :disabled="stopPending"
            @click="emit('stop')"
          >
            {{ stopPending ? '正在停止...' : stopLabel }}
          </button>
          <button
            v-else
            type="button"
            class="rounded-full bg-slate-900 px-4 py-2 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-300"
            :disabled="disabled"
            @click="emit('submit')"
          >
            {{ submitLabel }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
