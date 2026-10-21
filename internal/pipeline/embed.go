package pipeline

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"docai/config"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// ponytail: 6000 runes is a safe floor for the 8K-token window (CJK ~1 token
// per rune); raise per language mix if quality on long docs matters
const chunkMaxRunes = 6000

// ponytail: embeddings past 32 chunks are strided down to 32; store chunk
// vectors and do max-sim retrieval when long-doc recall starts to hurt
const maxChunks = 32

// chunkText splits text on blank-line paragraphs and packs them greedily into
// chunks of at most maxRunes runes. Oversized paragraphs are hard-split. No
// chunk is dropped: when the cap is exceeded, chunks are sampled uniformly.
func chunkText(text string, maxRunes int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}

	var chunks []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			chunks = append(chunks, strings.TrimSpace(string(cur)))
			cur = cur[:0]
		}
	}
	for _, para := range strings.Split(text, "\n\n") {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}
		for len(para) > 0 {
			runes := []rune(para)
			room := maxRunes - len(cur)
			if room <= 0 {
				flush()
				room = maxRunes
			}
			if len(runes) <= room {
				cur = append(cur, runes...)
				break
			}
			cur = append(cur, runes[:room]...)
			flush()
			para = string(runes[room:])
		}
		cur = append(cur, '\n', '\n')
	}
	flush()

	if len(chunks) > maxChunks {
		strided := make([]string, 0, maxChunks)
		for i := 0; i < maxChunks; i++ {
			idx := i * (len(chunks) - 1) / (maxChunks - 1)
			strided = append(strided, chunks[idx])
		}
		chunks = strided
	}
	return chunks
}

// documentInputs formats chunks for indexing: one input per chunk with the
// model's document prefix (title: {title} | text: {content}).
func documentInputs(title string, chunks []string) []string {
	if title == "" {
		title = "none"
	}
	inputs := make([]string, len(chunks))
	for i, c := range chunks {
		inputs[i] = "title: " + title + " | text: " + c
	}
	return inputs
}

// QueryInput formats a search query with the model's search query prefix.
func QueryInput(q string) string {
	return "task: search result | query: " + q
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Embed calls the OpenAI-compatible /v1/embeddings endpoint (llama.cpp and
// Ollama both implement it) and returns one vector per input.
func Embed(ctx context.Context, url, model string, inputs []string) ([][]float32, error) {
	body, err := json.Marshal(embedRequest{Model: model, Input: inputs})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL(url)+"/v1/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("embeddings http %d: %s", resp.StatusCode, snippet(string(data)))
	}
	var out embedResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if out.Error.Message != "" {
		return nil, errors.New(out.Error.Message)
	}
	if len(out.Data) != len(inputs) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(inputs), len(out.Data))
	}
	vecs := make([][]float32, len(out.Data))
	for i, d := range out.Data {
		if len(d.Embedding) == 0 {
			return nil, fmt.Errorf("empty embedding at index %d", i)
		}
		vecs[i] = d.Embedding
	}
	return vecs, nil
}

// meanPool averages vectors and L2-normalizes the result. Embedding one
// document is not a hot path, so a naive loop is fine.
func meanPool(vecs [][]float32) []float32 {
	if len(vecs) == 0 {
		return nil
	}
	out := make([]float32, len(vecs[0]))
	for _, v := range vecs {
		if len(v) != len(out) {
			return nil
		}
		for i, x := range v {
			out[i] += x
		}
	}
	n := float32(len(vecs))
	for i := range out {
		out[i] /= n
	}
	return Normalize(out)
}

// Normalize scales v to unit length (returns v unchanged when its norm is 0).
func Normalize(v []float32) []float32 {
	var norm float64
	for _, x := range v {
		norm += float64(x) * float64(x)
	}
	if norm == 0 {
		return v
	}
	inv := float32(1 / math.Sqrt(norm))
	for i := range v {
		v[i] *= inv
	}
	return v
}

// EmbedText embeds a document's text (chunked) into a single mean-pooled,
// normalized vector. Chunks are embedded one request at a time so a chunk
// rejected by the server (small batch/context) only skips that chunk. Empty
// text or an unconfigured model yields nil; zero usable chunks yields an error.
func EmbedText(ctx context.Context, cfg config.Config, title, text string) ([]float32, error) {
	url, model := cfg.Active().Embed()
	if model == "" {
		return nil, nil
	}
	chunks := chunkText(text, chunkMaxRunes)
	if len(chunks) == 0 {
		return nil, nil
	}
	var vecs [][]float32
	for _, input := range documentInputs(title, chunks) {
		one, err := Embed(ctx, url, model, []string{input})
		if err != nil {
			slog.Warn("embed: chunk failed", "model", model, "chunks_ok", len(vecs), "error", err)
			continue
		}
		vecs = append(vecs, one[0])
	}
	if len(vecs) == 0 {
		return nil, errors.New("all chunks failed to embed")
	}
	return meanPool(vecs), nil
}

// StoreEmbedding upserts the embedding row for a document. A nil vector clears
// any existing row (reprocess with embedding disabled or failed), so search
// never ranks a stale vector. Keeps documents free of hidden search fields.
func StoreEmbedding(app core.App, docID, model string, vec []float32) error {
	existing, err := app.FindFirstRecordByFilter("embeddings", "document = {:id}", dbx.Params{"id": docID})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if existing != nil {
		if vec == nil {
			return app.Delete(existing)
		}
		existing.Set("vector", vec)
		existing.Set("model", model)
		return app.Save(existing)
	}
	if vec == nil {
		return nil
	}
	collection, err := app.FindCollectionByNameOrId("embeddings")
	if err != nil {
		return err
	}
	record := core.NewRecord(collection)
	record.Set("document", docID)
	record.Set("vector", vec)
	record.Set("model", model)
	return app.Save(record)
}
