<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { useSettings } from "../composables/useSettings";
import ProviderSettingsCard from "../components/ProviderSettingsCard.vue";

const {
  settings,
  loading,
  error,
  saved,
  health,
  checking,
  backfilling,
  backfillResult,
  fetch,
  save,
  checkHealth,
  backfill,
} = useSettings();

const form = reactive({
  backend: "llamacpp",
  ollama: { url: "", vision_model: "", classify_model: "", embed_model: "", embed_url: "" },
  llamacpp: { url: "", vision_model: "", classify_model: "", embed_model: "", embed_url: "" },
  concurrency: 1,
});

const saving = ref(false);
const saveError = ref("");
const backfillError = ref("");

onMounted(async () => {
  await fetch();
  if (settings.value) Object.assign(form, settings.value);
});

async function onSave() {
  saving.value = true;
  saveError.value = "";
  try {
    await save({ ...form, concurrency: Number(form.concurrency) });
  } catch (e) {
    saveError.value = String(e);
  } finally {
    saving.value = false;
  }
}

async function onBackfill() {
  backfillError.value = "";
  try {
    await backfill();
  } catch (e) {
    backfillError.value = String(e);
  }
}
</script>

<template>
  <div class="flex flex-col gap-6 max-w-2xl">
    <div>
      <h1 class="font-headline text-headline-md text-on-surface">Settings</h1>
      <p class="text-sm text-on-surface-variant mt-1">
        Configure the AI providers used for document processing.
      </p>
    </div>

    <div v-if="loading" class="text-sm text-on-surface-variant">Loading...</div>
    <div v-else-if="error && !settings" class="text-sm text-status-error">{{ error }}</div>

    <form
      v-else
      class="flex flex-col gap-6"
      @submit.prevent="onSave"
    >
      <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl shadow-sm overflow-hidden">
        <div class="px-6 py-4 border-b border-surface-container-highest bg-surface-container-low">
          <h3 class="font-headline text-headline-sm text-on-surface">General</h3>
        </div>
        <div class="p-6 flex flex-col gap-5">
          <div class="flex flex-col gap-1.5">
            <label for="backend" class="text-xs font-medium text-on-surface-variant uppercase tracking-wider">
              Active Backend
            </label>
            <select
              id="backend"
              v-model="form.backend"
              class="text-sm px-3 py-2 rounded border border-outline-variant bg-surface text-on-surface focus:outline-none focus:border-primary"
            >
              <option value="llamacpp">llama.cpp (llama-server)</option>
              <option value="ollama">Ollama</option>
            </select>
          </div>

          <div class="flex flex-col gap-1.5">
            <label for="concurrency" class="text-xs font-medium text-on-surface-variant uppercase tracking-wider">
              Concurrency
            </label>
            <input
              id="concurrency"
              v-model.number="form.concurrency"
              type="number"
              min="1"
              max="16"
              class="text-sm px-3 py-2 rounded border border-outline-variant bg-surface text-on-surface focus:outline-none focus:border-primary w-32"
            />
            <p class="text-xs text-on-surface-variant">Documents processed in parallel (1–16).</p>
          </div>
        </div>
      </div>

      <ProviderSettingsCard
        v-if="form.backend === 'llamacpp'"
        id="llamacpp"
        title="llama.cpp"
        icon="memory"
        :active="true"
        :model="form.llamacpp"
        @update:model="form.llamacpp = $event"
      />
      <ProviderSettingsCard
        v-if="form.backend === 'ollama'"
        id="ollama"
        title="Ollama"
        icon="smart_toy"
        :active="true"
        :model="form.ollama"
        @update:model="form.ollama = $event"
      />

      <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl shadow-sm px-6 py-4 flex items-center gap-3">
        <button
          type="submit"
          class="text-xs font-medium px-4 py-2 rounded bg-primary text-on-primary hover:opacity-90 transition-opacity disabled:opacity-50"
          :disabled="saving"
        >
          {{ saving ? "Saving..." : "Save Settings" }}
        </button>
        <span v-if="saved" class="text-xs text-status-done flex items-center gap-1">
          <span class="material-symbols-outlined text-[16px]">check_circle</span>
          Saved
        </span>
        <span v-if="saveError" class="text-xs text-status-error">{{ saveError }}</span>
      </div>

      <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl shadow-sm px-6 py-4 flex items-center gap-3">
        <button
          type="button"
          class="text-xs font-medium px-4 py-2 rounded border border-outline-variant text-on-surface hover:bg-surface-container-low transition-colors disabled:opacity-50"
          :disabled="checking"
          @click="checkHealth"
        >
          {{ checking ? "Testing..." : "Test Connection" }}
        </button>
        <span
          v-if="health"
          class="text-xs flex items-center gap-1"
          :class="health.online ? 'text-status-done' : 'text-status-error'"
        >
          <span class="material-symbols-outlined text-[16px]">
            {{ health.online ? "check_circle" : "error" }}
          </span>
          {{ health.online ? "Provider online" : health.error || "Provider offline" }}
        </span>
      </div>

      <div class="bg-surface-container-lowest border border-surface-container-highest rounded-xl shadow-sm px-6 py-4 flex flex-wrap items-center gap-3">
        <button
          type="button"
          class="text-xs font-medium px-4 py-2 rounded border border-outline-variant text-on-surface hover:bg-surface-container-low transition-colors disabled:opacity-50"
          :disabled="backfilling"
          @click="onBackfill"
        >
          {{ backfilling ? "Queueing..." : "Embed All Documents" }}
        </button>
        <span class="text-xs text-on-surface-variant">
          Queue every processed document for semantic-search indexing (save settings first). Runs in the background; progress on the Activity page.
        </span>
        <span v-if="backfillResult" class="text-xs text-status-done">
          {{ backfillResult.queued }} queued, {{ backfillResult.skipped }} skipped
        </span>
        <span v-if="backfillError" class="text-xs text-status-error">{{ backfillError }}</span>
      </div>
    </form>
  </div>
</template>
