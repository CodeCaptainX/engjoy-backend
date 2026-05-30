package main

import (
	"fmt"
	"log"
	"path/filepath"

	config "sentenceminer/config"
	"sentenceminer/config/database"
	"sentenceminer/config/db"
	"sentenceminer/handler"
	"sentenceminer/internal/sentences/repository"
	"sentenceminer/routers"
)

func main() {
	cfg := config.NewConfig()

	dbPool := database.Connect(cfg.DatabaseURL)
	if err := db.RunMigrations(dbPool, filepath.Join("config", "db", "migrations")); err != nil {
		log.Fatalf("migrations error: %v", err)
	}

	// Initialize database with static sentence packs if empty
	repo := repository.NewSentenceRepository(dbPool)
	count, _ := repo.CountSentences()
	if count == 0 {
		log.Println("database empty, importing static sentence packs...")
		if _, err := repo.ImportStaticSentencePacks(); err != nil {
			log.Printf("failed to import static packs: %v", err)
		} else {
			log.Println("static sentence packs imported successfully")
		}
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
