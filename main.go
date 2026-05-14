package main

import (
	"context"
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
	if err := db.ApplySchema(context.Background(), dbPool, filepath.Join("config", "db", "schema.sql")); err != nil {
		log.Fatalf("apply schema error: %v", err)
	}

	app := routers.New()

	handler.RegisterRoutes(app, handler.RouteDependencies{
		DB: dbPool,
	})

	addr := fmt.Sprintf("%s:%s", cfg.AppHost, cfg.AppPort)
	log.Printf("listening on http:%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
