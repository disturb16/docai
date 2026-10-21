<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { pb } from "../lib/pb";

interface DailyBucket {
  date: string;
  total: number;
  done: number;
  error: number;
}

interface ProviderStats {
  count: number;
  done: number;
  error: number;
  avg_total_ms: number;
  avg_extract_ms: number;
  avg_ocr_ms: number;
  avg_classify_ms: number;
  avg_retries: number;
}

interface Analytics {
  totals: Record<string, number>;
  daily: DailyBucket[];
  providers: Record<string, ProviderStats>;
}

const analytics = ref<Analytics>();
const loading = ref(true);
const error = ref("");

onMounted(async () => {
  try {
    analytics.value = await pb.send<Analytics>("/api/docai/analytics", { method: "GET" });
  } catch (e) {
    error.value = String(e);
  } finally {
    loading.value = false;
  }
});

const totalDocs = computed(() => {
  const t = analytics.value?.totals ?? {};
  return Object.values(t).reduce((a, b) => a + b, 0);
});

const doneRate = computed(() => {
  const t = analytics.value?.totals;
  if (!t || !totalDocs.value) return 0;
  return Math.round((t.done / totalDocs.value) * 100);
});

const errorRate = computed(() => {
  const t = analytics.value?.totals;
  if (!t || !totalDocs.value) return 0;
  return Math.round((t.error / totalDocs.value) * 100);
});

const maxDaily = computed(() =>
  Math.max(1, ...(analytics.value?.daily.map((d) => d.total) ?? [1])),
);

const providerList = computed(() => {
  const p = analytics.value?.providers ?? {};
  return Object.entries(p)
    .map(([name, s]) => ({ name, ...s }))
    .sort((a, b) => b.count - a.count);
});

function maxStage(all: ProviderStats[], key: keyof ProviderStats): number {
  return Math.max(1, ...all.map((s) => (s[key] as number) ?? 0));
}

function fmtMs(ms: number): string {
  if (ms < 1000) return `${ms} ms`;
  return `${(ms / 1000).toFixed(1)} s`;
}

function fmtDay(iso: string): string {
  return new Date(iso + "T00:00:00").toLocaleDateString(undefined, {
    weekday: "short",
  });
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="font-headline text-headline-md text-on-surface">Analytics</h1>
      <p class="text-sm text-on-surface-variant mt-1">
        Processing stats and provider comparison across all documents.
      </p>
    </div>

    <p v-if="error" class="text-sm text-status-error">{{ error }}</p>
    <p v-else-if="loading" class="text-sm text-on-surface-variant">Loading…</p>

    <template v-else-if="analytics">
      <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl p-6 shadow-sm">
          <p class="text-xs text-on-surface-variant uppercase tracking-widest">Documents</p>
          <p class="font-headline text-headline-lg text-on-surface mt-2">{{ totalDocs }}</p>
        </div>
        <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl p-6 shadow-sm">
          <p class="text-xs text-on-surface-variant uppercase tracking-widest">Done</p>
          <p class="font-headline text-headline-lg text-status-done mt-2">
            {{ analytics.totals.done ?? 0 }}
            <span class="text-sm text-on-surface-variant font-normal">({{ doneRate }}%)</span>
          </p>
        </div>
        <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl p-6 shadow-sm">
          <p class="text-xs text-on-surface-variant uppercase tracking-widest">Errors</p>
          <p class="font-headline text-headline-lg text-status-error mt-2">
            {{ analytics.totals.error ?? 0 }}
            <span class="text-sm text-on-surface-variant font-normal">({{ errorRate }}%)</span>
          </p>
        </div>
        <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl p-6 shadow-sm">
          <p class="text-xs text-on-surface-variant uppercase tracking-widest">Pending</p>
          <p class="font-headline text-headline-lg text-on-surface mt-2">
            {{ (analytics.totals.pending ?? 0) + (analytics.totals.processing ?? 0) }}
          </p>
        </div>
      </div>

      <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl p-6 shadow-sm">
        <h3 class="text-xs text-on-surface-variant uppercase tracking-widest mb-4">Documents per day (14 days)</h3>
        <div class="flex items-end gap-2 h-40">
          <div
            v-for="b in analytics.daily"
            :key="b.date"
            class="flex-1 flex flex-col items-center justify-end gap-1 group"
          >
            <div
              class="w-full bg-primary/20 rounded-t transition-colors group-hover:bg-primary/40 relative"
              :style="{ height: `${(b.total / maxDaily) * 100}%`, minHeight: b.total ? '4px' : '2px' }"
              :title="`${b.date}: ${b.total} docs (${b.done} done, ${b.error} errors)`"
            ></div>
            <span class="text-[10px] text-on-surface-variant">{{ fmtDay(b.date) }}</span>
          </div>
        </div>
      </div>

      <div v-if="providerList.length" class="bg-surface-container-lowest border border-surface-container-highest rounded-xl p-6 shadow-sm">
        <h3 class="text-xs text-on-surface-variant uppercase tracking-widest mb-4">Provider comparison</h3>
        <div class="flex flex-col gap-6">
          <div v-for="p in providerList" :key="p.name" class="flex flex-col gap-2">
            <div class="flex justify-between items-baseline">
              <span class="font-headline text-headline-sm text-on-surface capitalize">{{ p.name }}</span>
              <span class="text-xs text-on-surface-variant">
                {{ p.count }} docs · {{ p.error }} errors ({{ p.count ? Math.round((p.error / p.count) * 100) : 0 }}%) · {{ p.avg_retries }} avg retries
              </span>
            </div>
            <div class="flex flex-col gap-1">
              <div class="flex items-center gap-2">
                <span class="text-xs text-on-surface-variant w-24 shrink-0">Total</span>
                <div class="bg-surface-container-high rounded-full h-2.5 flex-1 overflow-hidden">
                  <div
                    class="bg-primary h-full rounded-full"
                    :style="{ width: `${(p.avg_total_ms / maxStage(providerList, 'avg_total_ms')) * 100}%` }"
                  ></div>
                </div>
                <span class="text-xs text-on-surface w-16 text-right shrink-0">{{ fmtMs(p.avg_total_ms) }}</span>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-xs text-on-surface-variant w-24 shrink-0">OCR</span>
                <div class="bg-surface-container-high rounded-full h-2.5 flex-1 overflow-hidden">
                  <div
                    class="bg-status-processing h-full rounded-full"
                    :style="{ width: `${(p.avg_ocr_ms / maxStage(providerList, 'avg_ocr_ms')) * 100}%` }"
                  ></div>
                </div>
                <span class="text-xs text-on-surface w-16 text-right shrink-0">{{ fmtMs(p.avg_ocr_ms) }}</span>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-xs text-on-surface-variant w-24 shrink-0">Classify</span>
                <div class="bg-surface-container-high rounded-full h-2.5 flex-1 overflow-hidden">
                  <div
                    class="bg-status-done h-full rounded-full"
                    :style="{ width: `${(p.avg_classify_ms / maxStage(providerList, 'avg_classify_ms')) * 100}%` }"
                  ></div>
                </div>
                <span class="text-xs text-on-surface w-16 text-right shrink-0">{{ fmtMs(p.avg_classify_ms) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <p v-else class="text-sm text-on-surface-variant">
        No processed documents yet. Upload something to see stats.
      </p>
    </template>
  </div>
</template>
