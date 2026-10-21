package main

import (
	"log"

	"docai/config"
	"docai/internal/hooks"
	"docai/internal/queue"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/plugins/jsvm"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

func main() {
	app := pocketbase.New()

	jsvm.MustRegister(app, jsvm.Config{
		MigrationsDir: "pb_migrations",
	})

	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		TemplateLang: migratecmd.TemplateLangJS,
		Automigrate:  true,
		Dir:          "pb_migrations",
	})

	store := config.NewStore()
	q := queue.New(app, store)
	hooks.Register(app, store, q)

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
