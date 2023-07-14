package store

import (
	"context"
	"database/sql"
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
var ErrInvalidLogin = errors.New(`invalid login`)

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

func (db *Database) FindUserByLogin(ctx context.Context, login string) (*User, error) {
	row := db.DB.QueryRowContext(ctx,
		`SELECT id, login, password, token, bonuses FROM users WHERE login=$1 LIMIT(1)`,
		login)
	var u User
	err := row.Scan(&u.ID, &u.Login, &u.Password, &u.Token, &u.Bonuses)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidLogin
		}
		return nil, err
	}

	return &u, nil
}

func (db *Database) FindUserByToken(ctx context.Context, token string) (*User, error) {
	row := db.DB.QueryRowContext(ctx,
		`SELECT id, login, password, token, bonuses FROM users WHERE token=$1 LIMIT(1)`,
		token)
	var u User
	err := row.Scan(&u.ID, &u.Login, &u.Password, &u.Token, &u.Bonuses)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidLogin
		}
		return nil, err
	}

	return &u, nil
}

func (db *Database) FindOrdersByUserID(ctx context.Context, userID int) ([]Order, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT id, order_number, status, user_id, created_at FROM orders WHERE user_id=$1`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err = rows.Scan(&order.ID, &order.OrderNumber, &order.Status, &order.UserID, &order.CreatedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (db *Database) FindBonusTransactionsByUserID(ctx context.Context, userID int) ([]BonusTransaction, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT id, amount, type, user_id, order_id, created_at FROM bonus_transactions WHERE user_id=$1`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []BonusTransaction
	for rows.Next() {
		var tr BonusTransaction
		err = rows.Scan(&tr.ID, &tr.Amount, &tr.Type, &tr.UserID, &tr.OrderID, &tr.CreatedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tr)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

func (db *Database) GetWithdrawalSumByUserID(ctx context.Context, userID int) (int, error) {
	row := db.DB.QueryRowContext(ctx,
		`SELECT SUM(amount) AS total FROM bonus_transactions WHERE user_id=$1 AND type='withdrawal'`,
		userID)
	var total int
	err := row.Scan(&total)

	return total, err
}
