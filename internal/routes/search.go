package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"regexp"
	"sort"
	"strings"

	"docai/config"
	"docai/internal/pipeline"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"
)

var tokenRe = regexp.MustCompile(`[a-z0-9]+`)

// ponytail: brute-force cosine over all embedded docs; add an ANN index when
// the library grows past a few thousand docs
const (
	similarMinScore = 0.25
	similarLimit    = 10
)

// registerSearch adds GET /api/search?q=...&semantic=1 and
// GET /api/documents/{id}/similar. Query tokens are matched as whole words
// (AND semantics) against filename, title, summary, category, tags, entities,
// and ocr_text. Semantic ranking (embedding the query) only runs when the
// caller passes semantic=1 — the embed model is exercised on explicit
// request, not on every search. The similar endpoint ranks stored vectors
// directly, so it works offline and regardless of the currently configured
// model.
func registerSearch(app core.App, e *core.ServeEvent, store *config.ConfigStore) {
	e.Router.GET("/api/search", func(e *core.RequestEvent) error {
		q := strings.TrimSpace(e.Request.URL.Query().Get("q"))
		if q == "" {
			return e.JSON(200, map[string]any{"items": []*core.Record{}, "similar": []similarHit{}})
		}

		tokens := tokenize(q)
		// ponytail: in-memory scan of all records; switch to SQLite FTS5 or a
		// search_text column with LIKE when the library grows past a few hundred docs.
		records, err := app.FindAllRecords("documents")
		if err != nil {
			return e.InternalServerError("failed to fetch documents", err)
		}

		items := make([]*core.Record, 0, len(records))
		matched := make(map[string]struct{}, len(records))
		for _, r := range records {
			if allIn(tokens, haystack(r)) {
				items = append(items, r)
				matched[r.Id] = struct{}{}
			}
		}

		similar := []similarHit{}
		if e.Request.URL.Query().Get("semantic") == "1" {
			embeddings, err := app.FindAllRecords("embeddings")
			if err != nil {
				return e.InternalServerError("failed to fetch embeddings", err)
			}
			similar = semanticMatches(e.Request.Context(), store.Get(), byDocument(records), embeddings, q, matched)
		}

		return e.JSON(200, map[string]any{
			"items":   items,
			"similar": similar,
		})
	})

	e.Router.GET("/api/documents/{id}/similar", func(e *core.RequestEvent) error {
		id := e.Request.PathValue("id")
		if _, err := app.FindRecordById("documents", id); err != nil {
			return e.NotFoundError("document not found", err)
		}
		embedding, err := app.FindFirstRecordByFilter("embeddings", "document = {:id}", dbx.Params{"id": id})
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return e.InternalServerError("failed to fetch embedding", err)
		}
		if embedding == nil || len(toVector(embedding.Get("vector"))) == 0 {
			return e.JSON(200, map[string]any{"similar": []similarHit{}})
		}
		// ponytail: same in-memory scan as /api/search; ANN index when large
		records, err := app.FindAllRecords("documents")
		if err != nil {
			return e.InternalServerError("failed to fetch documents", err)
		}
		embeddings, err := app.FindAllRecords("embeddings")
		if err != nil {
			return e.InternalServerError("failed to fetch embeddings", err)
		}
		skip := map[string]struct{}{id: {}}
		return e.JSON(200, map[string]any{
			"similar": rankSimilar(toVector(embedding.Get("vector")), byDocument(records), embeddings, embedding.GetString("model"), skip),
		})
	})
}

// byDocument indexes documents by id for embedding-row lookups.
func byDocument(records []*core.Record) map[string]*core.Record {
	docs := make(map[string]*core.Record, len(records))
	for _, r := range records {
		docs[r.Id] = r
	}
	return docs
}

type similarHit struct {
	Item  *core.Record `json:"item"`
	Score float64      `json:"score"`
}

// semanticMatches embeds the query and ranks stored document embeddings by
// cosine similarity, skipping ids already returned as keyword hits. Failures
// (model offline, embedding disabled) degrade to an empty list so keyword
// search keeps working.
func semanticMatches(ctx context.Context, cfg config.Config, docs map[string]*core.Record, embeddings []*core.Record, q string, skip map[string]struct{}) []similarHit {
	url, model := cfg.Active().Embed()
	if model == "" {
		return []similarHit{}
	}
	vecs, err := pipeline.Embed(ctx, url, model, []string{pipeline.QueryInput(q)})
	if err != nil || len(vecs) == 0 {
		slog.Warn("query embedding failed", "error", err)
		return []similarHit{}
	}
	// don't trust the server's pooling flag: cosine assumes unit vectors
	return rankSimilar(pipeline.Normalize(vecs[0]), docs, embeddings, model, skip)
}

// rankSimilar returns the top-N done documents whose stored vector is closest
// (cosine) to query, skipping ids in skip and embeddings made by a different
// model (their vector lives in another space).
func rankSimilar(query []float32, docs map[string]*core.Record, embeddings []*core.Record, model string, skip map[string]struct{}) []similarHit {
	hits := make([]similarHit, 0, len(embeddings))
	for _, e := range embeddings {
		doc := docs[e.GetString("document")]
		if doc == nil {
			continue
		}
		if _, ok := skip[doc.Id]; ok {
			continue
		}
		if doc.GetString("status") != "done" {
			continue
		}
		if model != "" && e.GetString("model") != model {
			continue
		}
		vec := toVector(e.Get("vector"))
		if len(vec) != len(query) {
			continue
		}
		score := cosine(query, vec)
		if score >= similarMinScore {
			hits = append(hits, similarHit{Item: doc, Score: score})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].Score > hits[j].Score })
	if len(hits) > similarLimit {
		hits = hits[:similarLimit]
	}
	return hits
}

// toVector decodes a stored embedding. PocketBase hands JSON fields back as
// raw JSON bytes; the other cases tolerate plain Go values (e.g. in tests).
func toVector(v any) []float32 {
	switch arr := v.(type) {
	case types.JSONRaw:
		var floats []float64
		if err := json.Unmarshal(arr, &floats); err != nil {
			return nil
		}
		return toVector(floats)
	case []byte:
		return toVector(types.JSONRaw(arr))
	case []float32:
		return arr
	case []float64:
		out := make([]float32, len(arr))
		for i, x := range arr {
			out[i] = float32(x)
		}
		return out
	case []any:
		out := make([]float32, 0, len(arr))
		for _, x := range arr {
			f, ok := x.(float64)
			if !ok {
				return nil
			}
			out = append(out, float32(f))
		}
		return out
	}
	return nil
}

// cosine assumes unit-length vectors (both sides are L2-normalized).
func cosine(a, b []float32) float64 {
	var dot float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
	}
	return dot
}

func tokenize(s string) map[string]struct{} {
	tokens := make(map[string]struct{})
	for _, t := range tokenRe.FindAllString(strings.ToLower(s), -1) {
		tokens[t] = struct{}{}
	}
	return tokens
}

func haystack(r *core.Record) map[string]struct{} {
	parts := []string{
		r.GetString("filename"),
		r.GetString("title"),
		r.GetString("summary"),
		r.GetString("category"),
		r.GetString("ocr_text"),
	}
	for _, field := range []string{"tags", "entities"} {
		for _, v := range r.GetStringSlice(field) {
			parts = append(parts, v)
		}
	}
	return tokenize(strings.Join(parts, " "))
}

func allIn(tokens, haystack map[string]struct{}) bool {
	for t := range tokens {
		if _, ok := haystack[t]; !ok {
			return false
		}
	}
	return true
}
