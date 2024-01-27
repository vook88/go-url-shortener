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

type DB struct {
	db *sql.DB
}

func New(databaseDNS string) (*DB, error) {
	db, err := sql.Open("pgx", databaseDNS)
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (d *DB) Ping(ctx context.Context) error {
	if err := d.db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}

func (d *DB) RunMigrations() error {
	driver, err := postgres.WithInstance(d.db, &postgres.Config{})
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

func (d *DB) AddURL(ctx context.Context, id string, url string) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO url_mappings (short_url, long_url) VALUES ($1, $2)", id, url)
	return err
}

func (d *DB) GetURL(ctx context.Context, id string) (string, bool, error) {
	var url string
	err := d.db.QueryRowContext(ctx, "SELECT long_url FROM url_mappings WHERE short_url = $1", id).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return url, true, nil
}
