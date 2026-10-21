<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { marked } from "marked";
import { useDocument } from "../composables/useDocument";
import { progressLabel } from "../composables/useDocuments";
import { useFolders } from "../composables/useFolders";
import { breadcrumbOf } from "../lib/folders";
import { pb } from "../lib/pb";
import StatusBadge from "../components/StatusBadge.vue";

const route = useRoute();
const router = useRouter();
const id = route.params.id as string;
const {
  document,
  loading,
  error,
  similar,
  similarLoading,
  reembedding,
  fetch,
  subscribe,
  update,
  findSimilar,
  reembed,
} = useDocument(id);
const { folders, fetchAll: fetchFolders, subscribe: subscribeFolders } = useFolders();
const newTag = ref("");
const activeTab = ref<"text" | "original">("text");
const originalText = ref("");
const originalTextLoading = ref(false);
const menuOpen = ref(false);
const reembedMsg = ref("");
const reembedError = ref("");
let reembedMsgTimer: ReturnType<typeof setTimeout> | undefined;

let unsub: (() => void) | undefined;
let unsubFolders: (() => void) | undefined;

const embedding = computed(
  () => document.value?.embed_status === "pending" || document.value?.embed_status === "processing",
);

onMounted(async () => {
  fetch();
  fetchFolders();
  unsub = await subscribe();
  unsubFolders = await subscribeFolders();
});

onUnmounted(() => {
  unsub?.();
  unsubFolders?.();
  clearTimeout(reembedMsgTimer);
});

const folderOptions = computed(() =>
  folders.value
    .map((f) => ({
      id: f.id,
      label: breadcrumbOf(folders.value, f.id)
        .map((b) => b.name)
        .join(" / "),
    }))
    .sort((a, b) => a.label.localeCompare(b.label)),
);

async function moveToFolder(folderId: string) {
  await update({ folder: folderId });
}

function downloadUrl(): string {
  if (!document.value?.file) return "";
  return pb.files.getURL(document.value, document.value.file);
}

function isImage(): boolean {
  return document.value?.mime.startsWith("image/") ?? false;
}

function isPdf(): boolean {
  return document.value?.mime === "application/pdf";
}

function isText(): boolean {
  if (!document.value) return false;
  return (
    document.value.mime.startsWith("text/") ||
    /\.(md|txt)$/i.test(document.value.filename)
  );
}

function isMarkdown(): boolean {
  if (!document.value) return false;
  return (
    document.value.mime === "text/markdown" ||
    /\.md$/i.test(document.value.filename)
  );
}

function renderMarkdown(text: string): string {
  return marked.parse(text, { async: false }) as string;
}

async function loadOriginalText() {
  if (!document.value || !isText()) return;
  originalTextLoading.value = true;
  try {
    const res = await globalThis.fetch(downloadUrl());
    originalText.value = await res.text();
  } catch {
    originalText.value = "";
  } finally {
    originalTextLoading.value = false;
  }
}

function selectTab(tab: "text" | "original") {
  activeTab.value = tab;
  if (tab === "original" && isText() && !originalText.value) {
    loadOriginalText();
  }
}

async function remove() {
  if (!document.value) return;
  if (!confirm(`Delete "${document.value.filename}"?`)) return;
  await pb.collection("documents").delete(id);
  router.push("/");
}

async function reprocess() {
  await update({ status: "pending", error: "", retry_count: 0 });
}

async function onReembed() {
  menuOpen.value = false;
  reembedError.value = "";
  reembedMsg.value = "";
  try {
    await reembed();
    reembedMsg.value = "Embedding queued";
    clearTimeout(reembedMsgTimer);
    reembedMsgTimer = setTimeout(() => (reembedMsg.value = ""), 3000);
  } catch (e) {
    reembedError.value = String(e);
  }
}

async function addTag() {
  const tag = newTag.value.trim();
  if (!tag || !document.value) return;
  const tags = document.value.tags ?? [];
  if (tags.includes(tag)) {
    newTag.value = "";
    return;
  }
  await update({ tags: [...tags, tag] });
  newTag.value = "";
}

async function removeTag(tag: string) {
  if (!document.value) return;
  await update({ tags: (document.value.tags ?? []).filter((t) => t !== tag) });
}
</script>

