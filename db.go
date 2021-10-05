package main

import (
	"database/sql"
	"embed"
	"fmt"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

//NewDb tries to open the sqlite database
func NewDB(dbPath string) (*sql.DB, error) {
	sqliteDb, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open sqlite DB")
	}
	return sqliteDb, nil
}

//go:embed migrations
var migrations embed.FS

const schemaVersion = 1

func EnsureSchema(db *sql.DB) error {
	sourceInstance, err := httpfs.New(http.FS(migrations), "migrations")
	if err != nil {
		return fmt.Errorf("invalid source instance, %w", err)
	}
	targetInstance, err := sqlite.WithInstance(db, new(sqlite.Config))
	if err != nil {
		return fmt.Errorf("invalid target sqlite instance, %w", err)
	}
	m, err := migrate.NewWithInstance(
		"httpfs", sourceInstance, "sqlite", targetInstance)
	if err != nil {
		return fmt.Errorf("failed to initialize migrate instance, %w", err)
	}
	err = m.Migrate(schemaVersion)
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return sourceInstance.Close()
}
