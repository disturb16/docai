/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const documents = app.findCollectionByNameOrId("documents");
  if (!documents.fields.find((f) => f.name === "embed_status")) {
    documents.fields.add(
      new SelectField({
        name: "embed_status",
        maxSelect: 1,
        values: ["pending", "processing", "done", "error"],
      }),
    );
    documents.fields.add(new TextField({ name: "embed_error" }));
    app.save(documents);
  }
}, (app) => {
  const documents = app.findCollectionByNameOrId("documents");
  documents.fields.removeByName("embed_status");
  documents.fields.removeByName("embed_error");
  return app.save(documents);
});
