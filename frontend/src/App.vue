<template>
  <app-layout ref="layoutRef" />
</template>

<script lang="ts" setup>
import { ref, provide } from 'vue';
import AppLayout from './components/AppLayout.vue';

const layoutRef = ref();

// Provide the showNotification method to all child components
provide('showNotification', (text: string, color = 'success') => {
  layoutRef.value?.showNotification(text, color);
});
</script>
