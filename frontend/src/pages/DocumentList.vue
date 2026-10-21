<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useDocuments } from "../composables/useDocuments";
import { useFolders } from "../composables/useFolders";
import { breadcrumbOf, childrenOf, descendantCount, type Folder } from "../lib/folders";
import DocumentItem from "../components/DocumentItem.vue";
import FolderCard from "../components/FolderCard.vue";

const route = useRoute();
const router = useRouter();
const { documents, loading, error, fetchAll, subscribe, upload, remove } = useDocuments();
const { folders, fetchAll: fetchFolders, subscribe: subscribeFolders, create, rename, remove: removeFolder } = useFolders();
const filter = ref<"all" | "pending" | "processing" | "done" | "error">("all");
const filterOpen = ref(false);
const uploading = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);
const newFolderName = ref("");
const showCreate = ref(false);
const createInput = ref<HTMLInputElement | null>(null);
const renamingId = ref("");
const renameValue = ref("");

let unsubDocs: (() => void) | undefined;
let unsubFolders: (() => void) | undefined;

onMounted(async () => {
  fetchAll();
  fetchFolders();
  unsubDocs = await subscribe();
  unsubFolders = await subscribeFolders();
});

onUnmounted(() => {
  unsubDocs?.();
  unsubFolders?.();
});

const currentId = computed(() => (typeof route.query.folder === "string" ? route.query.folder : ""));
const breadcrumb = computed(() => breadcrumbOf(folders.value, currentId.value));
const subfolders = computed(() => childrenOf(folders.value, currentId.value));

const visible = computed(() => {
  const inFolder = documents.value.filter((d) => (d.folder ?? "") === currentId.value);
  return filter.value === "all"
    ? inFolder
    : inFolder.filter((d) => d.status === filter.value);
});

const folderDocCount = computed(() => {
  const counts = new Map<string, number>();
  for (const d of documents.value) {
    const key = d.folder ?? "";
    counts.set(key, (counts.get(key) ?? 0) + 1);
  }
  return counts;
});

function navigate(folderId: string) {
  router.push(folderId ? { query: { folder: folderId } } : { query: {} });
}

async function onFiles(event: Event) {
  const input = event.target as HTMLInputElement;
  if (!input.files?.length) return;
  uploading.value = true;
  try {
    await upload(Array.from(input.files), currentId.value);
  } finally {
    uploading.value = false;
    input.value = "";
  }
}

async function onDelete(doc: { id: string; filename: string }) {
  if (!confirm(`Delete "${doc.filename}"?`)) return;
  await remove(doc.id);
}

function openCreate() {
  newFolderName.value = "";
  showCreate.value = true;
  nextTick(() => createInput.value?.focus());
}

function closeCreate() {
  showCreate.value = false;
  newFolderName.value = "";
}

async function addFolder() {
  const name = newFolderName.value.trim();
  if (!name) return;
  await create(name, currentId.value);
  closeCreate();
}

async function startRename(folder: Folder) {
  renameValue.value = folder.name;
  renamingId.value = folder.id;
}

async function saveRename() {
  const name = renameValue.value.trim();
  if (!name || !renamingId.value) return;
  await rename(renamingId.value, name);
  renamingId.value = "";
}

