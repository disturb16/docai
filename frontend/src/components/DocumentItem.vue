<script setup lang="ts">
import { progressLabel, type Document } from "../composables/useDocuments";
import StatusBadge from "./StatusBadge.vue";

defineProps<{ document: Document }>();
const emit = defineEmits<{ delete: [] }>();

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
</script>

<template>
  <RouterLink
    :to="`/documents/${document.id}`"
    class="grid grid-cols-[3fr_1fr_1.5fr_3fr_1.5fr_2rem] gap-4 items-center px-list-item-px py-list-item-py hover:bg-surface-container-low transition-colors duration-150 group cursor-pointer bg-surface"
  >
    <div class="flex items-center gap-3 overflow-hidden">
      <span class="text-outline text-lg">▤</span>
      <span class="text-sm font-semibold text-on-surface truncate">{{ document.filename }}</span>
    </div>
    <div class="text-xs text-on-surface-variant">{{ formatSize(document.size) }}</div>
    <div class="flex items-center">
      <StatusBadge :status="document.status" />
    </div>
    <div
      v-if="document.status === 'done'"
      class="text-sm text-on-surface truncate"
    >
      {{ document.title }}
    </div>
    <div
      v-else-if="document.status === 'error' || (document.status === 'pending' && document.error)"
      class="text-sm text-status-error truncate"
    >
      {{ document.error }}
    </div>
    <div v-else class="text-sm text-on-surface-variant italic truncate">
      {{ document.status === "processing" ? progressLabel(document.progress) : "Queued for processing..." }}
    </div>
    <div class="flex items-center">
      <span
        v-if="document.category"
        class="text-xs bg-tertiary-container text-on-tertiary-container px-2 py-0.5 rounded"
      >
        {{ document.category }}
      </span>
      <span
        v-else
        class="text-xs bg-surface-container-low border border-surface-container-highest text-on-surface-variant px-2 py-0.5 rounded"
      >
        --
      </span>
    </div>
    <button
      class="text-on-surface-variant hover:text-status-error text-base leading-none justify-self-center"
      title="Delete"
      @click.prevent.stop="emit('delete')"
    >
      ×
    </button>
  </RouterLink>
</template>
