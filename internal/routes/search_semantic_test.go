package routes

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"docai/config"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

func docRecord(id, status, filename string) *core.Record {
	r := core.NewRecord(core.NewBaseCollection("documents"))
	r.Id = id
	r.Set("status", status)
	r.Set("filename", filename)
	return r
}

func embeddingRecord(id, docID, model string, raw types.JSONRaw) *core.Record {
	e := core.NewRecord(core.NewBaseCollection("embeddings"))
	e.Id = id
	e.Set("document", docID)
	e.Set("model", model)
	e.Set("vector", raw)
	return e
}

func TestSemanticMatches(t *testing.T) {
	queryVec := make([]float64, 768)
	queryVec[0] = 1
	queryVecJSON, _ := json.Marshal(queryVec)

	otherVec := make([]float64, 768)
	otherVec[1] = 1
	otherVecJSON, _ := json.Marshal(otherVec)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[{"embedding":` + string(queryVecJSON) + `}]}`))
	}))
	defer srv.Close()

	close := docRecord("close", "done", "close.pdf")
	far := docRecord("far", "done", "far.pdf")
	keywordHit := docRecord("keyword", "done", "keyword.pdf")
	notDone := docRecord("pending", "processing", "")
	staleModel := docRecord("stale", "done", "stale.pdf")

	docs := []*core.Record{close, far, keywordHit, notDone, staleModel}
	embeddings := []*core.Record{
		embeddingRecord("e1", "close", "m", types.JSONRaw(queryVecJSON)),
		embeddingRecord("e2", "far", "m", types.JSONRaw(otherVecJSON)),
		embeddingRecord("e3", "keyword", "m", types.JSONRaw(queryVecJSON)),
		embeddingRecord("e4", "pending", "m", types.JSONRaw(queryVecJSON)),
		embeddingRecord("e5", "stale", "other-model", types.JSONRaw(queryVecJSON)),
	}

	cfg := config.Config{
		Backend:  "llamacpp",
		LlamaCpp: config.Provider{URL: srv.URL, EmbedModel: "m"},
	}

	hits := semanticMatches(context.Background(), cfg, byDocument(docs), embeddings, "q", map[string]struct{}{keywordHit.Id: {}})
	if len(hits) != 1 || hits[0].Item.GetString("filename") != "close.pdf" {
		t.Fatalf("got %d hits, want exactly close.pdf", len(hits))
	}
	if math.Abs(hits[0].Score-1) > 1e-6 {
		t.Fatalf("got score %v, want 1", hits[0].Score)
	}
}

func TestSemanticMatchesDisabled(t *testing.T) {
	cfg := config.Config{Backend: "llamacpp", LlamaCpp: config.Provider{URL: "http://x"}}
	if hits := semanticMatches(context.Background(), cfg, nil, nil, "q", nil); len(hits) != 0 {
		t.Fatalf("got %v, want no hits when model unset", hits)
	}
}

func TestRankSimilar(t *testing.T) {
	vec := func(x, y float32) types.JSONRaw {
		v, _ := json.Marshal([]float32{x, y})
		return types.JSONRaw(v)
	}

	docs := byDocument([]*core.Record{
		docRecord("a", "done", ""),
		docRecord("b", "done", ""),
		docRecord("other", "done", ""),
		docRecord("self", "done", ""),
	})
	embeddings := []*core.Record{
		embeddingRecord("e1", "self", "m", vec(1, 0)),
		embeddingRecord("e2", "b", "m", vec(0.7, 0.7)),
		embeddingRecord("e3", "other", "n", vec(1, 0)),
		embeddingRecord("e4", "a", "m", vec(1, 0)),
	}

	hits := rankSimilar([]float32{1, 0}, docs, embeddings, "m", map[string]struct{}{"self": {}})
	if len(hits) != 2 || hits[0].Item.Id != "a" || hits[1].Item.Id != "b" {
		t.Fatalf("got %d hits, want a then b (self skipped, other model tagged)", len(hits))
	}
}

func TestRankSimilarWithoutModelTag(t *testing.T) {
	v, _ := json.Marshal([]float32{1, 0})
	docs := byDocument([]*core.Record{docRecord("r", "done", "")})
	embeddings := []*core.Record{embeddingRecord("e1", "r", "", types.JSONRaw(v))}

	hits := rankSimilar([]float32{1, 0}, docs, embeddings, "", nil)
	if len(hits) != 1 {
		t.Fatalf("got %d hits, want 1 when no model tag is given", len(hits))
	}
}

func TestRankSimilarOrphanEmbedding(t *testing.T) {
	v, _ := json.Marshal([]float32{1, 0})
	embeddings := []*core.Record{embeddingRecord("e1", "gone", "m", types.JSONRaw(v))}

	if hits := rankSimilar([]float32{1, 0}, map[string]*core.Record{}, embeddings, "m", nil); len(hits) != 0 {
		t.Fatalf("got %d hits, want 0 for an embedding whose document is missing", len(hits))
	}
}

func TestSemanticMatchesProviderDown(t *testing.T) {
	cfg := config.Config{Backend: "llamacpp", LlamaCpp: config.Provider{URL: "http://127.0.0.1:1", EmbedModel: "m"}}
	if hits := semanticMatches(context.Background(), cfg, nil, nil, "q", nil); len(hits) != 0 {
		t.Fatalf("got %v, want no hits when provider unreachable", hits)
	}
}
