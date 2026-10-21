<script setup lang="ts">
import { nextTick, ref } from "vue";
import type { Folder } from "../lib/folders";

const props = defineProps<{
  folder: Folder;
  fileCount: number;
}>();

const emit = defineEmits<{
  open: [];
  rename: [name: string];
  delete: [];
}>();

const renaming = ref(false);
const renameValue = ref("");
const renameInput = ref<HTMLInputElement | null>(null);

function startRename() {
  renameValue.value = props.folder.name;
  renaming.value = true;
  nextTick(() => renameInput.value?.focus());
}

function saveRename() {
  const name = renameValue.value.trim();
  if (!name) return;
  emit("rename", name);
  renaming.value = false;
}

function updatedLabel(iso: string): string {
  const time = new Date(iso).getTime();
  if (!iso || Number.isNaN(time)) return "Updated recently";
  const days = Math.round((Date.now() - time) / 86400000);
  if (days <= 0) return "Updated today";
  if (days === 1) return "Updated yesterday";
  if (days < 7) return `Updated ${days} days ago`;
  return `Updated ${new Date(iso).toLocaleDateString()}`;
}
</script>

<template>
  <div
    class="bg-surface-container-lowest border border-surface-container-highest rounded-xl p-4 shadow-sm hover:shadow-md hover:bg-surface-container-low transition-all cursor-pointer"
    @click="emit('open')"
  >
    <div class="flex items-start justify-between">
      <div class="w-10 h-10 rounded-lg bg-primary-container text-on-primary-container flex items-center justify-center">
        <span class="material-symbols-outlined text-[20px]">folder</span>
      </div>
      <div class="flex gap-0.5" @click.stop>
        <button
          v-if="renaming"
          class="p-1 rounded text-primary hover:bg-surface-container transition-colors"
          title="Confirm rename"
          @click="saveRename"
        >
          <span class="material-symbols-outlined text-[16px]">check</span>
        </button>
        <button
          v-else
          class="p-1 rounded text-on-surface-variant hover:text-on-surface hover:bg-surface-container transition-colors"
          title="Rename"
          @click="startRename"
        >
          <span class="material-symbols-outlined text-[16px]">edit</span>
        </button>
        <button
          class="p-1 rounded text-on-surface-variant hover:text-status-error hover:bg-status-error/10 transition-colors"
          title="Delete"
          @click="emit('delete')"
        >
          <span class="material-symbols-outlined text-[16px]">delete</span>
        </button>
      </div>
    </div>
    <div class="mt-3">
      <input
        v-if="renaming"
        ref="renameInput"
        v-model="renameValue"
        type="text"
        class="w-full text-sm px-2 py-1 rounded border border-outline-variant bg-surface text-on-surface focus:outline-none focus:border-primary"
        @keyup.enter="saveRename"
        @keyup.esc="renaming = false"
        @click.stop
      />
      <p v-else class="text-sm font-semibold text-on-surface truncate">{{ folder.name }}</p>
    </div>
    <div class="mt-3 flex items-center justify-between gap-2">
      <span class="text-xs text-on-surface-variant bg-surface-container rounded px-2 py-0.5">
        {{ fileCount === 1 ? "1 file" : `${fileCount} files` }}
      </span>
      <span class="text-xs text-on-surface-variant">{{ updatedLabel(folder.updated) }}</span>
    </div>
  </div>
</template>
