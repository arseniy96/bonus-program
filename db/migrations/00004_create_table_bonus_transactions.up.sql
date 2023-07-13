BEGIN TRANSACTION;

    CREATE TABLE IF NOT EXISTS bonus_transactions(
        amount INT,
        type VARCHAR,
        user_id INT,
        order_id INT,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id),
        CONSTRAINT fk_order FOREIGN KEY(order_id) REFERENCES orders(id)
    );

COMMIT;