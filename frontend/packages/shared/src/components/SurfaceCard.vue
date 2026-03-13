<script setup lang="ts">
withDefaults(
  defineProps<{
    tone?: 'default' | 'soft' | 'night';
    padded?: boolean;
  }>(),
  {
    tone: 'default',
    padded: true
  }
);
</script>

<template>
  <section
    class="surface-card"
    :class="[`surface-card--${tone}`, padded ? 'surface-card--padded' : '']"
  >
    <slot />
  </section>
</template>
