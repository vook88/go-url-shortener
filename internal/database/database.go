package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"

	errors2 "github.com/vook88/go-url-shortener/internal/errors"
	"github.com/vook88/go-url-shortener/internal/models"
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

func (d *DB) AddURL(ctx context.Context, userID int, id string, url string) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO url_mappings (short_url, long_url, user_id) VALUES ($1, $2, $3)", id, url, userID)
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

func (d *DB) BatchAddURL(ctx context.Context, userID int, urls []InsertURL) error {
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO url_mappings (short_url, long_url, user_id) VALUES ($1, $2, $3)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, url := range urls {
		_, err = stmt.ExecContext(ctx, url.ShortURL, url.OriginalURL, userID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (d *DB) GetURL(ctx context.Context, id string) (string, bool, error) {
	var row struct {
		url       string       `db:"long_url"`
		deletedAt sql.NullTime `db:"deleted_at"`
	}
	err := d.db.QueryRowContext(ctx, "SELECT long_url, deleted_at FROM url_mappings WHERE short_url = $1", id).Scan(&row.url, &row.deletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	if row.deletedAt.Valid {
		return "", false, errors2.ErrURLDeleted
	}
	return row.url, true, nil
}

func (d *DB) AddUser(ctx context.Context) (int, error) {
	var lastInsetID int64
	row := d.db.QueryRowContext(ctx, "INSERT INTO users DEFAULT VALUES RETURNING id").Scan(&lastInsetID)
	if row != nil && row.Error() != "" {
		return 0, errors.New(row.Error())
	}
	return int(lastInsetID), nil
}

func (d *DB) GetUserURLs(ctx context.Context, userID int) (models.BatchUserURLs, error) {
	rows, err := d.db.QueryContext(ctx, "SELECT long_url as original_url, short_url FROM url_mappings WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()
	var urls models.BatchUserURLs
	for rows.Next() {
		var url models.UserURL
		err = rows.Scan(&url.OriginalURL, &url.ShortURL)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func (d *DB) BatchDeleteURLs(ctx context.Context, urls []string) error {
	if len(urls) == 0 {
		return nil // Нет URL для удаления
	}

	placeholders := make([]string, len(urls))
	args := make([]interface{}, len(urls))
	for i, url := range urls {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = url
	}
	query := fmt.Sprintf("UPDATE url_mappings SET deleted_at = NOW() WHERE short_url IN (%s)", strings.Join(placeholders, ","))

	// Выполняем запрос
	_, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
