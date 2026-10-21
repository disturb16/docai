import { ref } from "vue";
import { pb } from "../lib/pb";

export interface Progress {
  stage: string;
  current?: number;
  total?: number;
}

export function progressLabel(p?: Progress): string {
  if (!p) return "Processing...";
  switch (p.stage) {
    case "extract":
      return "Extracting text...";
    case "ocr":
      return p.total && p.total > 1
        ? `OCR page ${p.current ?? 1}/${p.total}...`
        : "Reading image...";
    case "classify":
      return "Classifying...";
    case "embed":
      return "Embedding...";
    default:
      return "Processing...";
  }
}

export interface Document {
  id: string;
  file: string;
  filename: string;
  mime: string;
  size: number;
  status: "pending" | "processing" | "done" | "error";
  error: string;
  ocr_text: string;
  title: string;
  summary: string;
  category: string;
  tags: string[];
  entities: string[];
  language: string;
  model: string;
  folder: string;
  retry_count: number;
  embed_status: "" | "pending" | "processing" | "done" | "error";
  embed_error: string;
  progress?: Progress;
  created: string;
}

export function useDocuments() {
  const documents = ref<Document[]>([]);
  const loading = ref(false);
  const error = ref("");

  async function fetchAll() {
    loading.value = true;
    error.value = "";
    try {
      const records = await pb.collection("documents").getFullList<Document>({
        sort: "-created",
      });
      documents.value = records;
    } catch (e) {
      error.value = String(e);
    } finally {
      loading.value = false;
    }
  }

  function subscribe() {
    return pb.collection("documents").subscribe("*", (e) => {
      const record = e.record as Document;
      const index = documents.value.findIndex((d) => d.id === record.id);
      if (e.action === "delete") {
        if (index >= 0) documents.value.splice(index, 1);
        return;
      }
      if (index >= 0) {
        documents.value[index] = record;
      } else {
        documents.value.unshift(record);
      }
    });
  }

  async function remove(id: string) {
    await pb.collection("documents").delete(id);
  }

  async function upload(files: File[], folder = "") {
    const records: Document[] = [];
    for (const file of files) {
      const record = await pb.collection("documents").create<Document>({
        file,
        filename: file.name,
        mime: file.type,
        size: file.size,
        status: "pending",
        folder,
      });
      records.push(record);
    }
    return records;
  }

  return { documents, loading, error, fetchAll, subscribe, upload, remove };
}
