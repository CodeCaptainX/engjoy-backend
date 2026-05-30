package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	config "sentenceminer/config"
	"sentenceminer/config/database"
	"sentenceminer/config/db"
	"sentenceminer/handler"
	"sentenceminer/internal/sentences/repository"
	"sentenceminer/internal/sentences/service"
	"sentenceminer/routers"
)

func main() {
	generateFlag := flag.Bool("generate", false, "run sentence generation for all categories and exit")
	flag.Parse()

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

	if *generateFlag {
		log.Println("CLI mode: starting sentence generation...")
		sentenceSvc := service.NewSentenceService(dbPool)
		
		categories := []string{
			"daily-life", "travel", "airport", "restaurant", "hospital",
			"banking", "job-interview", "office", "shopping", "tech-support",
			"school", "sports", "phone-call", "emergency", "renting",
			"general", "deep-sea-exploration", "space-travel",
		}

		for _, cat := range categories {
			log.Printf("Generating for category: %s...", cat)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			_, err := sentenceSvc.GenerateAndSaveSentences(ctx, cat, "useful expressions", 10)
			cancel()
			if err != nil {
				log.Printf("Generation failed for %s: %v", cat, err)
			}
			time.Sleep(2 * time.Second)
		}
		log.Println("Generation complete. Exiting.")
		return
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
