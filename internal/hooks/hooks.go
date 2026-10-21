package hooks

import (
	"log"

	"docai/config"
	"docai/internal/queue"
	"docai/internal/routes"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/hook"
)

func Register(app *pocketbase.PocketBase, store *config.ConfigStore, q *queue.Queue) {
	app.OnRecordAfterCreateSuccess("documents").BindFunc(func(e *core.RecordEvent) error {
		q.Enqueue(e.Record.Id)
		q.Tick() // tries to start and if another doc is processing, the id stays queued
		return e.Next()
	})

	app.OnRecordAfterUpdateSuccess("documents").BindFunc(func(e *core.RecordEvent) error {
		// retry_count == 0 means a manual reprocess; the queue's own backoff
		// saves (retry_count > 0) must not re-enqueue immediately
		if e.Record.GetString("status") == "pending" && e.Record.GetInt("retry_count") == 0 {
			q.Enqueue(e.Record.Id)
			q.Tick()
		}
		return e.Next()
	})

	app.OnServe().Bind(&hook.Handler[*core.ServeEvent]{
		Func: func(e *core.ServeEvent) error {
			if err := store.LoadFromDB(app); err != nil {
				log.Printf("failed to load persisted settings: %v", err)
			}

			q.Recover()
			q.Tick()

			app.Cron().Add("docai-queue", "* * * * *", q.Start)

			routes.Register(app, e, store, q)

			return e.Next()
		},
		Priority: 999,
	})
}
