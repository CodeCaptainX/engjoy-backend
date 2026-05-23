package main

import (
	"fmt"
	"log"
	"path/filepath"

	config "sentenceminer/config"
	"sentenceminer/config/database"
	"sentenceminer/config/db"
	"sentenceminer/handler"
	"sentenceminer/routers"
)

func main() {
	cfg := config.NewConfig()

	dbPool := database.Connect(cfg.DatabaseURL)
	if err := db.RunMigrations(dbPool, filepath.Join("config", "db", "migrations")); err != nil {
		log.Fatalf("migrations error: %v", err)
	}

	app := routers.New()

	handler.RegisterRoutes(app, handler.RouteDependencies{
		DB:     dbPool,
		Config: cfg,
	})

	addr := fmt.Sprintf("%s:%s", cfg.AppHost, cfg.AppPort)
	log.Printf("listening on http:%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
