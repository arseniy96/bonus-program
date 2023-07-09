package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/arseniy96/bonus-program/internal/logger"
)

var ErrConflict = errors.New(`already exists`)

type Database struct {
	DB *sqlx.DB
}

func NewStore(dsn string) (*Database, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("migrations failed with error: %w", err)
	}
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	database := &Database{
		DB: db,
	}
	logger.Log.Info("Database connection was created")

	return database, nil
}

func (db *Database) Close() error {
	return db.DB.Close()
}

func runMigrations(dsn string) error {
	const migrationsPath = "../../db/migrations"
	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations: %w", err)
		}
	}
	return nil
}

func (db *Database) CreateUser(ctx context.Context, login, password string) error {
	_, err := db.DB.ExecContext(ctx, `INSERT INTO users(login, password) VALUES($1, $2)`, login, password)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		return ErrConflict
	}

	return err
}

func (db *Database) UpdateUserToken(ctx context.Context, login, token string) error {
	_, err := db.DB.ExecContext(ctx,
		`UPDATE users SET token=$1 WHERE login=$2`,
		token,
		login)
	return err
}
