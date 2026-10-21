/// <reference path="../pb_data/types.d.ts" />
migrate((app) => {
  const collection = app.findCollectionByNameOrId("documents");

  collection.fields.find((f) => f.name === "ocr_text").max = 1000000;

  return app.save(collection);
}, (app) => {
  const collection = app.findCollectionByNameOrId("documents");

  collection.fields.find((f) => f.name === "ocr_text").max = 0;

  return app.save(collection);
});
