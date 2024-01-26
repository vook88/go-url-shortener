package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(databaseDNS string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseDNS)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Ping(ctx context.Context, db *sql.DB) error {
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func RunMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/database/migrations",
		"mydatabase", driver)

	if err != nil {
		return err
	}

	err = m.Up()
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return nil
	}

	return err
}

func AddURL(ctx context.Context, db *sql.DB, id string, url string) error {
	_, err := db.ExecContext(ctx, "INSERT INTO url_mappings (short_url, long_url) VALUES ($1, $2)", id, url)
	return err
}

func GetURL(ctx context.Context, db *sql.DB, id string) (string, bool, error) {
	var url string
	err := db.QueryRowContext(ctx, "SELECT long_url FROM url_mappings WHERE short_url = $1", id).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return url, true, nil
}
