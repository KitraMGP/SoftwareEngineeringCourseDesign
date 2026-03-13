<script setup lang="ts">
defineProps<{
  title: string;
  description: string;
  actionLabel?: string;
}>();

const emit = defineEmits<{
  action: [];
}>();
</script>

<template>
  <div
    class="flex min-h-[260px] flex-col items-center justify-center gap-5 rounded-[28px] border border-white/50 bg-white/65 px-8 py-10 text-center shadow-soft backdrop-blur"
  >
    <div
      class="flex h-16 w-16 items-center justify-center rounded-full bg-slate-900 text-sm font-semibold uppercase tracking-[0.25em] text-white"
    >
      AI
    </div>
    <div class="space-y-2">
      <h3 class="font-serif text-2xl text-slate-900">
        {{ title }}
      </h3>
      <p class="max-w-xl text-sm leading-6 text-slate-500 md:text-base">
        {{ description }}
      </p>
    </div>
    <button
      v-if="actionLabel"
      type="button"
      class="rounded-full bg-slate-900 px-5 py-3 text-sm font-medium text-white transition hover:-translate-y-0.5 hover:bg-slate-800"
      @click="emit('action')"
    >
      {{ actionLabel }}
    </button>
  </div>
</template>
