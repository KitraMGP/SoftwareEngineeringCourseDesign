<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    modelValue: string;
    disabled?: boolean;
    hint?: string;
    placeholder?: string;
    submitLabel?: string;
  }>(),
  {
    disabled: false,
    hint: '',
    placeholder: '在支持发送的会话中输入你的问题。',
    submitLabel: '发送消息'
  }
);

const emit = defineEmits<{
  'update:modelValue': [value: string];
  submit: [];
}>();

function handleKeydown(event: KeyboardEvent) {
  if ((event.metaKey || event.ctrlKey) && event.key === 'Enter' && !props.disabled) {
    emit('submit');
  }
}
</script>

<template>
  <div class="border-t border-slate-200/70 px-5 pb-5 pt-4 lg:px-8">
    <div class="rounded-[28px] border border-white/70 bg-white/88 p-4 shadow-soft">
      <textarea
        :value="modelValue"
        rows="5"
        :placeholder="placeholder"
        class="w-full resize-none border-none bg-transparent text-sm leading-7 text-slate-700 outline-none placeholder:text-slate-400"
        :disabled="disabled"
        @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
        @keydown="handleKeydown"
      />
      <div class="mt-4 flex flex-col gap-3 border-t border-slate-100 pt-4 md:flex-row md:items-center md:justify-between">
        <p class="text-xs leading-5 text-slate-400">
          {{ hint }}
        </p>
        <button
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
</template>
