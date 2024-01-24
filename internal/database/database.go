package database

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func New(databaseDNS string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseDNS)
	if err != nil {
		return nil, err
	}
	return db, nil
}
