package pipeline

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"docai/config"
	"docai/internal/classify"

	"github.com/pocketbase/pocketbase/core"
)

// ponytail: hard cap on OCR pages; raise if scanned docs >10 pages matter
const maxOCRPagess = 10

type Result struct {
	Text     string
	Metadata classify.Metadata
	Model    string
	// Timings holds per-stage durations in milliseconds: extract, ocr,
	// classify. Stages not run for a document are absent.
	Timings map[string]int64
}

// Progress is a pipeline stage report persisted by the caller.
type Progress struct {
	Stage   string `json:"stage"`
	Current int    `json:"current,omitempty"`
	Total   int    `json:"total,omitempty"`
}

// Process runs the regular pipeline for a record: extract text, OCR if
// needed, and classify. Embedding is a separate, on-demand step (see
// EmbedText). The caller persists the result. onProgress (may be nil) reports
// stage changes as they happen.
func Process(ctx context.Context, app core.App, record *core.Record, cfg config.Config, onProgress func(Progress)) (Result, error) {
	file, err := openFile(app, record)
	if err != nil {
		return Result{}, err
	}
	defer file.Close()

	mime := record.GetString("mime")
	id := record.Id

	timings := map[string]int64{}
	text, err := extractText(ctx, file, mime, cfg, id, onProgress, timings)
	if err != nil {
		slog.Warn("pipeline: extract failed", "id", id, "mime", mime, "error", err)
		return Result{}, err
	}

	if onProgress != nil {
		onProgress(Progress{Stage: "classify"})
	}
	classifyStart := time.Now()
	meta, err := classifyText(ctx, text, cfg, id)
	timings["classify"] = msSince(classifyStart)
	if err != nil {
		// provider offline: propagate so the queue retries later
		var netErr net.Error
		if errors.As(err, &netErr) {
			return Result{}, err
		}
		// fallback: keep the raw text, no metadata
		return Result{Text: text, Model: "raw", Timings: timings}, nil
	}

	return Result{Text: text, Metadata: meta, Model: cfg.Backend, Timings: timings}, nil
}

type namedFile interface {
	io.ReadSeekCloser
	Name() string
}

func openFile(app core.App, record *core.Record) (namedFile, error) {
	filenames := record.GetStringSlice("file")
	if len(filenames) == 0 {
		return nil, errors.New("no file attached")
	}

	fsys, err := app.NewFilesystem()
	if err != nil {
		return nil, err
	}
	defer fsys.Close()

	reader, err := fsys.GetReader(filepath.Join(record.BaseFilesPath(), filenames[0]))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	tmp, err := os.CreateTemp("", "docai-*"+filepath.Ext(filenames[0]))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(tmp, reader); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, err
	}
	if _, err := tmp.Seek(0, 0); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return nil, err
	}

	return &autoRemoveFile{File: tmp}, nil
}

type autoRemoveFile struct {
	*os.File
}

func (f *autoRemoveFile) Close() error {
	err := f.File.Close()
	os.Remove(f.Name())
	return err
}

func extractText(ctx context.Context, file namedFile, mime string, cfg config.Config, id string, onProgress func(Progress), timings map[string]int64) (string, error) {
	switch {
	case mime == "text/plain" || mime == "text/markdown" || mime == "text/x-markdown":
		data, err := io.ReadAll(file)
		if err != nil {
			return "", err
		}
		return string(data), nil

	case mime == "image/png" || mime == "image/jpeg":
		if onProgress != nil {
			onProgress(Progress{Stage: "ocr", Current: 1, Total: 1})
		}
		start := time.Now()
		text, err := visionFile(ctx, file, cfg)
		timings["ocr"] = msSince(start)
		if err != nil {
			slog.Warn("pipeline: vision failed", "id", id, "mime", mime, "error", err)
		}
		return text, err

	default: // pdf, docx, odt, xlsx
		if onProgress != nil {
			onProgress(Progress{Stage: "extract"})
		}
		start := time.Now()
		text, err := litParse(ctx, file.Name())
		timings["extract"] = msSince(start)
		if err == nil && len(strings.TrimSpace(text)) >= 100 {
			return text, nil
		}
		slog.Warn("pipeline: lit parse produced no text (likely scanned), falling back to vision",
			"id", id, "chars", len(strings.TrimSpace(text)), "error", err)

		ocrStart := time.Now()
		text, err = visionPDF(ctx, file.Name(), cfg, id, onProgress)
		timings["ocr"] = msSince(ocrStart)
		return text, err
	}
}

func msSince(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}

