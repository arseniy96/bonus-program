package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

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
var ErrNowRows = errors.New(`missing data`)

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

func (db *Database) UpdateUserToken(ctx context.Context, login, token string, tokenExp time.Time) error {
	_, err := db.DB.ExecContext(ctx,
		`UPDATE users SET token=$1, token_exp_at=$2 WHERE login=$3`,
		token,
		tokenExp,
		login)
	return err
}

func (db *Database) FindUserByLogin(ctx context.Context, login string) (*User, error) {
	var u User
	err := db.DB.QueryRowContext(ctx,
		`SELECT id, login, password, token, bonuses, token_exp_at FROM users WHERE login=$1 LIMIT(1)`,
		login).Scan(&u.ID, &u.Login, &u.Password, &u.Token, &u.Bonuses, &u.TokenExpAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNowRows
		}
		return nil, err
	}

	return &u, nil
}

func (db *Database) FindUserByToken(ctx context.Context, token string) (*User, error) {
	var u User
	err := db.DB.QueryRowContext(ctx,
		`SELECT id, login, password, token, bonuses, token_exp_at FROM users WHERE token=$1 LIMIT(1)`,
		token).Scan(&u.ID, &u.Login, &u.Password, &u.Token, &u.Bonuses, &u.TokenExpAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNowRows
		}
		return nil, err
	}

	return &u, nil
}

func (db *Database) FindOrdersByUserID2(ctx context.Context, userID int) ([]Order, error) {
	rows, err := db.DB.QueryContext(ctx,
		`SELECT o.id, o.order_number, o.status, o.user_id, o.created_at, bt.amount FROM orders o LEFT JOIN bonus_transactions bt ON o.id = bt.order_id WHERE o.user_id=$1 AND o.status!=$2`,
		userID,
		OrderStatusWithdrawn) // FIXME: таким запросом не получится достать заказы, у которых нет записи в bonus_transactions
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

func (db *Database) FindOrdersByUserID(ctx context.Context, userID int) ([]Order, error) {
	orderRows, err := db.DB.QueryContext(ctx,
		`SELECT id, order_number, status, user_id, created_at FROM orders WHERE user_id=$1 AND status!=$2`,
		userID,
		OrderStatusWithdrawn)
	if err != nil {
		return nil, err
	}
	defer orderRows.Close()

	var orders []Order
	for orderRows.Next() {
		var order Order
		err = orderRows.Scan(&order.ID, &order.OrderNumber, &order.Status, &order.UserID, &order.CreatedAt)
		if err != nil {
			return nil, err
		}
		// FIXME: надо доставать бонусы из другой таблицы вместе с заказами.
		//  НО если транзакции нет, то падает ошибка, что мы через scan &order.BonusAmount пытаемся записать nil
		//  больше нет других идей, кроме как доставать отдельный и обрабатывать ошибку, если нет записи :facepalm:
		err = db.DB.QueryRowContext(ctx,
			`SELECT amount FROM bonus_transactions WHERE order_id=$1 AND type=$2`,
			order.ID, AccrualType).Scan(&order.BonusAmount)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
		}

		orders = append(orders, order)
	}
	err = orderRows.Err()
	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (db *Database) FindOrderByOrderNumber(ctx context.Context, orderNumber string) (*Order, error) {
	var order Order
	err := db.DB.QueryRowContext(ctx,
		`SELECT id, order_number, status, user_id, created_at FROM orders WHERE order_number=$1 LIMIT 1`,
		orderNumber).Scan(&order.ID, &order.OrderNumber, &order.Status, &order.UserID, &order.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNowRows
		}
		return nil, err
	}

	return &order, nil
}

func (db *Database) CreateOrder(ctx context.Context, userID int, orderNumber, status string) (*Order, error) {
	var order Order
	err := db.DB.QueryRowContext(ctx,
		`INSERT INTO orders(user_id, order_number, status) VALUES($1, $2, $3) RETURNING id, order_number, status, user_id, created_at`,
		userID, orderNumber, status).Scan(&order.ID, &order.OrderNumber, &order.Status, &order.UserID, &order.CreatedAt)

	return &order, err
}

func (db *Database) UpdateOrderStatus(ctx context.Context, order *Order, status string, bonus int) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`UPDATE orders SET status=$1 WHERE id=$2`,
		status, order.ID)
	if err != nil {
		return err
	}

	// если бонусы не начислены, не надо ничего обновлять
	if bonus != 0 {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO bonus_transactions(amount, type, user_id, order_id) VALUES($1, $2, $3, $4)`,
			bonus, AccrualType, order.UserID, order.ID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx,
			`UPDATE users SET bonuses=bonuses+$1 WHERE id=$2`,
			bonus,
			order.UserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
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
	var total sql.NullInt64
	err := db.DB.QueryRowContext(ctx,
		`SELECT SUM(amount) AS total FROM bonus_transactions WHERE user_id=$1 AND type=$2`,
		userID, WithdrawalType).Scan(&total)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return int(total.Int64), err
}

func (db *Database) SaveWithdrawBonuses(ctx context.Context, userID int, orderNumber string, amount int) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var orderID int
	err = tx.QueryRowContext(ctx,
		`INSERT INTO orders(order_number, status, user_id) VALUES($1, $2, $3) RETURNING id`,
		orderNumber, OrderStatusWithdrawn, userID).Scan(&orderID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO bonus_transactions(amount, type, user_id, order_id) VALUES($1, $2, $3, $4)`,
		amount, WithdrawalType, userID, orderID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE users SET bonuses=bonuses-$1 WHERE id=$2`,
		amount,
		userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
