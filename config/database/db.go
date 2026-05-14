package database

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(databaseURL string) *sqlx.DB {
	db, err := sqlx.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("db open error: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("db ping error: %v", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	return db
}
