BEGIN TRANSACTION;

    DROP TABLE IF EXISTS orders;

    DROP INDEX IF EXISTS order_number_idx;

COMMIT;