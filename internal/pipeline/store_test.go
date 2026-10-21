package pipeline

import (
	"encoding/json"
	"testing"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestStoreEmbedding(t *testing.T) {
	app, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer app.Cleanup()

	documents := core.NewBaseCollection("documents")
	if err := app.Save(documents); err != nil {
		t.Fatal(err)
	}
	embeddings := core.NewBaseCollection("embeddings")
	embeddings.Fields.Add(&core.RelationField{
		Name:          "document",
		Required:      true,
		MaxSelect:     1,
		CollectionId:  documents.Id,
		CascadeDelete: true,
	})
	embeddings.Fields.Add(&core.JSONField{Name: "vector"})
	embeddings.Fields.Add(&core.TextField{Name: "model"})
	if err := app.Save(embeddings); err != nil {
		t.Fatal(err)
	}

	doc := core.NewRecord(documents)
	if err := app.Save(doc); err != nil {
		t.Fatal(err)
	}

	find := func() (record *core.Record) {
		record, _ = app.FindFirstRecordByFilter("embeddings", "document = {:id}", dbx.Params{"id": doc.Id})
		return
	}
	numFloats := func(v any) int {
		raw, _ := json.Marshal(v)
		var out []float64
		if err := json.Unmarshal(raw, &out); err != nil {
			t.Fatal(err)
		}
		return len(out)
	}

	if err := StoreEmbedding(app, doc.Id, "m", []float32{1, 2}); err != nil {
		t.Fatal(err)
	}
	got := find()
	if got == nil || got.GetString("model") != "m" || numFloats(got.Get("vector")) != 2 {
		t.Fatalf("after insert got %v", got)
	}
	if rows, err := app.FindAllRecords("embeddings"); err != nil || len(rows) != 1 {
		t.Fatalf("want 1 row after insert, got %d (err %v)", len(rows), err)
	}

	if err := StoreEmbedding(app, doc.Id, "m2", []float32{3, 4, 5}); err != nil {
		t.Fatal(err)
	}
	got = find()
	if got == nil || got.GetString("model") != "m2" || numFloats(got.Get("vector")) != 3 {
		t.Fatalf("after update got %v", got)
	}
	if rows, err := app.FindAllRecords("embeddings"); err != nil || len(rows) != 1 {
		t.Fatalf("want 1 row after update, got %d (err %v)", len(rows), err)
	}

	if err := StoreEmbedding(app, doc.Id, "", nil); err != nil {
		t.Fatal(err)
	}
	if find() != nil {
		t.Fatal("nil vector must delete the row")
	}

	if err := StoreEmbedding(app, doc.Id, "", nil); err != nil {
		t.Fatal("clearing a missing row must be a no-op")
	}
}
