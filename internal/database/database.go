package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	errors2 "github.com/vook88/go-url-shortener/internal/errors"
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

func (d *DB) getShortURLByLongURL(ctx context.Context, id string) (string, bool, error) {
	var shortURL string
	err := d.db.QueryRowContext(ctx, "SELECT short_url FROM url_mappings WHERE long_url = $1", id).Scan(&shortURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return shortURL, true, nil
}

func (d *DB) AddURL(ctx context.Context, id string, url string) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO url_mappings (short_url, long_url) VALUES ($1, $2)", id, url)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == pgerrcode.UniqueViolation {
			shortURL, _, err2 := d.getShortURLByLongURL(ctx, url)
			if err2 != nil {
				return err2
			}
			return errors2.NewDuplicateURLError(shortURL)
		}
	}
	return err
}

type InsertURL struct {
	ShortURL    string
	OriginalURL string
}

func (d *DB) BatchAddURL(ctx context.Context, urls []InsertURL) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO url_mappings (short_url, long_url) VALUES ($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, url := range urls {
		_, err = stmt.ExecContext(ctx, url.ShortURL, url.OriginalURL)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
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
