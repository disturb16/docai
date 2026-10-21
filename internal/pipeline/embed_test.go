package pipeline

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"docai/config"
)

func TestChunkText(t *testing.T) {
	long := strings.Repeat("a", 25)
	chunks := chunkText("para one\n\n"+long, 10)
	if len(chunks) != 4 {
		t.Fatalf("got wrong chunk count: %d", len(chunks))
	}
	for _, c := range chunks[:3] {
		if len([]rune(c)) > 10 {
			t.Fatalf("chunk exceeds max runes: %q", c)
		}
	}
	if chunks[0] != "para one" {
		t.Fatalf("got first chunk %q", chunks[0])
	}
	if got := chunkText("   ", 10); got != nil {
		t.Fatalf("blank text should yield no chunks, got %v", got)
	}
}

func TestChunkTextSamplesAboveCap(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 100; i++ {
		b.WriteString("chunk body\n\n")
	}
	chunks := chunkText(b.String(), 5)
	if len(chunks) != maxChunks {
		t.Fatalf("got %d chunks, want %d", len(chunks), maxChunks)
	}
}

func TestDocumentInputsAndQueryInput(t *testing.T) {
	got := documentInputs("", []string{"body"})
	if got[0] != "title: none | text: body" {
		t.Fatalf("got %q", got[0])
	}
	got = documentInputs("My Doc", []string{"body"})
	if got[0] != "title: My Doc | text: body" {
		t.Fatalf("got %q", got[0])
	}
	if q := QueryInput("cats"); q != "task: search result | query: cats" {
		t.Fatalf("got %q", q)
	}
}

func TestMeanPool(t *testing.T) {
	vec := meanPool([][]float32{{3, 0}, {0, 4}})
	if len(vec) != 2 {
		t.Fatalf("got %v", vec)
	}
	var norm float64
	for _, x := range vec {
		norm += float64(x) * float64(x)
	}
	if math.Abs(norm-1) > 1e-6 {
		t.Fatalf("not unit length: %v", vec)
	}
	if math.Abs(float64(vec[0])-0.6) > 1e-6 || math.Abs(float64(vec[1])-0.8) > 1e-6 {
		t.Fatalf("got %v, want [0.6 0.8]", vec)
	}
	if meanPool(nil) != nil || meanPool([][]float32{{1, 2}, {1}}) != nil {
		t.Fatal("mismatched or empty input should yield nil")
	}
}

func TestNormalize(t *testing.T) {
	got := Normalize([]float32{3, 4})
	if math.Abs(float64(got[0])-0.6) > 1e-6 || math.Abs(float64(got[1])-0.8) > 1e-6 {
		t.Fatalf("got %v, want [0.6 0.8]", got)
	}
	if z := Normalize([]float32{0, 0}); z[0] != 0 || z[1] != 0 {
		t.Fatalf("zero vector should be returned unchanged, got %v", z)
	}
}

func TestEmbed(t *testing.T) {
	var gotBody embedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("got path %q", r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Write([]byte(`{"data":[{"embedding":[1,0]},{"embedding":[0,1]}]}`))
	}))
	defer srv.Close()

	vecs, err := Embed(context.Background(), srv.URL, "m", []string{"a", "b"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vecs) != 2 || vecs[0][0] != 1 || vecs[1][1] != 1 {
		t.Fatalf("got %v", vecs)
	}
	if gotBody.Model != "m" || len(gotBody.Input) != 2 {
		t.Fatalf("got request %+v", gotBody)
	}
}

func TestEmbedErrors(t *testing.T) {
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"error":{"message":"boom"}}`))
	}))
	defer bad.Close()
	if _, err := Embed(context.Background(), bad.URL, "m", []string{"a"}); err == nil {
		t.Fatal("want error on provider error")
	}

	short := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"data":[{"embedding":[1,0]}]}`))
	}))
	defer short.Close()
	if _, err := Embed(context.Background(), short.URL, "m", []string{"a", "b"}); err == nil {
		t.Fatal("want error on embedding count mismatch")
	}

	down := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`model not loaded`))
	}))
	defer down.Close()
	_, err := Embed(context.Background(), down.URL, "m", []string{"a"})
	if err == nil || !strings.Contains(err.Error(), "500") {
		t.Fatalf("want http 500 error, got %v", err)
	}
}

func TestSnippet(t *testing.T) {
	if got := snippet("  hello  "); got != "hello" {
		t.Fatalf("got %q", got)
	}
	long := strings.Repeat("x", 400)
	got := snippet(long)
	if !strings.HasSuffix(got, "...") || len([]rune(got)) != 303 {
		t.Fatalf("got %d runes", len([]rune(got)))
	}
}

func TestBaseURL(t *testing.T) {
	for in, want := range map[string]string{
		"http://127.0.0.1:8077/v1":  "http://127.0.0.1:8077",
		"http://127.0.0.1:8077/v1/": "http://127.0.0.1:8077",
		"http://127.0.0.1:8077":     "http://127.0.0.1:8077",
		"http://localhost:11434/":   "http://localhost:11434",
	} {
		if got := baseURL(in); got != want {
			t.Fatalf("baseURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLlamaGenerateSendsModel(t *testing.T) {
	var got llamaRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Errorf("got path %q", r.URL.Path)
		}
		json.NewDecoder(r.Body).Decode(&got)
		w.Write([]byte(`{"choices":[{"message":{"content":"hi"}}]}`))
	}))
	defer srv.Close()

	cfg := config.Config{LlamaCpp: config.Provider{URL: srv.URL + "/v1"}}
	out, err := llamaGenerate(context.Background(), cfg, "Gemma-4-E2B", "prompt", nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "hi" || got.Model != "Gemma-4-E2B" {
		t.Fatalf("got out=%q request=%+v", out, got)
	}
}
