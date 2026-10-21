import { assertEquals } from "@std/assert";

import { breadcrumbOf, childrenOf, descendantCount } from "./folders.ts";

const folders = [
  { id: "a", name: "A", parent: "", created: "" },
  { id: "b", name: "B", parent: "a", created: "" },
  { id: "c", name: "C", parent: "a", created: "" },
  { id: "d", name: "D", parent: "b", created: "" },
];

Deno.test("childrenOf returns direct children sorted by name", () => {
  assertEquals(
    childrenOf(folders, "a").map((f) => f.id),
    ["b", "c"],
  );
  assertEquals(childrenOf(folders, "d"), []);
});

Deno.test("breadcrumbOf walks up to the root", () => {
  assertEquals(
    breadcrumbOf(folders, "d").map((f) => f.id),
    ["a", "b", "d"],
  );
  assertEquals(breadcrumbOf(folders, "a").map((f) => f.id), ["a"]);
  assertEquals(breadcrumbOf(folders, "missing"), []);
});

Deno.test("descendantCount counts nested folders", () => {
  assertEquals(descendantCount(folders, "a"), 3);
  assertEquals(descendantCount(folders, "b"), 1);
  assertEquals(descendantCount(folders, "c"), 0);
});
