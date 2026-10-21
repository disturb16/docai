/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("documents");

  collection.fields.add(
    new AutodateField({
      name: "created",
      onCreate: true,
      onUpdate: false,
      system: true,
    }),
    new AutodateField({
      name: "updated",
      onCreate: true,
      onUpdate: true,
      system: true,
    }),
  );

  app.save(collection);

  app
    .db()
    .newQuery(
      "UPDATE documents SET created = datetime('now'), updated = datetime('now') WHERE created = ''",
    )
    .execute();
}, (app) => {
  const collection = app.findCollectionByNameOrId("documents");

  collection.fields.removeByName("created");
  collection.fields.removeByName("updated");

  return app.save(collection);
});
