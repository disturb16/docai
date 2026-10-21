<script setup lang="ts">
import { computed, onMounted } from "vue";
import { RouterLink, RouterView } from "vue-router";
import { useSettings } from "./composables/useSettings";

const { settings, fetch } = useSettings();

onMounted(fetch);

const backendLabel = computed(() =>
  settings.value?.backend === "ollama" ? "Ollama: Local" : "llama.cpp: Local",
);
</script>

<template>
  <div class="min-h-screen flex">
    <nav
      class="bg-surface-container-low border-r border-surface-container-highest w-64 fixed left-0 top-0 h-full flex-col py-gutter px-4 z-40 hidden md:flex"
    >
      <div class="mb-8 pl-2">
        <h1 class="font-headline text-headline-sm font-bold text-primary">DocAI</h1>
        <p class="text-xs text-on-surface-variant uppercase mt-1">Local-First AI</p>
      </div>
      <div class="flex flex-col gap-2 flex-grow">
        <RouterLink
          to="/"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-on-surface-variant hover:bg-surface-container-high transition-colors"
          active-class="bg-secondary-container text-on-secondary-container font-semibold"
        >
          <span class="material-symbols-outlined">dashboard</span>
          Dashboard
        </RouterLink>
        <RouterLink
          to="/search"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-on-surface-variant hover:bg-surface-container-high transition-colors"
          active-class="bg-secondary-container text-on-secondary-container font-semibold"
        >
          <span class="material-symbols-outlined">search</span>
          Search
        </RouterLink>
        <RouterLink
          to="/upload"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-on-surface-variant hover:bg-surface-container-high transition-colors"
          active-class="bg-secondary-container text-on-secondary-container font-semibold"
        >
          <span class="material-symbols-outlined">cloud_upload</span>
          Upload
        </RouterLink>
        <RouterLink
          to="/activity"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-on-surface-variant hover:bg-surface-container-high transition-colors"
          active-class="bg-secondary-container text-on-secondary-container font-semibold"
        >
          <span class="material-symbols-outlined">history</span>
          Activity
        </RouterLink>
        <RouterLink
          to="/analytics"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-on-surface-variant hover:bg-surface-container-high transition-colors"
          active-class="bg-secondary-container text-on-secondary-container font-semibold"
        >
          <span class="material-symbols-outlined">monitoring</span>
          Analytics
        </RouterLink>
        <RouterLink
          to="/settings"
          class="flex items-center gap-3 px-3 py-2 rounded-lg text-on-surface-variant hover:bg-surface-container-high transition-colors"
          active-class="bg-secondary-container text-on-secondary-container font-semibold"
        >
          <span class="material-symbols-outlined">settings</span>
          Settings
        </RouterLink>
      </div>
      <div class="mt-auto pt-4 border-t border-surface-container-highest">
        <p class="text-xs text-on-surface-variant">Local Mode</p>
      </div>
    </nav>
    <div class="flex-1 md:ml-64 flex flex-col min-h-screen">
      <header
        class="bg-surface border-b border-surface-container-highest h-16 flex items-center justify-between px-margin-page sticky top-0 z-50"
      >
        <h2 class="font-headline text-headline-sm font-bold text-primary">DocAI</h2>
        <div class="flex items-center gap-2 bg-surface-container px-3 py-1 rounded-full border border-surface-container-highest">
          <span class="w-2 h-2 rounded-full bg-status-done"></span>
          <span class="text-xs text-on-surface-variant">{{ backendLabel }}</span>
        </div>
      </header>
      <main class="flex-1 p-margin-page bg-surface-bright">
        <RouterView />
      </main>
    </div>
  </div>
</template>
