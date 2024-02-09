BEGIN TRANSACTION;

    CREATE TABLE IF NOT EXISTS orders(
        id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        order_number VARCHAR,
        user_id INT NOT NULL,
        status VARCHAR,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id)
    );

    CREATE UNIQUE INDEX IF NOT EXISTS order_number_idx on orders(order_number);

COMMIT;