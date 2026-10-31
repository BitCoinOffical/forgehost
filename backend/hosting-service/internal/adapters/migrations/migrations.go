package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const (
	dialect = "postgres"
)

//go:embed sql/*.sql
var embedMigrations embed.FS

func RunMigrations(pool *pgxpool.Pool) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(dialect); err != nil {
		return fmt.Errorf("goose.SetDialect: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)
	if err := goose.Up(db, "sql"); err != nil {
		return err
	}
	return nil
}

func RollbackLast(ctx context.Context, db *sql.DB) error {
	if err := goose.DownContext(ctx, db, "sql"); err != nil {
		return err
	}
	return nil
}