func litParse(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "lit", "parse", "--no-ocr", "--max-pages", "100", "-q", path)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func visionPDF(ctx context.Context, path string, cfg config.Config, id string, onProgress func(Progress)) (string, error) {
	dir, err := os.MkdirTemp("", "docai-pages-*")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(dir)

	cmd := exec.CommandContext(ctx, "lit", "screenshot", "--target-pages", fmt.Sprintf("1-%d", maxOCRPagess), "-q", "-o", dir, path)
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		slog.Error("pipeline: screenshot render failed", "id", id, "error", err)
		return "", err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var pages []string
	for _, entry := range entries {
		if !entry.IsDir() {
			pages = append(pages, entry.Name())
		}
	}

	var parts []string
	for i, name := range pages {
		if onProgress != nil {
			onProgress(Progress{Stage: "ocr", Current: i + 1, Total: len(pages)})
		}
		page, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		text, err := visionFile(ctx, page, cfg)
		page.Close()
		if err != nil {
			// connectivity errors abort the whole doc so the queue can retry
			// it later; other errors skip the page
			var netErr net.Error
			if errors.As(err, &netErr) {
				slog.Warn("pipeline: vision connectivity error, aborting doc", "id", id, "page", i+1, "error", err)
				return "", err
			}
			slog.Warn("pipeline: vision failed, skipping page", "id", id, "page", i+1, "error", err)
			continue
		}
		parts = append(parts, text)
	}

	if len(parts) == 0 {
		slog.Error("pipeline: no pages rendered", "id", id)
		return "", errors.New("no pages rendered")
	}
	return strings.Join(parts, "\n\n"), nil
}

func visionFile(ctx context.Context, file io.ReadSeekCloser, cfg config.Config) (string, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	return generate(ctx, cfg, cfg.Active().VisionModel, "transcribe all text from this image. Output only the text.", data)
}

func classifyText(ctx context.Context, text string, cfg config.Config, id string) (classify.Metadata, error) {
	prompt := `You are a document classifier. Analyze the following document text and output ONLY a strict JSON object with these keys:
{"title": string, "summary": string, "category": string, "tags": [string], "entities": [string], "language": string}

Detect the language of the document. Write title, summary, category, tags, and entities in the SAME language as the document. Set "language" to the ISO 639-1 code of that language.

Document text:
` + text

	model := cfg.Active().ClassifyModel
	raw, err := generate(ctx, cfg, model, prompt, nil)
	if err != nil {
		slog.Warn("pipeline: classify generate failed", "id", id, "model", model, "error", err)
		return classify.Metadata{}, err
	}
	meta, err := classify.Parse(raw)
	if err != nil {
		slog.Warn("pipeline: classify parse failed, using raw text", "id", id, "model", model, "response", snippet(raw))
		return classify.Metadata{}, err
	}
	return meta, nil
}

// baseURL normalizes a provider URL: trailing slash and an OpenAI-style /v1
// suffix are stripped, since callers append their own API path.
func baseURL(u string) string {
	u = strings.TrimRight(u, "/")
	u = strings.TrimSuffix(u, "/v1")
	return strings.TrimRight(u, "/")
}

// snippet truncates a model response for logging.
func snippet(s string) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > 300 {
		return string(r[:300]) + "..."
	}
	return s
}

func generate(ctx context.Context, cfg config.Config, model, prompt string, image []byte) (string, error) {
	if cfg.Backend == "llamacpp" {
		return llamaGenerate(ctx, cfg, model, prompt, image)
	}
	return ollamaGenerate(ctx, cfg, model, prompt, image)
}

type ollamaRequest struct {
	Model     string   `json:"model"`
	Prompt    string   `json:"prompt"`
	Images    []string `json:"images,omitempty"`
	Stream    bool     `json:"stream"`
	KeepAlive string   `json:"keep_alive"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

func ollamaGenerate(ctx context.Context, cfg config.Config, model, prompt string, image []byte) (string, error) {
	var images []string
	if image != nil {
		images = []string{base64.StdEncoding.EncodeToString(image)}
	}
	body, err := json.Marshal(ollamaRequest{
		Model:     model,
		Prompt:    prompt,
		Images:    images,
		Stream:    false,
		KeepAlive: "5m",
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL(cfg.Ollama.URL)+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("ollama generate http %d: %s", resp.StatusCode, snippet(string(data)))
	}

	var out ollamaResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	if out.Error != "" {
		return "", errors.New(out.Error)
	}
	return out.Response, nil
}

type llamaMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type llamaRequest struct {
	Model    string         `json:"model"`
	Messages []llamaMessage `json:"messages"`
	Stream   bool           `json:"stream"`
}

type llamaResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func llamaGenerate(ctx context.Context, cfg config.Config, model, prompt string, image []byte) (string, error) {
	var content interface{} = prompt
	if image != nil {
		mime := http.DetectContentType(image)
		content = []map[string]interface{}{
			{"type": "text", "text": prompt},
			{"type": "image_url", "image_url": map[string]string{
				"url": "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(image),
			}},
		}
	}

	body, err := json.Marshal(llamaRequest{
		Model:    model,
		Messages: []llamaMessage{{Role: "user", Content: content}},
		Stream:   false,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL(cfg.LlamaCpp.URL)+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("llama generate http %d: %s", resp.StatusCode, snippet(string(data)))
	}

	var out llamaResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return "", err
	}
	if out.Error.Message != "" {
		return "", errors.New(out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return "", errors.New("no choices in response")
	}
	return out.Choices[0].Message.Content, nil
}
