package queue

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"docai/config"
	"docai/internal/pipeline"

	"github.com/pocketbase/pocketbase/core"
)

// ponytail: retry cap; tune if the provider is down for days
const maxRetries = 5

// Queue is an in-memory FIFO with bounded concurrency. Records are re-fetched
// from the DB when dequeued, so a stale id is simply skipped. Metadata jobs
// are always dequeued before embed jobs, so uploads never wait behind a
// running backfill.
type Queue struct {
	app      core.App
	store    *config.ConfigStore
	mu       sync.Mutex
	ids      []string
	embedIDs []string
	active   int
}

func New(app core.App, store *config.ConfigStore) *Queue {
	return &Queue{app: app, store: store}
}

func (q *Queue) Enqueue(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if contains(q.ids, id) {
		return
	}
	q.ids = append(q.ids, id)
}

// EnqueueEmbed schedules the separate embed step for a document. It persists
// embed_status=pending (visible to realtime UI and boot recovery), clears any
// previous embed error, and resets the retry counter for a fresh budget.
func (q *Queue) EnqueueEmbed(id string) {
	record, err := q.app.FindRecordById("documents", id)
	if err != nil {
		slog.Warn("queue: embed record gone", "id", id, "error", err)
		return
	}
	record.Set("embed_status", "pending")
	record.Set("embed_error", "")
	record.Set("retry_count", 0)
	if err := q.app.Save(record); err != nil {
		slog.Error("queue: failed to persist embed pending", "id", id, "error", err)
		return
	}

	q.mu.Lock()
	if !contains(q.embedIDs, id) {
		q.embedIDs = append(q.embedIDs, id)
	}
	q.mu.Unlock()
	q.Tick()
}

// enqueueEmbedRetry queues an existing embed job without touching the record,
// preserving the retry counter for backoff.
func (q *Queue) enqueueEmbedRetry(id string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if !contains(q.embedIDs, id) {
		q.embedIDs = append(q.embedIDs, id)
	}
}

func contains(ids []string, id string) bool {
	for _, existing := range ids {
		if existing == id {
			return true
		}
	}
	return false
}

// Recover re-queues records at boot: processing records immediately (they
// were interrupted mid-flight), pending records only when their backoff has
// elapsed. The same applies to the embed step.
func (q *Queue) Recover() {
	records, err := q.app.FindRecordsByFilter("documents", "status = 'pending' || status = 'processing'", "-created", 0, 0)
	if err != nil {
		slog.Error("queue recover failed", "error", err)
		return
	}
	now := time.Now()
	for _, record := range records {
		if record.GetString("status") == "pending" && !q.retryDue(record, now) {
			continue
		}
		q.Enqueue(record.Id)
	}

	embeds, err := q.app.FindRecordsByFilter("documents", "embed_status = 'pending' || embed_status = 'processing'", "-created", 0, 0)
	if err != nil {
		slog.Error("queue embed recover failed", "error", err)
		return
	}
	now = time.Now()
	for _, record := range embeds {
		if record.GetString("embed_status") == "pending" && !q.retryDue(record, now) {
			continue
		}
		q.enqueueEmbedRetry(record.Id)
	}

	slog.Info("queue recovered", "count", len(records), "embeds", len(embeds))
}

// RequeuePending re-queues pending records whose retry backoff has elapsed.
// Called by the cron tick; processing records are left alone (they are
// already in flight).
func (q *Queue) RequeuePending() {
	records, err := q.app.FindRecordsByFilter("documents", "status = 'pending'", "-created", 0, 0)
	if err != nil {
		slog.Error("queue requeue failed", "error", err)
		return
	}
	now := time.Now()
	for _, record := range records {
		if !q.retryDue(record, now) {
			continue
		}
		q.Enqueue(record.Id)
	}
}

// RequeueEmbeds re-queues embed jobs whose retry backoff has elapsed.
// Interrupted processing jobs are handled by Recover at boot.
func (q *Queue) RequeueEmbeds() {
	records, err := q.app.FindRecordsByFilter("documents", "embed_status = 'pending'", "-created", 0, 0)
	if err != nil {
		slog.Error("queue embed requeue failed", "error", err)
		return
	}
	now := time.Now()
	for _, record := range records {
		if !q.retryDue(record, now) {
			continue
		}
		q.enqueueEmbedRetry(record.Id)
	}
}

func (q *Queue) Start() {
	q.RequeuePending()
	q.RequeueEmbeds()
	q.Tick()
}

// retryDue reports whether a pending record is ready for another attempt:
// exponential backoff (1m, 2m, 4m, 8m, 16m) on retry_count, no cap. Records
// past maxRetries are left for the user to reprocess.
func (q *Queue) retryDue(record *core.Record, now time.Time) bool {
	retries := record.GetInt("retry_count")
	if retries == 0 {
		return true
	}
	if retries > maxRetries {
		return false
	}
	delay := time.Duration(1<<(retries-1)) * time.Minute
	return now.Sub(record.GetDateTime("updated").Time()) >= delay
}

// Tick starts queued jobs while below the concurrency limit. Metadata jobs
// are picked first; embed jobs use the same limit. Safe to call from cron and
// upload hooks.
func (q *Queue) Tick() {
	q.mu.Lock()
	defer q.mu.Unlock()
	for q.active < q.store.Get().Concurrency {
		var id string
		embed := false
		switch {
		case len(q.ids) > 0:
			id, q.ids = q.ids[0], q.ids[1:]
		case len(q.embedIDs) > 0:
			id, q.embedIDs, embed = q.embedIDs[0], q.embedIDs[1:], true
		default:
			return
		}
		q.active++
		go q.run(id, embed)
	}
}

