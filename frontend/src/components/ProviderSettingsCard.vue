<script setup lang="ts">
import { computed } from "vue";
import type { ProviderSettings } from "../composables/useSettings";

const props = defineProps<{
  id: string;
  title: string;
  icon: string;
  active: boolean;
  model: ProviderSettings;
}>();

const emit = defineEmits<{
  "update:model": [value: ProviderSettings];
}>();

const cardClasses = computed(() =>
  props.active
    ? "border-primary/40 bg-surface-container-low"
    : "border-surface-container-highest bg-surface-container-lowest",
);

function set<K extends keyof ProviderSettings>(key: K, value: ProviderSettings[K]) {
  emit("update:model", { ...props.model, [key]: value });
}
</script>

<template>
  <section class="border rounded-xl p-6 flex flex-col gap-5" :class="cardClasses">
    <div class="flex items-center justify-between">
      <h3 class="font-headline text-headline-sm text-on-surface flex items-center gap-2">
        <span class="material-symbols-outlined text-[20px] text-on-surface-variant">{{ icon }}</span>
        {{ title }}
      </h3>
      <span
        v-if="active"
        class="text-xs font-medium px-2.5 py-1 rounded-full bg-primary text-on-primary"
      >
        Active
      </span>
    </div>

    <div class="flex flex-col gap-1.5">
      <label :for="`${id}_url`" class="text-xs font-medium text-on-surface-variant uppercase tracking-wider">
        URL
      </label>
      <input
        :id="`${id}_url`"
        :value="model.url"
        type="text"
        class="text-sm px-3 py-2 rounded border border-outline-variant bg-surface text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:border-primary"
        @input="set('url', ($event.target as HTMLInputElement).value)"
      />
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-5">
      <div class="flex flex-col gap-1.5">
        <label :for="`${id}_vision`" class="text-xs font-medium text-on-surface-variant uppercase tracking-wider">
          Vision Model
        </label>
        <input
          :id="`${id}_vision`"
          :value="model.vision_model"
          type="text"
          placeholder="Gemma-4-E4B"
          class="text-sm px-3 py-2 rounded border border-outline-variant bg-surface text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:border-primary"
          @input="set('vision_model', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="flex flex-col gap-1.5">
        <label :for="`${id}_classify`" class="text-xs font-medium text-on-surface-variant uppercase tracking-wider">
          Classification Model
        </label>
        <input
          :id="`${id}_classify`"
          :value="model.classify_model"
          type="text"
          placeholder="Gemma-4-E4B"
          class="text-sm px-3 py-2 rounded border border-outline-variant bg-surface text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:border-primary"
          @input="set('classify_model', ($event.target as HTMLInputElement).value)"
        />
      </div>
      <div class="flex flex-col gap-1.5">
        <label :for="`${id}_embed`" class="text-xs font-medium text-on-surface-variant uppercase tracking-wider">
          Embedding Model (optional)
        </label>
        <input
          :id="`${id}_embed`"
          :value="model.embed_model"
          type="text"
          placeholder="embeddinggemma-2"
          class="text-sm px-3 py-2 rounded border border-outline-variant bg-surface text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:border-primary"
          @input="set('embed_model', ($event.target as HTMLInputElement).value)"
        />
        <p class="text-xs text-on-surface-variant">Enables semantic search. Empty disables it.</p>
      </div>
      <div class="flex flex-col gap-1.5">
        <label :for="`${id}_embed_url`" class="text-xs font-medium text-on-surface-variant uppercase tracking-wider">
          Embedding URL (optional)
        </label>
        <input
          :id="`${id}_embed_url`"
          :value="model.embed_url"
          type="text"
          :placeholder="model.url"
          class="text-sm px-3 py-2 rounded border border-outline-variant bg-surface text-on-surface placeholder:text-on-surface-variant focus:outline-none focus:border-primary"
          @input="set('embed_url', ($event.target as HTMLInputElement).value)"
        />
        <p class="text-xs text-on-surface-variant">
          Separate server for the embedding model (llama.cpp serves one model per process). Empty uses the main URL.
        </p>
      </div>
    </div>
  </section>
</template>
