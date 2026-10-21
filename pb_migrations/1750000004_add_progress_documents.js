/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const documents = app.findCollectionByNameOrId("documents");
  if (!documents.fields.find((f) => f.name === "progress")) {
    documents.fields.add(new JSONField({ name: "progress" }));
    app.save(documents);
  }
}, (app) => {
  const documents = app.findCollectionByNameOrId("documents");
  documents.fields.removeByName("progress");
  return app.save(documents);
});
