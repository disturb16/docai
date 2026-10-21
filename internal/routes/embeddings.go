package routes

import (
	"docai/config"
	"docai/internal/queue"

	"github.com/pocketbase/pocketbase/core"
)

func registerEmbeddings(app core.App, e *core.ServeEvent, store *config.ConfigStore, q *queue.Queue) {
	e.Router.POST("/api/documents/{id}/embed", func(e *core.RequestEvent) error {
		cfg := store.Get()
		if cfg.Active().EmbedModel == "" {
			return e.BadRequestError("no embedding model configured", nil)
		}
		id := e.Request.PathValue("id")
		record, err := app.FindRecordById("documents", id)
		if err != nil {
			return e.NotFoundError("document not found", err)
		}
		if record.GetString("status") != "done" || record.GetString("ocr_text") == "" {
			return e.BadRequestError("document has no extracted text to embed", nil)
		}
		q.EnqueueEmbed(id)
		return e.JSON(200, map[string]string{"status": "queued"})
	})

	e.Router.POST("/api/docai/embeddings/backfill", func(e *core.RequestEvent) error {
		cfg := store.Get()
		model := cfg.Active().EmbedModel
		if model == "" {
			return e.BadRequestError("no embedding model configured", nil)
		}

		records, err := app.FindRecordsByFilter("documents", "status = 'done'", "-created", 0, 0)
		if err != nil {
			return e.InternalServerError("failed to fetch documents", err)
		}
		embeddings, err := app.FindAllRecords("embeddings")
		if err != nil {
			return e.InternalServerError("failed to fetch embeddings", err)
		}
		byDoc := make(map[string]*core.Record, len(embeddings))
		for _, emb := range embeddings {
			byDoc[emb.GetString("document")] = emb
		}

		queued, skipped := 0, 0
		for _, r := range records {
			existing := byDoc[r.Id]
			hasVec := existing != nil && len(toVector(existing.Get("vector"))) > 0
			if (hasVec && existing.GetString("model") == model) || r.GetString("ocr_text") == "" {
				skipped++
				continue
			}
			if s := r.GetString("embed_status"); s == "pending" || s == "processing" {
				skipped++
				continue
			}
			q.EnqueueEmbed(r.Id)
			queued++
		}

		return e.JSON(200, map[string]int{"queued": queued, "skipped": skipped})
	})
}
