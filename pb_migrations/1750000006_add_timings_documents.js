/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const documents = app.findCollectionByNameOrId("documents");
  if (!documents.fields.find((f) => f.name === "timings")) {
    documents.fields.add(new JSONField({ name: "timings" }));
    app.save(documents);
  }
}, (app) => {
  const documents = app.findCollectionByNameOrId("documents");
  documents.fields.removeByName("timings");
  return app.save(documents);
});
