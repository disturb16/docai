package routes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"docai/config"

	"github.com/pocketbase/pocketbase/core"
)

// registerSettings adds GET/POST /api/docai/settings for reading and updating
// the runtime config. Values persist in the _params table and apply to new
// pipeline runs. (Prefixed to avoid clashing with PocketBase's own
// /api/settings route.)
func registerSettings(app core.App, e *core.ServeEvent, store *config.ConfigStore) {
	e.Router.GET("/api/docai/settings", func(e *core.RequestEvent) error {
		return e.JSON(200, store.Get())
	})

	e.Router.GET("/api/docai/settings/health", func(e *core.RequestEvent) error {
		return e.JSON(200, probe(store.Get()))
	})

	e.Router.POST("/api/docai/settings", func(e *core.RequestEvent) error {
		var cfg config.Config
		if err := json.NewDecoder(e.Request.Body).Decode(&cfg); err != nil {
			return e.BadRequestError("invalid JSON", err)
		}
		if err := validate(cfg); err != nil {
			return e.BadRequestError(err.Error(), err)
		}
		store.Set(cfg)
		if err := store.Save(app); err != nil {
			return e.InternalServerError("failed to save settings", err)
		}
		return e.JSON(200, store.Get())
	})
}

func validate(cfg config.Config) error {
	if cfg.Backend != "ollama" && cfg.Backend != "llamacpp" {
		return errors.New("backend must be 'ollama' or 'llamacpp'")
	}
	if err := validateProvider("ollama", cfg.Ollama); err != nil {
		return err
	}
	if err := validateProvider("llamacpp", cfg.LlamaCpp); err != nil {
		return err
	}
	if cfg.Concurrency < 1 || cfg.Concurrency > 16 {
		return errors.New("concurrency must be between 1 and 16")
	}
	return nil
}

func validateProvider(name string, p config.Provider) error {
	if !strings.HasPrefix(p.URL, "http://") && !strings.HasPrefix(p.URL, "https://") {
		return errors.New(name + " URL must start with http:// or https://")
	}
	if strings.TrimSpace(p.VisionModel) == "" || strings.TrimSpace(p.ClassifyModel) == "" {
		return errors.New(name + " model names cannot be empty")
	}
	if p.EmbedURL != "" && !strings.HasPrefix(p.EmbedURL, "http://") && !strings.HasPrefix(p.EmbedURL, "https://") {
		return errors.New(name + " embedding URL must start with http:// or https:// (or be empty)")
	}
	return nil
}

type healthResponse struct {
	Online bool   `json:"online"`
	Error  string `json:"error,omitempty"`
}

// probe checks whether the active provider responds. llama-server exposes
// /health; Ollama exposes /api/tags.
func probe(cfg config.Config) healthResponse {
	path := "/health"
	if cfg.Backend == "ollama" {
		path = "/api/tags"
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(cfg.Active().URL + path)
	if err != nil {
		return healthResponse{Online: false, Error: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return healthResponse{Online: false, Error: "HTTP " + resp.Status}
	}
	return healthResponse{Online: true}
}
