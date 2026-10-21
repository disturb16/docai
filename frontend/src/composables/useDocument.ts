import { ref } from "vue";
import { pb } from "../lib/pb";
import type { Document } from "./useDocuments";

export interface SimilarHit {
  item: Document;
  score: number;
}

export function useDocument(id: string) {
  const document = ref<Document | null>(null);
  const loading = ref(false);
  const error = ref("");
  const similar = ref<SimilarHit[]>([]);
  const similarLoading = ref(false);
  const reembedding = ref(false);

  async function fetch() {
    loading.value = true;
    error.value = "";
    try {
      document.value = await pb.collection("documents").getOne<Document>(id);
    } catch (e) {
      error.value = String(e);
    } finally {
      loading.value = false;
    }
  }

  function subscribe() {
    return pb.collection("documents").subscribe(id, (e) => {
      document.value = e.action === "delete" ? null : (e.record as Document);
    });
  }

  async function update(data: Partial<Document>) {
    document.value = await pb.collection("documents").update<Document>(
      id,
      data,
    );
  }

  async function findSimilar() {
    similarLoading.value = true;
    try {
      const res = await pb.send(`/api/documents/${id}/similar`);
      similar.value = (res.similar ?? []) as SimilarHit[];
    } catch {
      similar.value = [];
    } finally {
      similarLoading.value = false;
    }
  }

  async function reembed() {
    reembedding.value = true;
    try {
      await pb.send(`/api/documents/${id}/embed`, { method: "POST" });
    } finally {
      reembedding.value = false;
    }
  }

  return {
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
  };
}
