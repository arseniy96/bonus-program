BEGIN TRANSACTION;

    CREATE TABLE IF NOT EXISTS users(
        id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
        login VARCHAR,
        password VARCHAR,
        token VARCHAR
    );

    CREATE UNIQUE INDEX IF NOT EXISTS login_idx on users(login);

COMMIT;