async function deleteFolder(folder: Folder) {
  const sub = descendantCount(folders.value, folder.id);
  const docs = folderDocCount.value.get(folder.id) ?? 0;
  const parts = [`Delete folder "${folder.name}"?`];
  if (docs) parts.push(`${docs} document${docs === 1 ? "" : "s"} will be deleted.`);
  if (sub) parts.push(`${sub} subfolder${sub === 1 ? "" : "s"} (and their contents) will be deleted.`);
  if (!confirm(parts.join("\n"))) return;
  await removeFolder(folder.id);
  if (folder.id === currentId.value) navigate("");
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div class="flex flex-col md:flex-row justify-between items-start md:items-center gap-4">
      <div>
        <h1 class="font-headline text-headline-md text-on-surface">Document Repository</h1>
        <p class="text-sm text-on-surface-variant mt-1">Manage and monitor AI processing queues.</p>
      </div>
    </div>

    <div class="flex flex-wrap items-center gap-2">
      <button
        class="text-xs font-medium px-3 py-1 rounded-full bg-surface border border-outline-variant text-on-surface hover:bg-surface-container"
        @click="navigate('')"
      >
        Root
      </button>
      <template v-for="(folder, i) in breadcrumb" :key="folder.id">
        <span class="text-on-surface-variant text-xs">/</span>
        <template v-if="renamingId === folder.id">
          <input
            v-model="renameValue"
            type="text"
            class="text-sm px-2 py-1 rounded border border-outline-variant bg-surface text-on-surface focus:outline-none focus:border-primary"
            @keyup.enter="saveRename"
            @keyup.esc="renamingId = ''"
          />
          <button
            class="p-1 rounded text-primary hover:bg-surface-container transition-colors"
            title="Confirm rename"
            @click="saveRename"
          >
            <span class="material-symbols-outlined text-[16px]">check</span>
          </button>
          <button
            class="p-1 rounded text-on-surface-variant hover:text-status-error hover:bg-status-error/10 transition-colors"
            title="Delete"
            @click="deleteFolder(folder)"
          >
            <span class="material-symbols-outlined text-[16px]">delete</span>
          </button>
        </template>
        <button
          v-else
          class="text-xs font-medium px-3 py-1 rounded-full group"
          :class="i === breadcrumb.length - 1
            ? 'bg-secondary-container text-on-secondary-container'
            : 'bg-surface border border-outline-variant text-on-surface hover:bg-surface-container'"
          @click="navigate(folder.id)"
        >
          {{ folder.name }}
          <span
            v-if="i === breadcrumb.length - 1"
            class="material-symbols-outlined text-[16px] align-middle opacity-0 group-hover:opacity-100 transition-opacity"
            title="Rename"
            @click.stop="startRename(folder)"
          >edit</span>
        </button>
      </template>
    </div>

    <div v-if="showCreate" class="fixed inset-0 z-20 flex items-center justify-center">
      <div class="fixed inset-0 bg-black/40" @click="closeCreate"></div>
      <div class="relative bg-surface border border-outline-variant rounded-xl shadow-lg p-5 w-full max-w-sm">
        <h2 class="font-headline text-headline-sm text-on-surface mb-4">Create Folder</h2>
        <input
          ref="createInput"
          v-model="newFolderName"
          type="text"
          placeholder="Folder name..."
          class="w-full text-sm px-3 py-1.5 rounded border border-outline-variant bg-surface text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:border-primary"
          @keyup.enter="addFolder"
          @keyup.esc="closeCreate"
        />
        <div class="flex justify-end gap-2 mt-4">
          <button
            class="text-xs font-medium px-3 py-1.5 rounded border border-outline-variant text-on-surface hover:bg-surface-container-high transition-colors"
            @click="closeCreate"
          >
            Cancel
          </button>
          <button
            class="text-xs font-medium px-3 py-1.5 rounded bg-primary text-on-primary hover:opacity-90 transition-opacity disabled:opacity-50"
            :disabled="!newFolderName.trim()"
            @click="addFolder"
          >
            Create
          </button>
        </div>
      </div>
    </div>

    <div class="relative flex items-center justify-end gap-2 border-b border-surface-container-highest pb-4">
      <button
        class="text-xs font-medium px-3 py-1.5 rounded-lg border border-outline-variant text-on-surface hover:bg-surface-container flex items-center gap-1.5 transition-colors"
        :class="filter !== 'all' ? 'bg-secondary-container text-on-secondary-container border-transparent' : ''"
        @click="filterOpen = !filterOpen"
      >
        <span class="material-symbols-outlined text-[16px]">filter_list</span>
        Filters{{ filter !== "all" ? `: ${filter}` : "" }}
      </button>
      <button
        class="text-xs font-medium px-3 py-1.5 rounded-lg border border-outline-variant text-on-surface hover:bg-surface-container flex items-center gap-1.5 transition-colors"
        @click="openCreate"
      >
        <span class="material-symbols-outlined text-[16px]">create_new_folder</span>
        New Folder
      </button>
      <button
        class="text-xs font-medium px-3 py-1.5 rounded-lg bg-primary text-on-primary flex items-center gap-1.5 transition-opacity hover:opacity-90 disabled:opacity-50"
        :disabled="uploading"
        @click="fileInput?.click()"
      >
        <span class="material-symbols-outlined text-[16px]">upload</span>
        {{ uploading ? "Uploading..." : "Upload Document" }}
      </button>
      <input
        ref="fileInput"
        type="file"
        multiple
        class="hidden"
        accept=".pdf,.docx,.odt,.md,.txt,.xlsx,.png,.jpg,.jpeg"
        @change="onFiles"
      />
      <div
        v-if="filterOpen"
        class="absolute top-full right-0 mt-2 z-10 bg-surface border border-outline-variant rounded-lg shadow-lg py-1 min-w-36"
      >
        <button
          v-for="f in (['all', 'pending', 'processing', 'done', 'error'] as const)"
          :key="f"
          class="w-full text-left text-xs font-medium px-3 py-1.5 capitalize hover:bg-surface-container transition-colors"
          :class="filter === f ? 'bg-secondary-container text-on-secondary-container' : 'text-on-surface'"
          @click="filter = f; filterOpen = false"
        >
          {{ f }}
        </button>
      </div>
      <div v-if="filterOpen" class="fixed inset-0 z-[5]" @click="filterOpen = false"></div>
    </div>

    <div v-if="subfolders.length" class="grid grid-cols-[repeat(auto-fill,minmax(220px,1fr))] gap-3">
      <FolderCard
        v-for="folder in subfolders"
        :key="folder.id"
        :folder="folder"
        :file-count="folderDocCount.get(folder.id) ?? 0"
        @open="navigate(folder.id)"
        @rename="(name) => rename(folder.id, name)"
        @delete="deleteFolder(folder)"
      />
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
        <p v-if="loading" class="p-6 text-sm text-on-surface-variant">Loading...</p>
        <p v-else-if="error" class="p-6 text-sm text-status-error">{{ error }}</p>
        <p v-else-if="visible.length === 0" class="p-6 text-sm text-on-surface-variant">
          No documents here. Upload one to get started.
        </p>
        <DocumentItem v-for="doc in visible" :key="doc.id" :document="doc" @delete="onDelete(doc)" />
      </div>
    </div>
  </div>
</template>
