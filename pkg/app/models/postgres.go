package models

import (
	"database/sql"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

func (cfg PostgresConfig) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode)
}

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "host.docker.internal",
		Port:     "5432",
		User:     "sa",
		Password: "@dmin1234",
		Database: "lenslocked",
		SSLMode:  "disable",
	}
}

// Open will open a SQL connection with the provided
// Postgres database. Callers of Open need to ensure
// the connection is eventually closed via the
// db.Close() method.
func Open(config PostgresConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", config.String())
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	return db, nil
}

func Migrate(db *sql.DB, dir string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+dir,
		"lenslocked", driver)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	err = m.Up()
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}

func MigrateFS(db *sql.DB, migrationsFS fs.FS, dir string) error {
	// In case the dir is an empty string, they probably meant the current directory and goose wants a period for that.
	if dir == "" {
		dir = "."
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	source, err := httpfs.New(http.FS(migrationsFS), dir)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	m, err := migrate.NewWithInstance("httpfs", source, "lenslocked", driver)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	err = m.Up()
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
