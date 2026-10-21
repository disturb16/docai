package config

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"

	"github.com/pocketbase/pocketbase/core"
)

type Provider struct {
	URL           string `json:"url"`
	VisionModel   string `json:"vision_model"`
	ClassifyModel string `json:"classify_model"`
	EmbedModel    string `json:"embed_model"`
	// EmbedURL is a separate server for the embedding model (llama.cpp serves
	// one model per process); empty means the main URL (Ollama's case).
	EmbedURL string `json:"embed_url"`
}

// Embed returns the embedding endpoint URL and model. An empty model means
// semantic search is disabled.
func (p Provider) Embed() (string, string) {
	if p.EmbedURL != "" {
		return p.EmbedURL, p.EmbedModel
	}
	return p.URL, p.EmbedModel
}

type Config struct {
	Backend     string   `json:"backend"`
	Ollama      Provider `json:"ollama"`
	LlamaCpp    Provider `json:"llamacpp"`
	Concurrency int      `json:"concurrency"`
}

// Active returns the provider settings for the selected backend.
func (c Config) Active() Provider {
	if c.Backend == "ollama" {
		return c.Ollama
	}
	return c.LlamaCpp
}

// UnmarshalJSON accepts both the current per-provider shape and the legacy
// flat shape (shared vision/classify models), migrating the latter by
// copying the shared models into both providers.
func (c *Config) UnmarshalJSON(data []byte) error {
	var probe map[string]json.RawMessage
	if err := json.Unmarshal(data, &probe); err != nil {
		return err
	}
	if _, ok := probe["ollama"]; ok {
		type plain Config
		var fresh plain
		if err := json.Unmarshal(data, &fresh); err != nil {
			return err
		}
		*c = Config(fresh)
		return nil
	}
	var legacy struct {
		Backend       string `json:"backend"`
		OllamaURL     string `json:"ollama_url"`
		LlamaURL      string `json:"llama_url"`
		VisionModel   string `json:"vision_model"`
		ClassifyModel string `json:"classify_model"`
		Concurrency   int    `json:"concurrency"`
	}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return err
	}
	*c = Config{
		Backend: legacy.Backend,
		Ollama: Provider{
			URL:           legacy.OllamaURL,
			VisionModel:   legacy.VisionModel,
			ClassifyModel: legacy.ClassifyModel,
		},
		LlamaCpp: Provider{
			URL:           legacy.LlamaURL,
			VisionModel:   legacy.VisionModel,
			ClassifyModel: legacy.ClassifyModel,
		},
		Concurrency: legacy.Concurrency,
	}
	return nil
}

// ConfigStore holds the runtime config. Env vars seed the defaults; values saved
// from the settings page override them and persist in the _params table.
type ConfigStore struct {
	mu  sync.RWMutex
	cfg Config
}

func NewStore() *ConfigStore {
	return &ConfigStore{cfg: Load()}
}

func (s *ConfigStore) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func (s *ConfigStore) Set(cfg Config) {
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()
}

// LoadFromDB overrides env defaults with the persisted config, if any.
func (s *ConfigStore) LoadFromDB(app core.App) error {
	param := &core.Param{}
	err := app.ModelQuery(param).Model("docai_config", param)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if param.Id == "" {
		return nil
	}
	var cfg Config
	if err := json.Unmarshal(param.Value, &cfg); err != nil {
		return err
	}
	s.Set(cfg)
	return nil
}

// Save persists the current config to the _params table.
func (s *ConfigStore) Save(app core.App) error {
	raw, err := json.Marshal(s.Get())
	if err != nil {
		return err
	}
	param := &core.Param{}
	err = app.ModelQuery(param).Model("docai_config", param)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if param.Id == "" {
		param.Id = "docai_config"
		param.MarkAsNew()
	}
	param.Value = raw
	return app.Save(param)
}

func Load() Config {
	return Config{
		Backend: envOr("DOCAI_BACKEND", "llamacpp"), // ollama or llamacpp
		Ollama: Provider{
			URL:           envOr("DOCAI_OLLAMA_URL", "http://localhost:11434"),
			VisionModel:   modelEnv("OLLAMA", "VISION", "Gemma-4-E4B"),
			ClassifyModel: modelEnv("OLLAMA", "CLASSIFY", "Gemma-4-E4B"),
			EmbedModel:    modelEnv("OLLAMA", "EMBED", ""),
			EmbedURL:      os.Getenv("DOCAI_OLLAMA_EMBED_URL"),
		},
		LlamaCpp: Provider{
			URL:           envOr("DOCAI_LLAMA_URL", "http://127.0.0.1:42951"),
			VisionModel:   modelEnv("LLAMA", "VISION", "Gemma-4-E4B"),
			ClassifyModel: modelEnv("LLAMA", "CLASSIFY", "Gemma-4-E4B"),
			EmbedModel:    modelEnv("LLAMA", "EMBED", ""),
			EmbedURL:      os.Getenv("DOCAI_LLAMA_EMBED_URL"),
		},
		Concurrency: envIntOr("DOCAI_CONCURRENCY", 1),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// modelEnv checks DOCAI_<PROVIDER>_<KIND>_MODEL first, then the shared
// DOCAI_<KIND>_MODEL, then the fallback.
func modelEnv(provider, kind, fallback string) string {
	if v := os.Getenv("DOCAI_" + provider + "_" + kind + "_MODEL"); v != "" {
		return v
	}
	if v := os.Getenv("DOCAI_" + kind + "_MODEL"); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return fallback
}
