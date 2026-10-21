<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useDocuments } from "../composables/useDocuments";
import { useFolders } from "../composables/useFolders";
import { childrenOf } from "../lib/folders";

interface UploadItem {
  id: number;
  name: string;
  size: number;
  status: "uploading" | "done" | "error";
  error?: string;
  docId?: string;
}

const { upload } = useDocuments();
const { folders, fetchAll: fetchFolders, subscribe: subscribeFolders } = useFolders();
const items = ref<UploadItem[]>([]);
const dragging = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);
const targetFolder = ref("");
let nextId = 1;

let unsubFolders: (() => void) | undefined;

onMounted(() => {
  fetchFolders();
  unsubFolders = subscribeFolders();
});

onUnmounted(() => unsubFolders?.());

const rootFolders = computed(() => childrenOf(folders.value, ""));

const activeCount = computed(() => items.value.filter((i) => i.status === "uploading").length);
const doneCount = computed(() => items.value.filter((i) => i.status === "done").length);

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

async function addFiles(files: File[]) {
  if (!files.length) return;
  const batch = files.map((f) => ({
    id: nextId++,
    name: f.name,
    size: f.size,
    status: "uploading" as const,
  }));
  items.value.unshift(...batch);
  for (let i = 0; i < batch.length; i++) {
    const item = batch[i];
    try {
      const records = await upload([files[i]], targetFolder.value);
      item.status = "done";
      item.docId = records[0]?.id;
    } catch (e) {
      item.status = "error";
      item.error = String(e);
    }
  }
}

function onBrowse() {
  fileInput.value?.click();
}

function onInput(event: Event) {
  const input = event.target as HTMLInputElement;
  if (input.files?.length) addFiles(Array.from(input.files));
  input.value = "";
}

function onDrop(event: DragEvent) {
  dragging.value = false;
  if (event.dataTransfer?.files.length) addFiles(Array.from(event.dataTransfer.files));
}

function removeItem(id: number) {
  items.value = items.value.filter((i) => i.id !== id);
}

function clearCompleted() {
  items.value = items.value.filter((i) => i.status !== "done");
}
</script>

<template>
  <div class="flex flex-col gap-6">
    <div>
      <h1 class="font-headline text-headline-md text-on-surface">Upload Documents</h1>
      <p class="text-sm text-on-surface-variant mt-1">
        Securely add files to your local repository for AI processing.
      </p>
    </div>

    <div class="flex items-center gap-3">
      <label for="upload-folder" class="text-xs font-medium text-on-surface-variant uppercase tracking-wider">
        Destination Folder
      </label>
      <select
        id="upload-folder"
        v-model="targetFolder"
        class="text-sm px-3 py-2 rounded border border-outline-variant bg-surface text-on-surface focus:outline-none focus:border-primary"
      >
        <option value="">Root</option>
        <option v-for="folder in rootFolders" :key="folder.id" :value="folder.id">
          {{ folder.name }}
        </option>
      </select>
    </div>

    <div
      class="border-2 border-dashed rounded-xl bg-surface-container-lowest flex flex-col items-center justify-center py-16 px-6 cursor-pointer transition-colors duration-200"
      :class="dragging ? 'border-primary bg-surface-container' : 'border-outline-variant hover:bg-surface-container'"
      @click="onBrowse"
      @dragover.prevent="dragging = true"
      @dragleave.prevent="dragging = false"
      @drop.prevent="onDrop"
    >
      <div class="bg-primary-container/10 p-4 rounded-full mb-4">
        <span class="material-symbols-outlined text-4xl text-primary">upload_file</span>
      </div>
      <h3 class="font-headline text-headline-sm text-on-surface mb-2">Drag &amp; Drop files here</h3>
      <p class="text-sm text-on-surface-variant mb-6 text-center max-w-md">
        Or click to browse your local file system. Documents remain securely on this machine.
      </p>
      <button
        class="bg-primary text-on-primary text-xs font-medium px-6 py-2.5 rounded-lg hover:opacity-90 transition-opacity shadow-sm"
        @click.stop="onBrowse"
      >
        Browse Files
      </button>
      <div class="mt-8 flex gap-4 text-xs text-on-surface-variant">
        <span class="flex items-center gap-1"><span class="material-symbols-outlined text-[16px]">picture_as_pdf</span> PDF</span>
        <span class="flex items-center gap-1"><span class="material-symbols-outlined text-[16px]">description</span> DOCX</span>
        <span class="flex items-center gap-1"><span class="material-symbols-outlined text-[16px]">code</span> MD</span>
        <span class="flex items-center gap-1"><span class="material-symbols-outlined text-[16px]">image</span> Images</span>
      </div>
      <input
        ref="fileInput"
        type="file"
        multiple
        class="hidden"
        accept=".pdf,.docx,.odt,.md,.txt,.xlsx,.png,.jpg,.jpeg"
        @change="onInput"
      />
    </div>

    <div v-if="items.length">
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-headline text-headline-sm text-on-surface">
          Upload Queue ({{ activeCount }})
        </h3>
        <button
          v-if="doneCount"
          class="text-xs font-medium text-primary hover:underline"
          @click="clearCompleted"
        >
          Clear Completed
        </button>
      </div>
      <div class="flex flex-col gap-3">
        <div
          v-for="item in items"
          :key="item.id"
          class="bg-surface-container-lowest border rounded-lg p-list-item-py px-list-item-px flex items-center justify-between shadow-sm"
          :class="item.status === 'uploading' ? 'border-primary/20' : 'border-outline-variant'"
        >
          <div class="flex items-center gap-4 flex-1 min-w-0">
            <span
              class="material-symbols-outlined"
              :class="{
                'text-primary animate-pulse': item.status === 'uploading',
                'text-status-done': item.status === 'done',
                'text-status-error': item.status === 'error',
              }"
            >
              {{ item.status === "uploading" ? "cloud_upload" : item.status === "done" ? "check_circle" : "error" }}
            </span>
            <div class="flex-1 min-w-0">
              <div class="flex justify-between items-center gap-4 mb-1">
                <span class="text-sm font-semibold text-on-surface truncate">{{ item.name }}</span>
                <span class="text-xs text-on-surface-variant whitespace-nowrap">
                  <template v-if="item.status === 'uploading'">Uploading…</template>
                  <template v-else-if="item.status === 'done'">
                    <RouterLink :to="`/documents/${item.docId}`" class="text-primary hover:underline">
                      Uploaded — processing in background
                    </RouterLink>
                  </template>
                  <template v-else>Failed</template>
                </span>
              </div>
              <div class="w-full bg-surface-container-highest rounded-full h-1.5">
                <div
                  class="h-1.5 rounded-full transition-all duration-300"
                  :class="{
                    'bg-primary w-1/2 animate-pulse': item.status === 'uploading',
                    'bg-status-done w-full': item.status === 'done',
                    'bg-status-error w-full': item.status === 'error',
                  }"
                ></div>
              </div>
              <p v-if="item.error" class="text-xs text-status-error mt-1 truncate">{{ item.error }}</p>
            </div>
          </div>
          <button
            v-if="item.status !== 'uploading'"
            class="ml-4 text-on-surface-variant hover:text-status-error transition-colors"
            title="Remove"
            @click="removeItem(item.id)"
          >
            <span class="material-symbols-outlined text-[20px]">close</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
