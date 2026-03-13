<script setup lang="ts">
import { computed } from 'vue';

const props = withDefaults(
  defineProps<{
    tone?: 'info' | 'success' | 'warning' | 'danger';
    dot?: boolean;
  }>(),
  {
    tone: 'info',
    dot: true
  }
);

const classes = computed(() => `status-pill status-pill--${props.tone}`);
</script>

<template>
  <span :class="classes">
    <span v-if="dot" class="status-pill__dot" />
    <slot />
  </span>
</template>
