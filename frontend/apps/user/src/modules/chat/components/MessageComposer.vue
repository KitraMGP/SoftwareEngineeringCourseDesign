<script setup lang="ts">
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
    placeholder: '在支持发送的会话中输入你的问题。',
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

function handleKeydown(event: KeyboardEvent) {
  if (
    (event.metaKey || event.ctrlKey) &&
    event.key === 'Enter' &&
    !props.disabled &&
    !props.showStopAction
  ) {
    emit('submit');
  }
}
</script>

<template>
  <div class="shrink-0 border-t border-slate-200/70 px-5 pb-4 pt-3 lg:px-8">
    <div class="rounded-[28px] border border-white/70 bg-white/88 p-3.5 shadow-soft">
      <textarea
        :value="modelValue"
        rows="3"
        :placeholder="placeholder"
        class="max-h-32 min-h-[84px] w-full resize-none overflow-y-auto border-none bg-transparent text-sm leading-6 text-slate-700 outline-none placeholder:text-slate-400"
        :disabled="disabled"
        @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
        @keydown="handleKeydown"
      />
      <div class="mt-3 flex flex-col gap-3 border-t border-slate-100 pt-3 md:flex-row md:items-center md:justify-between">
        <p class="text-xs leading-5 text-slate-400">
          {{ hint }}
        </p>
        <div class="flex items-center justify-end gap-3">
          <button
            v-if="showStopAction"
            type="button"
            class="rounded-full border border-rose-200 bg-rose-50 px-5 py-3 text-sm font-semibold text-rose-600 transition hover:border-rose-300 hover:bg-rose-100 disabled:cursor-not-allowed disabled:border-slate-200 disabled:bg-slate-100 disabled:text-slate-400"
            :disabled="stopPending"
            @click="emit('stop')"
          >
            {{ stopPending ? '正在停止...' : stopLabel }}
          </button>
          <button
            v-else
            type="button"
            class="rounded-full bg-slate-900 px-5 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-300"
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
