<script setup lang="ts">
defineProps<{
  eyebrow?: string;
  title: string;
  description?: string;
}>();
</script>

<template>
  <div class="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
    <div class="space-y-2">
      <p v-if="eyebrow" class="text-xs font-semibold uppercase tracking-[0.35em] text-slate-400">
        {{ eyebrow }}
      </p>
      <h2 class="font-serif text-2xl text-slate-900 md:text-3xl">
        {{ title }}
      </h2>
      <p v-if="description" class="max-w-2xl text-sm leading-6 text-slate-500 md:text-base">
        {{ description }}
      </p>
    </div>

    <div v-if="$slots.actions" class="flex items-center gap-3">
      <slot name="actions" />
    </div>
  </div>
</template>
