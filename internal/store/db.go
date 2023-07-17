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
	const migrationsPath = "db/migrations"
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
	var u User
	err := db.DB.QueryRowContext(ctx,
		`SELECT id, login, password, token, bonuses FROM users WHERE login=$1 LIMIT(1)`,
		login).Scan(&u.ID, &u.Login, &u.Password, &u.Token, &u.Bonuses)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidLogin
		}
		return nil, err
	}

	return &u, nil
}

func (db *Database) FindUserByToken(ctx context.Context, token string) (*User, error) {
	var u User
	err := db.DB.QueryRowContext(ctx,
		`SELECT id, login, password, token, bonuses FROM users WHERE token=$1 LIMIT(1)`,
		token).Scan(&u.ID, &u.Login, &u.Password, &u.Token, &u.Bonuses)
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
		`SELECT o.id, o.order_number, o.status, o.user_id, o.created_at, bt.amount FROM orders o JOIN bonus_transactions bt ON o.id = bt.order_id WHERE o.user_id=$1 AND bt.type=$2`,
		userID,
		AccrualType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err = rows.Scan(&order.ID, &order.OrderNumber, &order.Status, &order.UserID, &order.CreatedAt, &order.BonusAmount)
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
		`SELECT b.id, b.amount, b.type, b.user_id, b.order_id, o.order_number, b.created_at FROM bonus_transactions b JOIN orders o ON b.order_id=o.id WHERE b.user_id=$1`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []BonusTransaction
	for rows.Next() {
		var tr BonusTransaction
		err = rows.Scan(&tr.ID, &tr.Amount, &tr.Type, &tr.UserID, &tr.OrderID, &tr.OrderNumber, &tr.CreatedAt)
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
	var total int
	err := db.DB.QueryRowContext(ctx,
		`SELECT SUM(amount) AS total FROM bonus_transactions WHERE user_id=$1 AND type=$2`,
		userID, WithdrawalType).Scan(&total)

	return total, err
}

func (db *Database) SaveWithdrawBonuses(ctx context.Context, userID int, orderNumber string, sum float64) error {
	amount := int(sum * 100)

	// TODO: обернуть всё в транзакцию
	var userBonuses int
	err := db.DB.QueryRowContext(ctx,
		`SELECT bonuses FROM users WHERE id=$1`,
		userID).Scan(&userBonuses)
	if err != nil {
		return err
	}

	var orderID int
	err = db.DB.QueryRowContext(ctx,
		`INSERT INTO orders(order_number, status, user_id) VALUES($1, $2, $3) RETURNING id`,
		orderNumber, OrderStatusWithdrawn, userID).Scan(&orderID)
	if err != nil {
		return err
	}

	_, err = db.DB.ExecContext(ctx,
		`INSERT INTO bonus_transactions(amount, type, user_id, order_id) VALUES($1, $2, $3, $4)`,
		amount, WithdrawalType, userID, orderID)
	if err != nil {
		return err
	}

	_, err = db.DB.ExecContext(ctx,
		`UPDATE users SET bonuses=$1 WHERE id=$2`,
		userBonuses-amount,
		userID)
	return err
}
