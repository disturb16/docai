/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = new Collection({
    name: "documents",
    type: "base",
    listRule: "1=1",
    viewRule: "1=1",
    createRule: "1=1",
    updateRule: "1=1",
    deleteRule: "1=1",
    fields: [
      {
        name: "file",
        type: "file",
        maxSelect: 1,
        maxSize: 52428800,
        mimeTypes: [
          "application/pdf",
          "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
          "application/vnd.oasis.opendocument.text",
          "text/markdown",
          "text/x-markdown",
          "text/plain",
          "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
          "image/png",
          "image/jpeg",
        ],
      },
      { name: "filename", type: "text" },
      { name: "mime", type: "text" },
      { name: "size", type: "number" },
      {
        name: "status",
        type: "select",
        maxSelect: 1,
        values: ["pending", "processing", "done", "error"],
      },
      { name: "error", type: "text" },
      { name: "ocr_text", type: "text" },
      { name: "title", type: "text" },
      { name: "summary", type: "text" },
      { name: "category", type: "text" },
      { name: "tags", type: "json" },
      { name: "entities", type: "json" },
      { name: "language", type: "text" },
      { name: "model", type: "text" },
    ],
  });

  return app.save(collection);
}, (app) => {
  const collection = app.findCollectionByNameOrId("documents");

  return app.delete(collection);
});
