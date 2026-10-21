/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  let folders;
  try {
    folders = app.findCollectionByNameOrId("folders");
  } catch {
    folders = new Collection({
      name: "folders",
      type: "base",
      listRule: "1=1",
      viewRule: "1=1",
      createRule: "1=1",
      updateRule: "1=1",
      deleteRule: "1=1",
      fields: [{ name: "name", type: "text" }],
    });
    app.save(folders);
  }

  if (!folders.fields.find((f) => f.name === "parent")) {
    folders.fields.add(
      new RelationField({
        name: "parent",
        collectionId: folders.id,
        maxSelect: 1,
        cascadeDelete: true,
      }),
    );
    app.save(folders);
  }

  const documents = app.findCollectionByNameOrId("documents");
  if (!documents.fields.find((f) => f.name === "folder")) {
    documents.fields.add(
      new RelationField({
        name: "folder",
        collectionId: folders.id,
        maxSelect: 1,
        cascadeDelete: true,
      }),
    );
    app.save(documents);
  }
}, (app) => {
  const documents = app.findCollectionByNameOrId("documents");
  documents.fields.removeByName("folder");
  app.save(documents);

  const folders = app.findCollectionByNameOrId("folders");

  return app.delete(folders);
});
