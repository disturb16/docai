/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const documents = app.findCollectionByNameOrId("documents");
  if (!documents.fields.find((f) => f.name === "retry_count")) {
    documents.fields.add(new NumberField({ name: "retry_count", onlyInt: true, min: 0 }));
    app.save(documents);
  }
}, (app) => {
  const documents = app.findCollectionByNameOrId("documents");
  documents.fields.removeByName("retry_count");
  return app.save(documents);
});