func (q *Queue) run(id string, embed bool) {
	defer func() {
		q.mu.Lock()
		q.active--
		q.mu.Unlock()
		q.Tick()
	}()
	if embed {
		q.processEmbed(id)
		return
	}
	q.process(id)
}

// process runs the regular pipeline (extract, OCR, classify) and persists the
// result. Embedding is not part of it.
func (q *Queue) process(id string) {
	record, err := q.app.FindRecordById("documents", id)
	if err != nil {
		slog.Warn("queue: record gone", "id", id, "error", err)
		return
	}

	if record.GetString("status") != "pending" && record.GetString("status") != "processing" {
		return
	}

	record.Set("status", "processing")
	if err := q.app.Save(record); err != nil {
		slog.Error("queue: failed to mark processing", "id", id, "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*60*60*1000*1000*1000) // 2h
	defer cancel()

	onProgress := func(p pipeline.Progress) {
		record.Set("progress", p)
		if err := q.app.Save(record); err != nil {
			slog.Warn("queue: failed to persist progress", "id", id, "error", err)
		}
	}

	cfg := q.store.Get()
	result, err := pipeline.Process(ctx, q.app, record, cfg, onProgress)
	if err != nil {
		slog.Warn("queue: processing failed", "id", id, "error", err)
		if isConnectivityError(err) {
			retries := record.GetInt("retry_count")
			if retries >= maxRetries {
				record.Set("status", "error")
				record.Set("error", "AI provider offline after "+strconv.Itoa(maxRetries)+" retries")
			} else {
				// provider offline: back off and retry via cron
				record.Set("status", "pending")
				record.Set("retry_count", retries+1)
				record.Set("error", "AI provider offline, will retry")
			}
		} else {
			record.Set("status", "error")
			record.Set("error", err.Error())
		}
		// the previous embedding (if any) is kept until an explicit re-embed
	} else {
		record.Set("status", "done")
		record.Set("error", "")
		record.Set("retry_count", 0)
		record.Set("ocr_text", result.Text)
		record.Set("model", result.Model)
		record.Set("timings", result.Timings)
		if result.Metadata.Title != "" || result.Metadata.Summary != "" {
			record.Set("title", result.Metadata.Title)
			record.Set("summary", result.Metadata.Summary)
			record.Set("category", result.Metadata.Category)
			record.Set("tags", result.Metadata.Tags)
			record.Set("entities", result.Metadata.Entities)
			record.Set("language", result.Metadata.Language)
		}
	}

	if err := q.app.Save(record); err != nil {
		slog.Error("queue: failed to persist result", "id", id, "error", err)
	}
	record.Set("progress", nil)
	if err := q.app.Save(record); err != nil {
		slog.Warn("queue: failed to clear progress", "id", id, "error", err)
	}

	if err == nil && cfg.Active().EmbedModel != "" {
		q.EnqueueEmbed(id)
	}
}

// processEmbed is the separate embed step: it embeds the stored text and
// upserts the embeddings row. Failures retry with the usual backoff when the
// provider is unreachable; other failures are terminal for the document.
func (q *Queue) processEmbed(id string) {
	record, err := q.app.FindRecordById("documents", id)
	if err != nil {
		slog.Warn("queue: embed record gone", "id", id, "error", err)
		return
	}

	cfg := q.store.Get()
	model := cfg.Active().EmbedModel

	reset := func(status, msg string) {
		record.Set("embed_status", status)
		record.Set("embed_error", msg)
		if err := q.app.Save(record); err != nil {
			slog.Warn("queue: failed to persist embed state", "id", id, "error", err)
		}
	}

	// nothing to embed (model removed mid-queue, or the document is being
	// reprocessed); a completed reprocess enqueues the embed again
	if model == "" || record.GetString("status") != "done" || strings.TrimSpace(record.GetString("ocr_text")) == "" {
		reset("", "")
		return
	}

	record.Set("embed_status", "processing")
	record.Set("progress", pipeline.Progress{Stage: "embed"})
	if err := q.app.Save(record); err != nil {
		slog.Error("queue: failed to mark embed processing", "id", id, "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*60*60*1000*1000*1000) // 2h
	defer cancel()

	vec, err := pipeline.EmbedText(ctx, cfg, record.GetString("title"), record.GetString("ocr_text"))
	switch {
	case err != nil:
		slog.Warn("queue: embedding failed", "id", id, "error", err)
		if isConnectivityError(err) {
			retries := record.GetInt("retry_count")
			if retries >= maxRetries {
				reset("error", "AI provider offline after "+strconv.Itoa(maxRetries)+" retries")
			} else {
				record.Set("retry_count", retries+1)
				reset("pending", "AI provider offline, will retry")
			}
		} else {
			reset("error", err.Error())
		}
	case vec == nil:
		reset("", "")
	default:
		if err := pipeline.StoreEmbedding(q.app, id, model, vec); err != nil {
			slog.Warn("queue: failed to store embedding", "id", id, "error", err)
			reset("error", "failed to store embedding")
			break
		}
		record.Set("retry_count", 0)
		reset("done", "")
	}

	record.Set("progress", nil)
	if err := q.app.Save(record); err != nil {
		slog.Warn("queue: failed to clear embed progress", "id", id, "error", err)
	}
}

// isConnectivityError reports whether err is a network-level failure (provider
// unreachable), as opposed to a provider response error.
func isConnectivityError(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr)
}
