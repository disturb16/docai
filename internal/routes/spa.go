package routes

import (
	"os"
	"path/filepath"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/osutils"
)

// registerSPA serves the built frontend for all unmatched GET routes.
func registerSPA(e *core.ServeEvent) {
	if e.Router.HasRoute("GET", "/{path...}") {
		return
	}
	e.Router.GET("/{path...}", apis.Static(os.DirFS(publicDir()), true))
}

func publicDir() string {
	if osutils.IsProbablyGoRun() {
		return "./frontend/pb_public"
	}
	return filepath.Join(filepath.Dir(os.Args[0]), "frontend", "pb_public")
}
