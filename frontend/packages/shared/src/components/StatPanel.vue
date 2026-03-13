<script setup lang="ts">
defineProps<{
  label: string;
  value: string;
  helper?: string;
}>();
</script>

<template>
  <div class="rounded-[24px] border border-white/50 bg-white/70 p-5 shadow-soft backdrop-blur">
    <p class="text-xs uppercase tracking-[0.3em] text-slate-400">
      {{ label }}
    </p>
    <p class="mt-4 font-serif text-3xl text-slate-900">
      {{ value }}
    </p>
    <p v-if="helper" class="mt-3 text-sm text-slate-500">
      {{ helper }}
    </p>
  </div>
</template>
