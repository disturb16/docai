export interface Folder {
  id: string;
  name: string;
  parent: string;
  created: string;
  updated: string;
}

export function childrenOf(folders: Folder[], parentId: string): Folder[] {
  return folders
    .filter((f) => f.parent === parentId)
    .sort((a, b) => a.name.localeCompare(b.name));
}

export function breadcrumbOf(folders: Folder[], id: string): Folder[] {
  const byId = new Map(folders.map((f) => [f.id, f]));
  const path: Folder[] = [];
  let current = byId.get(id);
  while (current) {
    path.unshift(current);
    current = byId.get(current.parent);
  }
  return path;
}

export function descendantCount(folders: Folder[], id: string): number {
  const byParent = new Map<string, string[]>();
  for (const f of folders) {
    if (!f.parent) continue;
    const list = byParent.get(f.parent) ?? [];
    list.push(f.id);
    byParent.set(f.parent, list);
  }
  let count = 0;
  const stack = [id];
  while (stack.length) {
    const current = stack.pop()!;
    for (const child of byParent.get(current) ?? []) {
      count++;
      stack.push(child);
    }
  }
  return count;
}
