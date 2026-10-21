// Package routes holds all custom HTTP routes: search, settings, and the SPA
// catch-all. Everything else is served by PocketBase's built-in API.
package routes

import (
	"docai/config"
	"docai/internal/queue"

	"github.com/pocketbase/pocketbase/core"
)

func Register(app core.App, e *core.ServeEvent, store *config.ConfigStore, q *queue.Queue) {
	registerSearch(app, e, store)
	registerSettings(app, e, store)
	registerAnalytics(app, e)
	registerEmbeddings(app, e, store, q)
	registerSPA(e)
}
