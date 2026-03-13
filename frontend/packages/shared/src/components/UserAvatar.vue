<script setup lang="ts">
import { computed, ref, watch } from 'vue';

const props = withDefaults(
  defineProps<{
    label?: string | null;
    avatarUrl?: string | null;
    size?: number;
    tone?: 'dark' | 'light';
  }>(),
  {
    label: '',
    avatarUrl: null,
    size: 44,
    tone: 'dark'
  }
);

const loadFailed = ref(false);

watch(
  () => props.avatarUrl,
  () => {
    loadFailed.value = false;
  }
);

const initials = computed(() => {
  const raw = (props.label || '访客').trim();
  const compact = raw.replace(/\s+/g, '');

  if (!compact) {
    return '访';
  }

  return compact.slice(0, 2).toUpperCase();
});

const avatarStyle = computed(() => ({
  width: `${props.size}px`,
  height: `${props.size}px`
}));

const fallbackClasses = computed(() =>
  props.tone === 'light'
    ? 'border-white/60 bg-white text-slate-900'
    : 'border-slate-900/10 bg-slate-900 text-white'
);
</script>

<template>
  <span
    class="inline-flex shrink-0 items-center justify-center overflow-hidden rounded-full border text-xs font-semibold uppercase tracking-[0.14em] shadow-soft"
    :class="fallbackClasses"
    :style="avatarStyle"
  >
    <img
      v-if="avatarUrl && !loadFailed"
      :src="avatarUrl"
      :alt="label || 'avatar'"
      class="h-full w-full object-cover"
      @error="loadFailed = true"
    >
    <span v-else>{{ initials }}</span>
  </span>
</template>
