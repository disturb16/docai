<script setup lang="ts">
import { computed, onMounted, onUnmounted } from "vue";
import { progressLabel, useDocuments } from "../composables/useDocuments";
import StatusBadge from "../components/StatusBadge.vue";

const { documents, fetchAll, subscribe } = useDocuments();

let unsub: (() => void) | undefined;

onMounted(async () => {
  fetchAll();
  unsub = await subscribe();
});

onUnmounted(() => unsub?.());

const processing = computed(() => documents.value.filter((d) => d.status === "processing"));
const embedding = computed(() =>
  documents.value.filter((d) => d.embed_status === "pending" || d.embed_status === "processing"),
);
const active = computed(() => processing.value.length + embedding.value.length);
const queue = computed(() =>
  documents.value.filter((d) => d.status === "pending" || d.status === "processing"),
);

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-4">
      <div>
        <h1 class="font-headline text-headline-md text-on-surface">Activity Monitor</h1>
        <p class="text-sm text-on-surface-variant mt-1">
          Real-time status of local AI inference and document processing.
        </p>
      </div>
      <div
        class="flex items-center gap-2 bg-surface-container px-3 py-1.5 rounded-full border border-surface-container-highest w-fit"
      >
        <span
          class="w-2 h-2 rounded-full"
          :class="active ? 'bg-status-processing animate-pulse' : 'bg-status-done'"
        ></span>
        <span class="text-xs text-on-surface-variant uppercase tracking-wider">
          {{ active ? "Engine Active" : "Engine Idle" }}
        </span>
      </div>
    </div>

    <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl p-6 shadow-sm">
      <h3 class="text-xs text-on-surface-variant uppercase tracking-widest mb-4">Now Processing</h3>
      <div v-if="processing.length" class="flex flex-col gap-5">
        <div v-for="doc in processing" :key="doc.id" class="flex justify-between items-start gap-4">
          <div class="flex items-center gap-3 min-w-0">
            <span class="material-symbols-outlined text-primary">description</span>
            <RouterLink
              :to="`/documents/${doc.id}`"
              class="font-headline text-headline-sm text-on-surface truncate hover:text-primary transition-colors"
            >
              {{ doc.filename }}
            </RouterLink>
          </div>
          <div class="flex items-center gap-1.5 bg-surface-container-high px-2 py-1 rounded-md shrink-0">
            <span class="material-symbols-outlined text-[16px] text-status-processing animate-spin">sync</span>
            <span class="text-xs text-on-surface-variant">{{ progressLabel(doc.progress) }}</span>
          </div>
        </div>
      </div>
      <div v-else class="flex items-center gap-3 text-sm text-on-surface-variant py-4">
        <span class="material-symbols-outlined">check_circle</span>
        No document is currently being processed.
      </div>
    </div>

    <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl p-6 shadow-sm">
      <h3 class="text-xs text-on-surface-variant uppercase tracking-widest mb-4">Embedding Queue</h3>
      <div v-if="embedding.length" class="flex flex-col gap-5">
        <div v-for="doc in embedding" :key="doc.id" class="flex justify-between items-start gap-4">
          <div class="flex items-center gap-3 min-w-0">
            <span class="material-symbols-outlined text-primary">hub</span>
            <RouterLink
              :to="`/documents/${doc.id}`"
              class="font-headline text-headline-sm text-on-surface truncate hover:text-primary transition-colors"
            >
              {{ doc.filename }}
            </RouterLink>
          </div>
          <div class="flex items-center gap-1.5 bg-surface-container-high px-2 py-1 rounded-md shrink-0">
            <span class="material-symbols-outlined text-[16px] text-status-processing animate-spin">sync</span>
            <span class="text-xs text-on-surface-variant">
              {{ doc.embed_status === "processing" ? "Embedding..." : "Queued for embedding..." }}
            </span>
          </div>
        </div>
      </div>
      <div v-else class="flex items-center gap-3 text-sm text-on-surface-variant py-4">
        <span class="material-symbols-outlined">check_circle</span>
        No document is currently being embedded.
      </div>
    </div>

    <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl shadow-sm overflow-hidden">
      <div class="px-6 py-4 border-b border-surface-container-highest bg-surface-container-low flex justify-between items-center">
        <h3 class="font-headline text-headline-sm text-on-surface">Processing Queue</h3>
        <span class="text-xs text-on-surface-variant">{{ queue.length }} item{{ queue.length === 1 ? "" : "s" }} pending</span>
      </div>
      <div v-if="queue.length" class="overflow-x-auto">
        <table class="w-full text-left border-collapse">
          <thead>
            <tr class="bg-surface-container-lowest text-xs text-on-surface-variant border-b border-surface-container-highest uppercase">
              <th class="py-3 px-6 font-medium">Document</th>
              <th class="py-3 px-6 font-medium">Size</th>
              <th class="py-3 px-6 font-medium">Status</th>
            </tr>
          </thead>
          <tbody class="text-sm text-on-surface">
            <tr
              v-for="doc in queue"
              :key="doc.id"
              class="border-b border-surface-container-highest last:border-b-0 hover:bg-surface-container-low transition-colors duration-150"
            >
              <td class="py-list-item-py px-6">
                <RouterLink :to="`/documents/${doc.id}`" class="flex items-center gap-3">
                  <span class="material-symbols-outlined text-outline">
                    {{ doc.mime.startsWith("image/") ? "image" : "description" }}
                  </span>
                  <span class="font-medium">{{ doc.filename }}</span>
                </RouterLink>
              </td>
              <td class="py-list-item-py px-6 text-xs text-on-surface-variant">{{ formatSize(doc.size) }}</td>
              <td class="py-list-item-py px-6">
                <StatusBadge :status="doc.status" />
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-else class="px-6 py-6 text-sm text-on-surface-variant">Queue is empty.</p>
    </div>
  </div>
</template>
