<script setup lang="ts">
defineProps<{ status: string }>();

const colors: Record<string, string> = {
  pending: "bg-status-pending",
  processing: "bg-status-processing",
  done: "bg-status-done",
  error: "bg-status-error",
};
</script>

<template>
  <span
    class="inline-flex items-center gap-1.5 bg-surface-container px-2 py-0.5 rounded-full border border-surface-container-highest"
  >
    <span
      class="w-2 h-2 rounded-full"
      :class="[colors[status] ?? 'bg-status-pending', status === 'processing' ? 'animate-pulse' : '']"
    ></span>
    <span class="text-xs font-medium text-on-surface capitalize">{{ status }}</span>
  </span>
</template>
