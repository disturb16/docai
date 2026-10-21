import { ref } from "vue";
import { pb } from "../lib/pb";
import type { Folder } from "../lib/folders";

export function useFolders() {
  const folders = ref<Folder[]>([]);
  const loading = ref(false);
  const error = ref("");

  async function fetchAll() {
    loading.value = true;
    error.value = "";
    try {
      const records = await pb.collection("folders").getFullList<Folder>({
        sort: "name",
      });
      folders.value = records;
    } catch (e) {
      error.value = String(e);
    } finally {
      loading.value = false;
    }
  }

  function subscribe() {
    return pb.collection("folders").subscribe("*", (e) => {
      const record = e.record as Folder;
      const index = folders.value.findIndex((f) => f.id === record.id);
      if (e.action === "delete") {
        if (index >= 0) folders.value.splice(index, 1);
        return;
      }
      if (index >= 0) {
        folders.value[index] = record;
      } else {
        folders.value.push(record);
      }
    });
  }

  async function create(name: string, parent: string) {
    return await pb.collection("folders").create<Folder>({ name, parent });
  }

  async function rename(id: string, name: string) {
    await pb.collection("folders").update(id, { name });
  }

  async function remove(id: string) {
    await pb.collection("folders").delete(id);
  }

  return {
    folders,
    loading,
    error,
    fetchAll,
    subscribe,
    create,
    rename,
    remove,
  };
}
