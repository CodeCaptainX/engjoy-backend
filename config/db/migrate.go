package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
)

func RunMigrations(db *sqlx.DB, migrationsDir string) error {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		upPart := extractUpPart(string(content))
		if upPart == "" {
			continue
		}

		// Split by Statement markers if present, or just run the whole thing
		// Simple approach: just run the whole Up block as one or multiple statements
		if strings.Contains(upPart, "-- +goose StatementBegin") {
			statements := extractStatements(upPart)
			for _, stmt := range statements {
				if strings.TrimSpace(stmt) == "" {
					continue
				}
				if _, err := db.Exec(stmt); err != nil {
					return fmt.Errorf("error in %s: %v", file, err)
				}
			}
		} else {
			if _, err := db.Exec(upPart); err != nil {
				return fmt.Errorf("error in %s: %v", file, err)
			}
		}
	}
	return nil
}

func extractUpPart(content string) string {
	lines := strings.Split(content, "\n")
	var upLines []string
	isUp := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-- +goose Up") {
			isUp = true
			continue
		}
		if strings.HasPrefix(trimmed, "-- +goose Down") {
			break
		}
		if isUp {
			if strings.HasPrefix(trimmed, "-- +goose") && !strings.HasPrefix(trimmed, "-- +goose Statement") {
				continue
			}
			upLines = append(upLines, line)
		}
	}

	return strings.Join(upLines, "\n")
}

func extractStatements(upPart string) []string {
	var statements []string
	lines := strings.Split(upPart, "\n")
	var currentStmt []string
	inStatement := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "-- +goose StatementBegin") {
			inStatement = true
			continue
		}
		if strings.HasPrefix(trimmed, "-- +goose StatementEnd") {
			inStatement = false
			statements = append(statements, strings.Join(currentStmt, "\n"))
			currentStmt = nil
			continue
		}
		if inStatement {
			currentStmt = append(currentStmt, line)
		} else if trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			// Outside statement blocks, we might have simple statements
			// But for simplicity with goose files, we usually expect blocks or simple SQL
			currentStmt = append(currentStmt, line)
		}
	}
	if len(currentStmt) > 0 {
		statements = append(statements, strings.Join(currentStmt, "\n"))
	}
	return statements
}
