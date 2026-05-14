package db

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"
)

func ApplySchema(ctx context.Context, db *sqlx.DB, schemaPath string) error {
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return err
	}

	statements := splitSQL(string(content))
	for _, stmt := range statements {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func splitSQL(sql string) []string {
	var out []string
	var buf bytes.Buffer
	for _, r := range sql {
		if r == ';' {
			out = append(out, buf.String())
			buf.Reset()
			continue
		}
		buf.WriteRune(r)
	}
	if buf.Len() > 0 {
		out = append(out, buf.String())
	}
	return out
}
