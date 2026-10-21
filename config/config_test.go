package config

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalJSON(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want Config
	}{
		{
			name: "current shape",
			raw:  `{"backend":"ollama","ollama":{"url":"http://o:1","vision_model":"ov","classify_model":"oc"},"llamacpp":{"url":"http://l:2","vision_model":"lv","classify_model":"lc"},"concurrency":3}`,
			want: Config{
				Backend:     "ollama",
				Ollama:      Provider{URL: "http://o:1", VisionModel: "ov", ClassifyModel: "oc"},
				LlamaCpp:    Provider{URL: "http://l:2", VisionModel: "lv", ClassifyModel: "lc"},
				Concurrency: 3,
			},
		},
		{
			name: "legacy flat shape migrates shared models to both providers",
			raw:  `{"backend":"llamacpp","ollama_url":"http://o:1","llama_url":"http://l:2","vision_model":"v","classify_model":"c","concurrency":2}`,
			want: Config{
				Backend:     "llamacpp",
				Ollama:      Provider{URL: "http://o:1", VisionModel: "v", ClassifyModel: "c"},
				LlamaCpp:    Provider{URL: "http://l:2", VisionModel: "v", ClassifyModel: "c"},
				Concurrency: 2,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got Config
			if err := json.Unmarshal([]byte(tc.raw), &got); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

func TestActive(t *testing.T) {
	cfg := Config{
		Backend:  "ollama",
		Ollama:   Provider{URL: "http://o"},
		LlamaCpp: Provider{URL: "http://l"},
	}
	if got := cfg.Active(); got != cfg.Ollama {
		t.Fatalf("got %+v, want ollama provider", got)
	}
	cfg.Backend = "llamacpp"
	if got := cfg.Active(); got != cfg.LlamaCpp {
		t.Fatalf("got %+v, want llamacpp provider", got)
	}
}

func TestEmbed(t *testing.T) {
	p := Provider{URL: "http://main", EmbedModel: "m"}
	if url, model := p.Embed(); url != "http://main" || model != "m" {
		t.Fatalf("got %q %q, want main URL fallback", url, model)
	}
	p.EmbedURL = "http://embed"
	if url, _ := p.Embed(); url != "http://embed" {
		t.Fatalf("got %q, want embed URL override", url)
	}
	if _, model := (Provider{URL: "http://main"}).Embed(); model != "" {
		t.Fatalf("got %q, want empty model when unset", model)
	}
}
