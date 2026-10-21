import { ref } from "vue";

export interface ProviderSettings {
  url: string;
  vision_model: string;
  classify_model: string;
  embed_model: string;
  embed_url: string;
}

export interface Settings {
  backend: "ollama" | "llamacpp";
  ollama: ProviderSettings;
  llamacpp: ProviderSettings;
  concurrency: number;
}

export interface Health {
  online: boolean;
  error?: string;
}

export interface BackfillResult {
  queued: number;
  skipped: number;
}

export function useSettings() {
  const settings = ref<Settings | null>(null);
  const loading = ref(false);
  const error = ref("");
  const saved = ref(false);
  const health = ref<Health | null>(null);
  const checking = ref(false);
  const backfilling = ref(false);
  const backfillResult = ref<BackfillResult | null>(null);

  async function fetch() {
    loading.value = true;
    error.value = "";
    try {
      const res = await globalThis.fetch("/api/docai/settings");
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      settings.value = await res.json();
    } catch (e) {
      error.value = String(e);
    } finally {
      loading.value = false;
    }
  }

  async function save(next: Settings) {
    error.value = "";
    saved.value = false;
    try {
      const res = await globalThis.fetch("/api/docai/settings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(next),
      });
      if (!res.ok) {
        const body = await res.json().catch(() => null);
        throw new Error(body?.message ?? `HTTP ${res.status}`);
      }
      settings.value = await res.json();
      saved.value = true;
    } catch (e) {
      error.value = String(e);
      throw e;
    }
  }

  async function checkHealth() {
    checking.value = true;
    try {
      const res = await globalThis.fetch("/api/docai/settings/health");
      health.value = res.ok
        ? await res.json()
        : { online: false, error: `HTTP ${res.status}` };
    } catch (e) {
      health.value = { online: false, error: String(e) };
    } finally {
      checking.value = false;
    }
  }

  async function backfill() {
    backfilling.value = true;
    backfillResult.value = null;
    try {
      const res = await globalThis.fetch("/api/docai/embeddings/backfill", {
        method: "POST",
      });
      if (!res.ok) {
        const body = await res.json().catch(() => null);
        throw new Error(body?.message ?? `HTTP ${res.status}`);
      }
      backfillResult.value = await res.json();
    } finally {
      backfilling.value = false;
    }
  }

  return {
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
  };
}