<template>
  <div v-if="loading" class="text-sm text-on-surface-variant">Loading...</div>
  <div v-else-if="error" class="text-sm text-status-error">{{ error }}</div>
  <div v-else-if="document" class="flex flex-col gap-6">
    <div class="flex items-center justify-between gap-4">
      <div class="flex items-center gap-3 min-w-0">
        <RouterLink
          :to="document.folder ? { path: '/', query: { folder: document.folder } } : '/'"
          class="text-on-surface-variant hover:text-primary"
          >←</RouterLink
        >
        <h1 class="font-headline text-headline-sm text-on-surface truncate">{{ document.filename }}</h1>
        <StatusBadge :status="document.status" />
      </div>
      <div class="flex items-center gap-2">
        <button
          v-if="document.status === 'done'"
          class="text-xs font-medium px-3 py-1.5 rounded border border-outline-variant text-on-surface hover:bg-surface-container-high transition-colors disabled:opacity-50"
          :disabled="similarLoading"
          @click="findSimilar"
        >
          {{ similarLoading ? "Searching..." : "Find Similar" }}
        </button>
        <div
          v-if="document.status === 'done' || document.status === 'error'"
          class="relative flex"
        >
          <button
            class="text-xs font-medium px-3 py-1.5 border border-outline-variant text-on-surface hover:bg-surface-container-high transition-colors"
            :class="document.status === 'done' ? 'rounded-l' : 'rounded'"
            @click="reprocess"
          >
            Re-process
          </button>
          <button
            v-if="document.status === 'done'"
            class="text-xs font-medium px-1.5 py-1.5 rounded-r border border-l-0 border-outline-variant text-on-surface hover:bg-surface-container-high transition-colors disabled:opacity-50"
            :disabled="reembedding"
            title="More re-processing options"
            @click="menuOpen = !menuOpen"
          >
            <span class="material-symbols-outlined text-[16px] block leading-none">expand_more</span>
          </button>
          <div v-if="menuOpen" class="fixed inset-0 z-10" @click="menuOpen = false"></div>
          <div
            v-if="menuOpen"
            class="absolute right-0 top-full mt-1 z-20 bg-surface border border-outline-variant rounded shadow-lg py-1 min-w-[140px]"
          >
            <button
              class="w-full text-left text-xs px-3 py-1.5 text-on-surface hover:bg-surface-container-high transition-colors disabled:opacity-50"
              :disabled="reembedding || embedding"
              @click="onReembed"
            >
              {{ embedding ? "Embedding..." : reembedding ? "Queueing..." : "Re-embed" }}
            </button>
          </div>
        </div>
        <span v-if="reembedMsg" class="text-xs text-status-done">{{ reembedMsg }}</span>
        <span v-if="reembedError" class="text-xs text-status-error">{{ reembedError }}</span>
        <a
          :href="downloadUrl()"
          class="text-xs font-medium px-3 py-1.5 rounded border border-outline-variant text-on-surface hover:bg-surface-container-high transition-colors"
        >
          Download Original
        </a>
        <button
          class="text-xs font-medium px-3 py-1.5 rounded border border-status-error text-status-error hover:bg-status-error/10 transition-colors"
          @click="remove"
        >
          Delete
        </button>
      </div>
    </div>

    <div class="flex flex-col lg:flex-row gap-6">
      <section class="flex-[3] bg-surface-container-lowest border border-surface-container-highest rounded-xl overflow-hidden">
        <div class="h-12 border-b border-surface-container-highest flex items-center px-4 bg-surface-bright gap-6">
          <button
            class="text-xs font-medium h-12 flex items-center transition-colors"
            :class="activeTab === 'text' ? 'text-primary border-b-2 border-primary' : 'text-on-surface-variant hover:text-on-surface'"
            @click="selectTab('text')"
          >
            Extracted Text
          </button>
          <button
            class="text-xs font-medium h-12 flex items-center transition-colors"
            :class="activeTab === 'original' ? 'text-primary border-b-2 border-primary' : 'text-on-surface-variant hover:text-on-surface'"
            @click="selectTab('original')"
          >
            Preview
          </button>
        </div>
        <div v-if="activeTab === 'text'" class="p-6 md:p-8 overflow-y-auto max-h-[70vh]">
          <pre
            v-if="document.ocr_text"
            class="whitespace-pre-wrap font-body text-sm text-slate-700 leading-relaxed"
          >{{ document.ocr_text }}</pre>
          <p v-else class="text-sm text-on-surface-variant italic">
            {{ document.status === "error" ? "No text extracted." : document.status === "processing" ? progressLabel(document.progress) : "Waiting for processing..." }}
          </p>
        </div>
        <div v-else class="p-6 md:p-8 overflow-y-auto max-h-[70vh]">
          <img
            v-if="isImage()"
            :src="downloadUrl()"
            :alt="document.filename"
            class="max-w-full h-auto rounded border border-surface-container-highest"
          />
          <iframe
            v-else-if="isPdf()"
            :src="downloadUrl()"
            class="w-full h-[70vh] rounded border border-surface-container-highest"
          ></iframe>
          <div v-else-if="isText()" class="flex flex-col gap-4">
            <p v-if="originalTextLoading" class="text-sm text-on-surface-variant italic">Loading...</p>
            <div
              v-else-if="originalText && isMarkdown()"
              class="markdown-body font-body text-sm text-slate-700 leading-relaxed"
              v-html="renderMarkdown(originalText)"
            ></div>
            <pre
              v-else-if="originalText"
              class="whitespace-pre-wrap font-body text-sm text-slate-700 leading-relaxed"
            >{{ originalText }}</pre>
            <p v-else class="text-sm text-on-surface-variant italic">Could not load the original file.</p>
          </div>
          <div v-else class="flex flex-col items-start gap-4">
            <p class="text-sm text-on-surface-variant italic">
              Preview not available for this file type.
            </p>
            <a
              :href="downloadUrl()"
              class="text-xs font-medium px-3 py-1.5 rounded border border-outline-variant text-on-surface hover:bg-surface-container-high transition-colors"
            >
              Download Original
            </a>
          </div>
        </div>
      </section>

      <aside class="flex-[2] bg-surface-bright border border-surface-container-highest rounded-xl overflow-hidden">
        <div class="h-12 border-b border-surface-container-highest flex items-center px-4 bg-surface">
          <span class="text-xs font-medium text-on-surface">AI Extracted Metadata</span>
        </div>
        <div class="p-4 space-y-6">
          <div class="bg-surface-container-low border border-surface-container-highest rounded p-4">
            <h4 class="text-xs text-on-surface-variant uppercase tracking-wider mb-2">Summary</h4>
            <p class="text-sm text-on-surface leading-snug">{{ document.summary || "--" }}</p>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1">
              <span class="text-xs text-on-surface-variant uppercase">Category</span>
              <span class="text-sm font-medium text-on-surface">{{ document.category || "--" }}</span>
            </div>
            <div class="flex flex-col gap-1">
              <span class="text-xs text-on-surface-variant uppercase">Language</span>
              <span class="text-sm font-medium text-on-surface">{{ document.language || "--" }}</span>
            </div>
          </div>
          <div class="flex flex-col gap-1">
            <label for="doc-folder" class="text-xs text-on-surface-variant uppercase">Folder</label>
            <select
              id="doc-folder"
              :value="document.folder ?? ''"
              class="text-sm px-3 py-1.5 rounded border border-outline-variant bg-surface text-on-surface focus:outline-none focus:border-primary"
              @change="moveToFolder(($event.target as HTMLSelectElement).value)"
            >
              <option value="">Root</option>
              <option v-for="folder in folderOptions" :key="folder.id" :value="folder.id">
                {{ folder.label }}
              </option>
            </select>
          </div>
          <div class="w-full h-px bg-surface-container-highest"></div>
          <div>
            <h4 class="text-xs text-on-surface-variant uppercase tracking-wider mb-3">Extracted Entities</h4>
            <ul v-if="document.entities?.length" class="space-y-2">
              <li v-for="entity in document.entities" :key="entity" class="text-sm text-on-surface">
                {{ entity }}
              </li>
            </ul>
            <p v-else class="text-sm text-on-surface-variant">--</p>
          </div>
          <div class="w-full h-px bg-surface-container-highest"></div>
          <div>
            <h4 class="text-xs text-on-surface-variant uppercase tracking-wider mb-3">AI Tags</h4>
            <div v-if="document.tags?.length" class="flex flex-wrap gap-2 mb-3">
              <span
                v-for="tag in document.tags"
                :key="tag"
                class="px-2.5 py-1 bg-surface-container-high border border-surface-container-highest rounded text-on-surface text-xs flex items-center gap-1.5"
              >
                {{ tag }}
                <button
                  class="text-on-surface-variant hover:text-status-error leading-none"
                  title="Remove tag"
                  @click="removeTag(tag)"
                >
                  ×
                </button>
              </span>
            </div>
            <p v-else class="text-sm text-on-surface-variant mb-3">--</p>
            <div class="flex gap-2">
              <input
                v-model="newTag"
                type="text"
                placeholder="Add tag..."
                class="flex-1 min-w-0 text-sm px-3 py-1.5 rounded border border-outline-variant bg-surface text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:border-primary"
                @keyup.enter="addTag"
              />
              <button
                class="text-xs font-medium px-3 py-1.5 rounded bg-primary text-on-primary hover:opacity-90 transition-opacity disabled:opacity-50"
                :disabled="!newTag.trim()"
                @click="addTag"
              >
                Add
              </button>
            </div>
          </div>
          <div v-if="similar.length" class="bg-surface-container-low border border-surface-container-highest rounded">
            <h4 class="text-xs text-on-surface-variant uppercase tracking-wider px-3 py-2 border-b border-surface-container-highest">
              Similar Documents
            </h4>
            <RouterLink
              v-for="hit in similar"
              :key="hit.item.id"
              :to="`/documents/${hit.item.id}`"
              class="flex items-center gap-2 px-3 py-2 border-b last:border-b-0 border-surface-container-highest hover:bg-surface-container-high transition-colors"
            >
              <span class="text-sm text-on-surface truncate flex-1">{{ hit.item.filename }}</span>
              <span class="text-xs text-on-surface-variant tabular-nums">{{ (hit.score * 100).toFixed(0) }}%</span>
            </RouterLink>
          </div>
          <div v-if="document.model" class="p-3 bg-primary-fixed-dim/20 rounded border border-primary-fixed">
            <span class="text-xs text-on-surface-variant">Processed locally with {{ document.model }}.</span>
          </div>
          <div
            v-if="document.embed_status"
            class="p-3 rounded border"
            :class="document.embed_status === 'error'
              ? 'bg-status-error/10 border-status-error/30'
              : 'bg-surface-container-low border-surface-container-highest'"
          >
            <span class="text-xs text-on-surface-variant">
              Embedding:
              <template v-if="document.embed_status === 'done'">indexed for semantic search</template>
              <template v-else-if="document.embed_status === 'processing'">indexing...</template>
              <template v-else-if="document.embed_status === 'pending'">
                {{ document.embed_error || "queued" }}
              </template>
              <template v-else>{{ document.embed_error || "failed" }}</template>
            </span>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<style scoped>
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4) {
  font-family: var(--font-headline, inherit);
  font-weight: 600;
  color: var(--color-on-surface, #1e293b);
  margin: 1.25em 0 0.5em;
  line-height: 1.3;
}
.markdown-body :deep(h1) { font-size: 1.5em; }
.markdown-body :deep(h2) { font-size: 1.25em; }
.markdown-body :deep(h3) { font-size: 1.1em; }
.markdown-body :deep(h4) { font-size: 1em; }
.markdown-body :deep(p) { margin: 0.5em 0; }
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin: 0.5em 0;
  padding-left: 1.5em;
}
.markdown-body :deep(ul) { list-style: disc; }
.markdown-body :deep(ol) { list-style: decimal; }
.markdown-body :deep(li) { margin: 0.25em 0; }
.markdown-body :deep(a) {
  color: var(--color-primary, #3b82f6);
  text-decoration: underline;
}
.markdown-body :deep(code) {
  font-family: ui-monospace, monospace;
  font-size: 0.9em;
  background: var(--color-surface-container-high, #e2e8f0);
  padding: 0.1em 0.35em;
  border-radius: 4px;
}
.markdown-body :deep(pre) {
  background: var(--color-surface-container-high, #e2e8f0);
  padding: 0.75em 1em;
  border-radius: 8px;
  overflow-x: auto;
  margin: 0.75em 0;
}
.markdown-body :deep(pre code) {
  background: none;
  padding: 0;
}
.markdown-body :deep(blockquote) {
  border-left: 3px solid var(--color-outline-variant, #cbd5e1);
  padding-left: 1em;
  margin: 0.75em 0;
  color: var(--color-on-surface-variant, #64748b);
}
.markdown-body :deep(hr) {
  border: none;
  border-top: 1px solid var(--color-surface-container-highest, #cbd5e1);
  margin: 1em 0;
}
.markdown-body :deep(img) { max-width: 100%; }
.markdown-body :deep(table) {
  border-collapse: collapse;
  margin: 0.75em 0;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid var(--color-surface-container-highest, #cbd5e1);
  padding: 0.4em 0.75em;
}
</style>
