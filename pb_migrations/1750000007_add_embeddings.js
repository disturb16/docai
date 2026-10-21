/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const documents = app.findCollectionByNameOrId("documents");
  const embeddings = new Collection({
    name: "embeddings",
    type: "base",
    fields: [
      {
        name: "document",
        type: "relation",
        required: true,
        maxSelect: 1,
        collectionId: documents.id,
        cascadeDelete: true,
      },
      { name: "vector", type: "json" },
      { name: "model", type: "text" },
    ],
    indexes: [
      "CREATE UNIQUE INDEX `idx_embeddings_document` ON `embeddings` (`document`)",
    ],
  });
  return app.save(embeddings);
}, (app) => {
  const embeddings = app.findCollectionByNameOrId("embeddings");

  return app.delete(embeddings);
});
