<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from "vue";
import { pb } from "../lib/pb";
import { useDocuments, type Document } from "../composables/useDocuments";
import type { SimilarHit } from "../composables/useDocument";
import DocumentItem from "../components/DocumentItem.vue";

const { remove } = useDocuments();
const query = ref("");
const results = ref<Document[]>([]);
const similar = ref<SimilarHit[]>([]);
const loading = ref(false);
const error = ref("");
const searched = ref(false);
const semantic = ref(false);
let timer: ReturnType<typeof setTimeout> | undefined;
let unsub: (() => void) | undefined;

onMounted(async () => {
  unsub = await pb.collection("documents").subscribe("*", (e) => {
    const record = e.record as Document;
    const index = results.value.findIndex((d) => d.id === record.id);
    if (e.action === "delete") {
      if (index >= 0) results.value.splice(index, 1);
      similar.value = similar.value.filter((h) => h.item.id !== record.id);
      return;
    }
    if (index >= 0) results.value[index] = record;
    const hit = similar.value.find((h) => h.item.id === record.id);
    if (hit) hit.item = record;
  });
});

onUnmounted(() => unsub?.());

async function search() {
  const q = query.value.trim();
  if (!q) {
    results.value = [];
    similar.value = [];
    searched.value = false;
    return;
  }
  loading.value = true;
  error.value = "";
  try {
    const res = await pb.send("/api/search", {
      query: semantic.value ? { q, semantic: 1 } : { q },
    });
    results.value = res.items as Document[];
    similar.value = semantic.value ? ((res.similar ?? []) as SimilarHit[]) : [];
    searched.value = true;
  } catch (e) {
    error.value = String(e);
  } finally {
    loading.value = false;
  }
}

watch(query, () => {
  clearTimeout(timer);
  timer = setTimeout(search, 300);
});

watch(semantic, search);

async function onDelete(doc: { id: string; filename: string }) {
  if (!confirm(`Delete "${doc.filename}"?`)) return;
  await remove(doc.id);
  results.value = results.value.filter((d) => d.id !== doc.id);
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="font-headline text-headline-md text-on-surface">Search</h1>
      <p class="text-sm text-on-surface-variant mt-1">Search filenames, titles, summaries, categories, tags, entities, and extracted text.</p>
    </div>

    <div class="flex items-center gap-3 max-w-xl">
      <div class="relative flex-1">
        <span class="material-symbols-outlined absolute left-3 top-1/2 -translate-y-1/2 text-on-surface-variant text-lg">search</span>
        <input
          v-model="query"
          type="text"
          placeholder="Search documents..."
          class="w-full bg-surface border border-surface-container-highest rounded-lg pl-10 pr-4 py-2.5 text-sm text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:border-primary"
        />
      </div>
      <button
        class="text-xs font-medium px-3 py-2.5 rounded-lg border flex items-center gap-1.5 whitespace-nowrap transition-colors"
        :class="semantic
          ? 'bg-secondary-container text-on-secondary-container border-transparent'
          : 'border-outline-variant text-on-surface hover:bg-surface-container'"
        :title="semantic ? 'Also rank by meaning (embeds the query)' : 'Keyword search only'"
        @click="semantic = !semantic"
      >
        <span class="material-symbols-outlined text-[16px]">hub</span>
        Semantic
      </button>
    </div>

    <div class="bg-surface border border-surface-container-highest rounded-xl overflow-hidden flex-1 flex flex-col">
      <div
        class="grid grid-cols-[3fr_1fr_1.5fr_3fr_1.5fr_2rem] gap-4 bg-surface-container-low px-list-item-px py-list-item-py border-b border-surface-container-highest"
      >
        <div class="text-xs text-on-surface-variant uppercase tracking-wider">Filename</div>
        <div class="text-xs text-on-surface-variant uppercase tracking-wider">Size</div>
        <div class="text-xs text-on-surface-variant uppercase tracking-wider">Status</div>
        <div class="text-xs text-on-surface-variant uppercase tracking-wider">AI-Generated Title</div>
        <div class="text-xs text-on-surface-variant uppercase tracking-wider">Category</div>
        <div></div>
      </div>
      <div class="flex-1 overflow-auto flex flex-col divide-y divide-surface-container-highest">
        <p v-if="loading" class="p-6 text-sm text-on-surface-variant">Searching...</p>
        <p v-else-if="error" class="p-6 text-sm text-status-error">{{ error }}</p>
        <p v-else-if="!searched" class="p-6 text-sm text-on-surface-variant">
          Type to search across all documents.
        </p>
        <p v-else-if="results.length === 0" class="p-6 text-sm text-on-surface-variant">
          No documents match your search.
        </p>
        <DocumentItem v-for="doc in results" :key="doc.id" :document="doc" @delete="onDelete(doc)" />
      </div>
    </div>

    <div
      v-if="searched && similar.length"
      class="bg-surface border border-surface-container-highest rounded-xl overflow-hidden"
    >
      <div class="px-6 py-3 border-b border-surface-container-highest bg-surface-container-low flex items-center justify-between">
        <h2 class="text-sm font-medium text-on-surface">Semantically similar</h2>
        <span class="text-xs text-on-surface-variant">Matched by meaning, not exact words</span>
      </div>
      <div class="flex flex-col divide-y divide-surface-container-highest">
        <RouterLink
          v-for="hit in similar"
          :key="hit.item.id"
          :to="`/documents/${hit.item.id}`"
          class="flex items-center gap-4 px-list-item-px py-list-item-py hover:bg-surface-container-low transition-colors"
        >
          <span class="text-sm font-semibold text-on-surface truncate flex-1">{{ hit.item.filename }}</span>
          <span class="text-xs text-on-surface-variant truncate flex-1">{{ hit.item.title }}</span>
          <span class="text-xs text-on-surface-variant tabular-nums">{{ (hit.score * 100).toFixed(0) }}%</span>
        </RouterLink>
      </div>
    </div>
  </div>
</template>
